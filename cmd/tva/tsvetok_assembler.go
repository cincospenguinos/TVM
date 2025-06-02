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

// NewAssemblerFromString returns a TsvetokAssembler instance with the provided string as assembly code.
// Note that this does not return any errors or attempt to assemble the underlying assembly code
func NewAssemblerFromString(programStr string) *TsvetokAssembler {
	return &TsvetokAssembler{programStr}
}

type instructionBuilder struct {
	OpCode int
	Params []int
}

func (i *instructionBuilder) addParam(paramStr string, paramIndex int) error {
	numericPattern := regexp.MustCompile(`^\d+$`)
	paramStr = strings.ReplaceAll(paramStr, ",", "")

	if strings.Contains(paramStr, "$") {
		paramStr = strings.ReplaceAll(paramStr, "$", "")
	} else if strings.Contains(paramStr, "i") || numericPattern.MatchString(paramStr) {
		paramStr = strings.ReplaceAll(paramStr, "i", "")

		// TODO: Is there a nicer way we can calculate this?
		if paramIndex == 0 {
			i.OpCode += 100
		} else if paramIndex == 1 {
			i.OpCode += 1000
		} else if paramIndex == 2 {
			i.OpCode += 10000
		} else {
			return fmt.Errorf("cannot have more than three params for any operation")
		}
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

	for lineIndex, line := range newLines {
		builder := &instructionBuilder{}
		chunks := spacesPattern.Split(line, -1)
		operation := chunks[0]
		params := chunks[1:]

		switch operation {
		case "add":
			builder.OpCode = 1
		case "mlt":
			builder.OpCode = 2
		case "in":
			builder.OpCode = 3
		case "out":
			builder.OpCode = 4
		case "seq":
			builder.OpCode = 5
		case "hlt":
			builder.OpCode = 9
		default:
			// TODO: Do we want to just gather and report all of the errors instead of stopping assembly at the first one?
			return []int{}, fmt.Errorf("unknown instruction '%v' on line %v", operation, lineIndex)
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
