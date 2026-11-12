package main

import (
	"fmt"
	"regexp"
	"strings"
)

// labelNamePattern is the shape of a valid label name: a '.' followed by one or more of the
// letters, digits, dashes, and underscores. This is the single source of truth for the name
// shape; the patterns below are its anchorings (to an entire line for a regular declaration,
// to the start of a line for error reporting, or to a parameter for a reference)
var labelNamePattern = `\.[A-Za-z0-9_-]+`

// labelDeclarationPattern matches a valid label declaration line: the name shape anchored to
// the entire line, which a regular label must occupy of its own. It also matches a valid
// label name in any chunk, since a chunk carries no whitespace; isDataLabel uses it that way
var labelDeclarationPattern = regexp.MustCompile(`^` + labelNamePattern + `$`)

// labelPrefixPattern matches the start of a valid label declaration; it is used to tell a label
// that is followed by extraneous content apart from a label whose name is invalid to begin with
var labelPrefixPattern = regexp.MustCompile(`^` + labelNamePattern)

// labelReferencePattern matches a valid reference to a label in a parameter position: a format
// indicator (capture group 1; 'i' for immediate or '$' for address) followed by the name
// shape
var labelReferencePattern = regexp.MustCompile(`^([i$])` + labelNamePattern + `$`)

// labelReference is a reference to a label: the format indicator ('i' for immediate, '$' for
// address) and the name of the label referred to
type labelReference struct {
	indicator string
	name      string
}

// labelExtractor is a helper object that extracts labels from a provided line of assembly
// code, keeping track of what original line number it was either declared or used, and is
// to be used to supply correct indexing information at assembly time. Note that there are
// three usages of labels: regular declarations (e.g. `.main`, which occupy a line of their
// own), data declarations (e.g. `.data 12`, a label name followed by data values), and
// accessions (e.g. `add i.main, 1, $.main`)
type labelExtractor struct {
	labelsDeclared      map[string]tsvasmLine
	labelAccessions     map[string]tsvasmLine
	labelAccessionOrder []string
}

func newLabelExtractor() *labelExtractor {
	return &labelExtractor{
		labelsDeclared:  make(map[string]tsvasmLine),
		labelAccessions: make(map[string]tsvasmLine),
	}
}

// Extracts labels from a given line of assembly, returning an error if the line declares a
// label that is syntactically invalid or already declared, or references a label with invalid
// syntax
func (l *labelExtractor) ExtractLabels(line tsvasmLine) error {
	if strings.HasPrefix(line.assemblyCode, ".") {
		return l.declareLabel(line)
	}

	for _, param := range line.extractChunks()[1:] {
		paramSansComma := strings.ReplaceAll(param, ",", "")

		reference, isReference, err := parseLabelReference(paramSansComma)
		if err != nil {
			return err
		}

		// Only record the first occurrence of a reference; it is what errors will point at
		if isReference && !l.LabelAccessed(reference.name) {
			l.labelAccessions[reference.name] = line
			l.labelAccessionOrder = append(l.labelAccessionOrder, reference.name)
		}
	}

	return nil
}

// declareLabel validates that the line is a label declaration (regular or data) and records
// it, returning an error if the label is invalid or was already declared
func (l *labelExtractor) declareLabel(line tsvasmLine) error {
	if isDataLabel(line) {
		return l.declareDataLabel(line)
	}

	name, err := parseLabelDeclaration(line.assemblyCode)
	if err != nil {
		return err
	}

	return l.recordDeclaration(name, line)
}

// declareDataLabel records a data label (a label name followed by data values): the name is
// declared exactly as a regular label, and every 'i.'/'$.' reference in the values is
// recorded for ValidateAccessions
func (l *labelExtractor) declareDataLabel(line tsvasmLine) error {
	dataLabel, err := parseDataLabel(line.assemblyCode)
	if err != nil {
		return err
	}

	if err := l.recordDeclaration(dataLabel.name, line); err != nil {
		return err
	}

	for _, value := range dataLabel.values {
		if value.kind == dataValueLiteral {
			continue
		}

		// Only record the first occurrence of a reference; it is what errors will point at
		if !l.LabelAccessed(value.label) {
			l.labelAccessions[value.label] = line
			l.labelAccessionOrder = append(l.labelAccessionOrder, value.label)
		}
	}

	return nil
}

// recordDeclaration records the name as declared on the given line, returning an error if the
// name was already declared
func (l *labelExtractor) recordDeclaration(name string, line tsvasmLine) error {
	if firstDeclaration, declared := l.labelsDeclared[name]; declared {
		return fmt.Errorf("'%v' is already declared on line %v: '%v'", name, firstDeclaration.lineNumber, firstDeclaration.assemblyCode)
	}

	l.labelsDeclared[name] = line

	return nil
}

// parseLabelDeclaration validates that the given (already trimmed) line is a regular label
// declaration and returns the declared name. A line whose first chunk is a valid name
// followed by data values is a data label (see parseDataLabel) and never reaches here; a
// line with a valid name followed by other junk (`.foo.`) is a label that was not left
// alone on its line; and a line whose name is invalid to begin with is simply an invalid
// label
func parseLabelDeclaration(line string) (string, error) {
	if labelDeclarationPattern.MatchString(line) {
		return line[1:], nil
	}

	if prefix := labelPrefixPattern.FindString(line); prefix != "" {
		return "", fmt.Errorf("invalid label '%v': a label declaration must be alone on its line", prefix)
	}

	name := line[1:]
	if name == "" {
		return "", fmt.Errorf("invalid label: a label is a '.' followed by at least one of letters, digits, dashes, or underscores")
	}

	return "", fmt.Errorf("invalid label '%v': a label is a '.' followed by one or more of letters, digits, dashes, or underscores", name)
}

// parseLabelReference parses the given (already comma-stripped) parameter as a label reference.
// ok is true when the parameter is a syntactically valid reference (i.<label> or $.<label>);
// an error is returned when the parameter attempts to be a reference but is malformed; and
// neither holds when the parameter is not a reference at all
func parseLabelReference(param string) (reference labelReference, ok bool, err error) {
	if matches := labelReferencePattern.FindStringSubmatch(param); matches != nil {
		return labelReference{indicator: matches[1], name: matches[0][2:]}, true, nil
	}

	if strings.HasPrefix(param, ParamIndicatorImmediate+".") || strings.HasPrefix(param, ParamIndicatorMemoryAddress+".") {
		return labelReference{}, false, fmt.Errorf("invalid label reference '%v': a reference is 'i.<label>' or '$.<label>' where <label> is letters, digits, dashes, or underscores", param)
	}

	return labelReference{}, false, nil
}

// Returns true if the label name provided is declared in the program
func (l *labelExtractor) LabelDeclared(labelName string) bool {
	_, found := l.labelsDeclared[labelName]
	return found
}

// Returns true if the label name provided is used as an accession in the program
func (l *labelExtractor) LabelAccessed(labelName string) bool {
	_, found := l.labelAccessions[labelName]
	return found
}

// ValidateAccessions ensures that every label that was referenced was also declared, returning
// an error (naming the label and the line of its first reference) for the first undeclared
// reference found
func (l *labelExtractor) ValidateAccessions() error {
	for _, labelName := range l.labelAccessionOrder {
		if !l.LabelDeclared(labelName) {
			reference := l.labelAccessions[labelName]
			return fmt.Errorf("'%v' is referenced on line %v but never declared", labelName, reference.lineNumber)
		}
	}

	return nil
}
