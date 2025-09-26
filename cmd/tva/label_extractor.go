package tva

import (
	"fmt"
	"regexp"
	"strings"
)

// labelExtractor is a helper object that extracts labels from a provided line of assembly
// code, keeping track of what original line number it was either declared or used, and is
// to be used to supply correct indexing information at assembly time. Note that there are
// two usages of labels: declarations (e.g. `.my-variable 0`) and accessions (e.g. 
// `add i.my-variable, 1, $.my-variable`)
type labelExtractor struct {
	labelsDeclared  map[string]tsvasmLine
	labelAccessions map[string]tsvasmLine
}

func newLabelExtractor() (*labelExtractor) {
	return &labelExtractor{
		labelsDeclared:  make(map[string]tsvasmLine),
		labelAccessions: make(map[string]tsvasmLine),
	}
}

// Extracts labels from a given line of assembly
func (l* labelExtractor) ExtractLabels(line tsvasmLine) error {
	if strings.HasPrefix(line.assemblyCode, ".") {
		labelName := strings.ReplaceAll(line.assemblyCode, ".", "")

		if l.LabelDeclared(labelName) {
			line, _ := l.labelsDeclared[labelName]

			// TODO: Extract to separate error struct?
			return fmt.Errorf("'%v' is already declared on line %v: '%v'", labelName, line.lineNumber, line.assemblyCode)
		}

		l.labelsDeclared[labelName] = line

		return nil
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

// Returns true if the label name provided is declared in the program
func (l* labelExtractor) LabelDeclared(labelName string) bool {
	_, found := l.labelsDeclared[labelName]
	return found
}

// Returns true if the label name provided is used as an accession in the program
func (l* labelExtractor) LabelAccessed(labelName string) bool {
	_, found := l.labelAccessions[labelName]
	return found
}
