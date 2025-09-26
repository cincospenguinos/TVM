package tva

import (
	"errors"
	"fmt"
	"regexp"
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
	assembledProgram := make([]int, 0)
	for _, line := range a.generateLinesFromOriginalAssembly() {
		builder := &instructionBuilder{}
		chunks := line.extractChunks()

		// TODO: Do we want to just gather and report all of the errors instead of stopping assembly at the first one?
		err := builder.setOperation(chunks[0])
		if err != nil {
			return []int{}, errors.Join(err, fmt.Errorf("error on line '%v'", line.lineNumber))
		}

		for index, paramStr := range chunks[1:] {
			err := builder.addParam(paramStr, index)
			if err != nil {
				return []int{}, errors.Join(err, fmt.Errorf("error on line '%v'", line.lineNumber))
			}
		}

		assembledProgram = append(assembledProgram, builder.toIntcode()...)
	}

	return assembledProgram, nil
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
