package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TsvetokAssembler assembles a given TVA program and converts it into a TVM executable
type TsvetokAssembler struct {
	originalAssembly string
}

// NewAssemblerFromString returns a TsvetokAssembler instance with the provided string as assembly code.
// Note that this does not return any errors or attempt to assemble the underlying assembly code
func NewAssemblerFromString(programStr string) *TsvetokAssembler {
	return &TsvetokAssembler{programStr}
}

// Assemble() munges the original assembly program and returns a sequence of integers executable by
// a Tsvetok Virtual Machine, or an error if one was discovered or occurrred during assembly.
func (a *TsvetokAssembler) Assemble() ([]int, error) {
	linesOfCode := a.generateLinesFromOriginalAssembly()

	labelExtractor := newLabelExtractor()
	for _, line := range linesOfCode {
		if err := labelExtractor.ExtractLabels(line); err != nil {
			return []int{}, errors.Join(err, fmt.Errorf("error on line '%v'", line.lineNumber))
		}
	}

	if err := labelExtractor.ValidateAccessions(); err != nil {
		return []int{}, err
	}

	labelIndexes := make(map[string]int)
	currentIndex := 0
	for _, line := range linesOfCode {
		if isLabelDeclaration(line) {
			labelIndexes[line.assemblyCode[1:]] = currentIndex
			continue // a declaration emits no words
		}

		currentIndex += instructionWords(line.extractChunks())
	}

	instructions := make([]*instructionBuilder, 0)
	for _, line := range linesOfCode {
		if isLabelDeclaration(line) {
			continue // a declaration emits no words
		}

		builder := &instructionBuilder{}
		chunks := line.extractChunks()

		// TODO: Do we want to just gather and report all of the errors instead of stopping assembly at the first one?
		err := builder.setOperation(chunks[0])
		if err != nil {
			return []int{}, errors.Join(err, fmt.Errorf("error on line '%v'", line.lineNumber))
		}

		for index, paramStr := range chunks[1:] {
			param := strings.ReplaceAll(paramStr, ",", "")

			if reference, isReference, err := parseLabelReference(param); err != nil {
				return []int{}, errors.Join(err, fmt.Errorf("error on line '%v'", line.lineNumber))
			} else if isReference {
				targetIndex, resolved := labelIndexes[reference.name]
				if !resolved {
					// Cannot happen once the extract and validate passes have run, but guard
					// against a silent misassembly rather than emitting a bogus word
					return []int{}, errors.Join(fmt.Errorf("reference to '%v' could not be resolved", reference.name), fmt.Errorf("error on line '%v'", line.lineNumber))
				}

				// Rewrite the reference to a plain parameter with the resolved index; the
				// instructionBuilder never needs to know about labels
				param = reference.indicator + strconv.Itoa(targetIndex)
			}

			err = builder.addParam(param, index)
			if err != nil {
				return []int{}, errors.Join(err, fmt.Errorf("error on line '%v'", line.lineNumber))
			}
		}

		err = builder.finish(len(chunks) - 1)
		if err != nil {
			return []int{}, errors.Join(err, fmt.Errorf("error on line '%v'", line.lineNumber))
		}

		instructions = append(instructions, builder)
	}

	assembledProgram := make([]int, 0)
	for _, builder := range instructions {
		assembledProgram = append(assembledProgram, builder.toIntcode()...)
	}

	return assembledProgram, nil
}

// isLabelDeclaration reports whether the line is a label declaration (a label occupying its
// entire line). Declaration syntax is already validated by the extract pass, so a line that
// reached the index or build pass and matches this pattern is a known declaration
func isLabelDeclaration(line tsvasmLine) bool {
	return labelDeclarationPattern.MatchString(line.assemblyCode)
}

// instructionWords returns how many words of intcode the given instruction chunks expand to:
// the chunks themselves, plus any parameters that pseudo-operations inject on the user's
// behalf. This is the only place that knows a pseudo-operation's expansion cost, so it must
// be extended whenever a new pseudo-operation (jif, etc.) is added
func instructionWords(chunks []string) int {
	words := len(chunks)

	switch chunks[0] {
	case "jmp":
		words += 1 // jmp expands to jit, which adds a constant-true condition
	}

	return words
}

// tsvasmLine is an intermediary struct that represents the original line of assembly code
// and what line number it originally was in. Keeping the two together allows for better
// debug information and error reporting
type tsvasmLine struct {
	assemblyCode string
	lineNumber   int
}

// extractChunks returns each piece string in a line separated by one or more space characters
func (l tsvasmLine) extractChunks() []string {
	spacesPattern := regexp.MustCompile(`\s+`)
	return spacesPattern.Split(l.assemblyCode, -1)
}

// generateLinesFromOriginalAssembly() is a helper function to convert all lines out to a POJO struct.
// No struct is returned for any lines that consist solely of comments
func (a *TsvetokAssembler) generateLinesFromOriginalAssembly() []tsvasmLine {
	commentsPattern := regexp.MustCompile(`(?m)#.*$`)
	newLines := make([]tsvasmLine, 0)

	noComments := commentsPattern.ReplaceAllLiteralString(a.originalAssembly, "")
	for lineIndex, line := range strings.Split(noComments, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		newLines = append(newLines, tsvasmLine{trimmedLine, lineIndex + 1})
	}

	return newLines
}
