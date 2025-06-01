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

func (i *instructionBuilder) addParam(paramStr string, paramIndex int) error {
	paramStr = strings.ReplaceAll(paramStr, ",", "")
	if strings.Contains(paramStr, "$") {
		paramStr = strings.ReplaceAll(paramStr, "$", "")
	}

	paramVal, err := strconv.Atoi(paramStr)
	if err != nil {
		return err
	}

	// TODO: Do we need to respect the index? Like do an insert?
	i.Params = append(i.Params, paramVal)

	return nil
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

		for index, paramStr := range params {
			err := builder.addParam(paramStr, index)
			if err != nil {
				return []int{}, err
			}
		}

		assembledProgram = append(assembledProgram, builder.toIntcode()...)
	}

	return assembledProgram, nil
}
