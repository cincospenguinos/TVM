package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// dataValueKind is the form of a parsed data value token
type dataValueKind int

const (
	// dataValueLiteral is an integer literal: a number with an optional '-' sign and/or an
	// 'i' prefix (12, -1, i12, i-12). Per the TVA agreement, a bare number is shorthand for
	// a prepended 'i' (immediate), so the two spellings are equivalent
	dataValueLiteral dataValueKind = iota

	// dataValueIndexRef is an 'i.<label>' reference: the word emitted is the label's index
	dataValueIndexRef

	// dataValueDeref is a '$.<label>' reference: the word emitted is the word stored at the
	// label's index, copied from wherever that word ends up (which may be emitted after this
	// word, so the copy is deferred until the whole program is assembled)
	dataValueDeref
)

// dataValue is a single parsed data value: either a literal integer or a reference to a label
type dataValue struct {
	kind  dataValueKind
	value int    // the parsed number, for dataValueLiteral
	label string // the referenced label name, for the reference kinds
}

// dataLabel is a parsed data label declaration: a label name (the first chunk of the line)
// followed by the data values that occupy the words beginning at the label's index
type dataLabel struct {
	name   string
	values []dataValue
}

// pendingDataReference is a data word that was emitted as a 0 placeholder and is still
// waiting for the real value of the word at a label's index; resolveDataReferences
// finalizes it once the target word is known
type pendingDataReference struct {
	slot  int    // the position of the placeholder in the assembled word list
	label string // the label whose index points at the word to copy
}

// dataValueNumberPattern matches an integer data value, with an optional 'i' (immediate)
// prefix and/or a '-' sign: 12, -1, i12, i-12
var dataValueNumberPattern = regexp.MustCompile(`^i?-?\d+$`)

// isDataLabel reports whether the line is a data label: a valid label name occupying the first
// chunk, followed by one or more data values. A label that occupies its entire line is a
// regular declaration, not a data label, and a line whose first chunk is not a valid name
// (`.foo.`, `.$foo`, …) is neither
func isDataLabel(line tsvasmLine) bool {
	chunks := strings.Fields(line.assemblyCode)
	return len(chunks) > 1 && labelDeclarationPattern.MatchString(chunks[0])
}

// parseDataLabel parses a data label line into its name and values, returning an error if the
// first chunk is not a valid label name or any value token is not a recognized form. The line
// must have been classified as a data label by isDataLabel, but the name and values are
// re-validated here rather than trusting that classification
func parseDataLabel(line string) (dataLabel, error) {
	chunks := strings.Fields(line)
	if len(chunks) < 2 {
		return dataLabel{}, fmt.Errorf("'%v' is not a data label: a data label is a label name followed by at least one data value", line)
	}

	if !labelDeclarationPattern.MatchString(chunks[0]) {
		return dataLabel{}, fmt.Errorf("invalid label '%v': a label is a '.' followed by one or more of letters, digits, dashes, or underscores", chunks[0])
	}

	name := chunks[0][1:]
	values := make([]dataValue, 0, len(chunks)-1)
	for _, chunk := range chunks[1:] {
		value, err := parseDataValue(chunk)
		if err != nil {
			return dataLabel{}, fmt.Errorf("invalid label '%v': %v", name, err)
		}

		values = append(values, value)
	}

	return dataLabel{name: name, values: values}, nil
}

// parseDataValue parses a single data value token. A token is exactly one of: an integer with
// an optional 'i' prefix and/or '-' sign, an 'i.<label>' reference, or a '$.<label>'
// reference. Anything else (a bare word, a float, a trailing comma, …) is an error. There is
// no comma tolerance here, unlike instruction parameters
func parseDataValue(token string) (dataValue, error) {
	if matches := labelReferencePattern.FindStringSubmatch(token); matches != nil {
		kind := dataValueIndexRef
		if matches[1] == ParamIndicatorMemoryAddress {
			kind = dataValueDeref
		}

		return dataValue{kind: kind, label: token[2:]}, nil
	}

	if dataValueNumberPattern.MatchString(token) {
		value, err := strconv.Atoi(strings.TrimPrefix(token, "i"))
		if err != nil {
			return dataValue{}, fmt.Errorf("'%v' is not a valid data value: %v", token, err) // cannot happen
		}

		return dataValue{kind: dataValueLiteral, value: value}, nil
	}

	return dataValue{}, fmt.Errorf("'%v' is not a valid data value: a data value is an integer (with an optional 'i' prefix and/or '-' sign), an 'i.<label>' reference, or a '$.<label>' reference", token)
}

// appendWords appends the words for the data label to the assembled program: one word per
// value, in order. A literal or 'i.<label>' value resolves to its word immediately; a
// '$.<label>' value emits a 0 placeholder and a pendingDataReference, because the word to
// copy may not have been emitted yet (a forward reference), and resolving it eagerly would
// hide a cycle between data labels. It returns the updated program and the new pending
// references
func (d dataLabel) appendWords(labelIndexes map[string]int, program []int) ([]int, []pendingDataReference, error) {
	pending := make([]pendingDataReference, 0)

	for _, value := range d.values {
		switch value.kind {
		case dataValueLiteral:
			program = append(program, value.value)
		case dataValueIndexRef:
			index, resolved := labelIndexes[value.label]
			if !resolved {
				// Cannot happen once the extract and validate passes have run, but guard
				// against a silent misassembly rather than emitting a bogus word
				return program, nil, fmt.Errorf("reference to '%v' could not be resolved", value.label)
			}

			program = append(program, index)
		case dataValueDeref:
			if _, resolved := labelIndexes[value.label]; !resolved {
				return program, nil, fmt.Errorf("reference to '%v' could not be resolved", value.label)
			}

			program = append(program, 0)
			pending = append(pending, pendingDataReference{slot: len(program) - 1, label: value.label})
		}
	}

	return program, pending, nil
}

// resolveDataReferences finalizes every pending '$.' data reference in the assembled program.
// A pending slot's word is the word at the index of the label it refers to; if that word is
// itself pending (the data labels refer to one another), the reference is followed until a
// finalized word is reached, and the whole chain takes its value from that word. Revisiting a
// slot within a single walk is a cycle, which cannot be resolved, and is an error. Backward
// references (whose target word was emitted before the placeholder) resolve through this same
// code path with no special case, and a program with no pending references is a no-op
func resolveDataReferences(program []int, labelIndexes map[string]int, pending []pendingDataReference) error {
	slotToLabel := make(map[int]string, len(pending))
	for _, ref := range pending {
		slotToLabel[ref.slot] = ref.label
	}

	for _, ref := range pending {
		if _, stillPending := slotToLabel[ref.slot]; !stillPending {
			continue // already finalized as part of a chain walked from an earlier slot
		}

		// The chain of pending slots this reference walks, and the label each one waits on
		chain := []int{ref.slot}
		chainLabels := []string{ref.label}
		chainIndex := map[int]int{ref.slot: 0}

		for {
			chasing := slotToLabel[chain[len(chain)-1]]
			target, resolved := labelIndexes[chasing]
			if !resolved {
				return fmt.Errorf("reference to '%v' could not be resolved", chasing)
			}

			if target >= len(program) {
				return fmt.Errorf("cannot resolve '$.%v' data value: the label points past the end of the program, and there is no word to copy", chasing)
			}

			if start, revisited := chainIndex[target]; revisited {
				cycle := append(append([]string{}, chainLabels[start:]...), chainLabels[start])
				return fmt.Errorf("cycle in '$.' data references: %v", strings.Join(cycle, " -> "))
			}

			if nextLabel, isPending := slotToLabel[target]; isPending {
				chainIndex[target] = len(chain)
				chain = append(chain, target)
				chainLabels = append(chainLabels, nextLabel)
				continue
			}

			for _, slot := range chain {
				program[slot] = program[target]
				delete(slotToLabel, slot)
			}

			break
		}
	}

	return nil
}
