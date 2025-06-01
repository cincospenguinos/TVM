package tva

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TsvetokAssembler assembles a given TVA program and converts it into a TVM executable
type TsvetokAssembler struct {
	originalAssembly string
}

func NewAssemblerFromString(programStr string) *TsvetokAssembler {
	return &TsvetokAssembler{programStr}
}

type instructionBuilder struct {
	OpCode int
	Params []int
}

func (i *instructionBuilder) toIntcode() []int {
	intcode := []int{i.OpCode}

	for _, p := range i.Params {
		intcode = append(intcode, p)
	}

	return intcode
}

func (a *TsvetokAssembler) Assemble() ([]int, error) {
	spacesPattern := regexp.MustCompile(`\s+`)

	assembledProgram := make([]int, 0)
	newLines := strings.Split(a.originalAssembly, "\n")

	for lineNumber, line := range newLines {
		builder := &instructionBuilder{}
		chunks := spacesPattern.Split(line, -1)
		operation := chunks[0]
		params := chunks[1:]

		switch operation {
		case "hlt":
			builder.OpCode = 9
		case "add":
			builder.OpCode = 1
		default:
			return []int{}, fmt.Errorf("unknown instruction '%v' on line %v", line, lineNumber)
		}

		for _, paramStr := range params {
			numStr := strings.ReplaceAll(paramStr, "$", "")
			numStr = strings.ReplaceAll(numStr, ",", "")
			num, err := strconv.Atoi(numStr)
			if err != nil {
				return []int{}, err
			}

			builder.Params = append(builder.Params, num)
		}

		assembledProgram = append(assembledProgram, builder.toIntcode()...)
	}

	return assembledProgram, nil
}
