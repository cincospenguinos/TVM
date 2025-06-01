package tva

import (
	"strings"
)

// TsvetokAssembler assembles a given TVA program and converts it into a TVM executable
type TsvetokAssembler struct {
	originalAssembly string
}

func NewAssemblerFromString(programStr string) *TsvetokAssembler {
	return &TsvetokAssembler{programStr}
}

func (a *TsvetokAssembler) Assemble() ([]int, error) {
	assembledProgram := make([]int, 0)
	newLines := strings.Split(a.originalAssembly, "\n")

	for _, line := range newLines {
		if line == "hlt" {
			assembledProgram = append(assembledProgram, 9)
		}
	}

	return assembledProgram, nil
}
