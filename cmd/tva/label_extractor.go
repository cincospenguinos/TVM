package main

import (
	"fmt"
	"regexp"
	"strings"
)

// labelDeclarationPattern matches a valid label declaration line. A label must occupy its entire
// line: a '.' followed by one or more of the letters, digits, dashes, and underscores
var labelDeclarationPattern = regexp.MustCompile(`^\.[A-Za-z0-9_-]+$`)

// labelPrefixPattern matches the start of a valid label declaration; it is used to tell a label
// that is followed by extraneous content apart from a label whose name is invalid to begin with
var labelPrefixPattern = regexp.MustCompile(`^\.[A-Za-z0-9_-]+`)

// labelExtractor is a helper object that extracts labels from a provided line of assembly
// code, keeping track of what original line number it was either declared or used, and is
// to be used to supply correct indexing information at assembly time. Note that there are
// two usages of labels: declarations (e.g. `.main`, which must occupy a line of their own)
// and accessions (e.g. `add i.main, 1, $.main`)
type labelExtractor struct {
	labelsDeclared  map[string]tsvasmLine
	labelAccessions map[string]tsvasmLine
}

func newLabelExtractor() *labelExtractor {
	return &labelExtractor{
		labelsDeclared:  make(map[string]tsvasmLine),
		labelAccessions: make(map[string]tsvasmLine),
	}
}

// Extracts labels from a given line of assembly, returning an error if the line declares a
// label that is syntactically invalid or already declared
func (l *labelExtractor) ExtractLabels(line tsvasmLine) error {
	if strings.HasPrefix(line.assemblyCode, ".") {
		return l.declareLabel(line)
	}

	labelAccessionPattern := regexp.MustCompile(`[\$i]\.`)
	chunks := line.extractChunks()
	params := chunks[1:]

	for _, param := range params {
		if labelAccessionPattern.MatchString(param) {
			paramSansComma := strings.ReplaceAll(param, ",", "")
			labelName := labelAccessionPattern.ReplaceAllLiteralString(paramSansComma, "")
			l.labelAccessions[labelName] = line
		}
	}

	return nil
}

// declareLabel validates that the line is a label declaration and records it, returning an
// error if the label is invalid or was already declared
func (l *labelExtractor) declareLabel(line tsvasmLine) error {
	name, err := parseLabelDeclaration(line.assemblyCode)
	if err != nil {
		return err
	}

	if firstDeclaration, declared := l.labelsDeclared[name]; declared {
		return fmt.Errorf("'%v' is already declared on line %v: '%v'", name, firstDeclaration.lineNumber, firstDeclaration.assemblyCode)
	}

	l.labelsDeclared[name] = line

	return nil
}

// parseLabelDeclaration validates that the given (already trimmed) line is a label declaration
// and returns the declared name. A line that contains a valid label name followed by
// extraneous content is a label that was not left alone on its line; a line whose name is
// invalid to begin with is simply an invalid label
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
