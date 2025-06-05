package tva

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tvm "tvm/internal/virtual_machine"
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

// instructionBuilder represents a single intcode operation. Use toIntcode() to expand it out to its
// proper instruction values
type instructionBuilder struct {
	OpCode int
	Params []int
}


// setOperation() converts the operation provided to its proper opcode, or returns an error if the operation
// does not exist
func (i *instructionBuilder) setOperation(operation string) error {
	switch operation {
	case "add":
		i.OpCode = 1
	case "mlt":
		i.OpCode = 2
	case "in":
		i.OpCode = 3
	case "out":
		i.OpCode = 4
	case "seq":
		i.OpCode = 5
	case "jit":
		i.OpCode = 6
	case "hlt":
		i.OpCode = 9
	default:
		fmt.Errorf("unknown instruction '%v'", operation)
	}

	return nil
}

// TODO: Make these constant values in tvm package
var registerValueMap = map[string]int {
	"r0": 0,
	"r1": 1,
	"r2": 2,
	"r3": 3,
	"r4": 4,
	"t0": 5,
	"t1": 6,
	"t2": 7,
	"t3": 8,
	"t4": 9,
	"t5": 10,
	"t6": 11,
	"t7": 12,
	"t8": 13,
	"la": tvm.RegisterLastAddress,
}

// addParam() adds a parameter to this instruction builder given the string value and where
// in the instruction it is found. Returns an error if the parameter is malformed
func (i *instructionBuilder) addParam(paramStr string, paramIndex int) error {
	numericPattern := regexp.MustCompile(`^\d+$`)
	paramStr = strings.ReplaceAll(paramStr, ",", "")

	var paramFormat tvm.ParamFormat
	if strings.Contains(paramStr, "$") {
		paramStr = strings.ReplaceAll(paramStr, "$", "")
		paramFormat = tvm.ParamFormatAddress
	} else if strings.Contains(paramStr, "i") || numericPattern.MatchString(paramStr) {
		paramStr = strings.ReplaceAll(paramStr, "i", "")
		paramFormat = tvm.ParamFormatImmediate
	} else if strings.Contains(paramStr, "r") || strings.Contains(paramStr, "t") || paramStr == "la" {
		registerValue, registerExists := registerValueMap[paramStr]
		if !registerExists {
			return fmt.Errorf("invalid register param '%v'", paramStr)
		}

		paramStr = fmt.Sprintf("%v", registerValue)
		paramFormat = tvm.ParamFormatRegister
	} else {
		return fmt.Errorf("unknown parameter format '%v'", paramStr)
	}

	err := i.updateOpcodeForParam(paramFormat, paramIndex)
	if err != nil {
		return err
	}

	paramVal, err := strconv.Atoi(paramStr)
	if err != nil {
		return err
	}

	// TODO: Do we need to respect the index? Like do an insert?
	i.Params = append(i.Params, paramVal)

	return nil
}

func (i *instructionBuilder) updateOpcodeForParam(paramFormat tvm.ParamFormat, index int) error {
	if index > 2 {
		return fmt.Errorf("cannot have more than three params for any operation")
	}

	multiplier := 100
	for index > 0 {
		multiplier *= 10
		index -= 1
	}

	i.OpCode += int(paramFormat) * multiplier
	return nil
}

// toIntcode() returns the sequence of integers that matches the inputs it received
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
	for _, line := range a.generateLinesFromOriginalAssembly() {
		builder := &instructionBuilder{}
		chunks := spacesPattern.Split(line.assemblyCode, -1)

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
