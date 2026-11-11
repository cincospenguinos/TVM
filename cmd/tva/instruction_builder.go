package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tvm "tvm/internal/virtual_machine"
)

const (
	ParamIndicatorReservedRegister  = "r"
	ParamIndicatorTemporaryRegister = "t"
	ParamIndicatorMemoryAddress     = "$"
	ParamIndicatorImmediate         = "i"
	ParamLastAddressRegister        = "la"
)

var registerValueMap = map[string]int{
	"r0": tvm.RegisterReserved0,
	"r1": tvm.RegisterReserved1,
	"r2": tvm.RegisterReserved2,
	"r3": tvm.RegisterReserved3,
	"r4": tvm.RegisterReserved4,
	"t0": tvm.RegisterTemporary0,
	"t1": tvm.RegisterTemporary1,
	"t2": tvm.RegisterTemporary2,
	"t3": tvm.RegisterTemporary3,
	"t4": tvm.RegisterTemporary4,
	"t5": tvm.RegisterTemporary5,
	"t6": tvm.RegisterTemporary6,
	"t7": tvm.RegisterTemporary7,
	"la": tvm.RegisterLastAddress,
}

// instructionBuilder represents a single intcode operation. Use toIntcode() to expand it out to its
// proper instruction values
type instructionBuilder struct {
	OpCode int
	Params []int

	// paramSlotOffset is the number of leading parameter slots the operation pre-populates on
	// its own (the jmp pseudo-operation, which pre-populates jit's condition slot); the
	// user-visible parameters are placed starting at this offset
	paramSlotOffset int

	// paramMinCount is the minimum number of user-visible parameters the operation requires;
	// zero means the operation is fine without any
	paramMinCount int

	// paramCountLimit is the maximum number of user-visible parameters the operation accepts;
	// zero means there is no limit beyond the machine's three slots
	paramCountLimit int
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
	case "jmp":
		// jmp is a pseudo-instruction that expands to jit with a constant-true condition,
		// so the jump is always taken. The user's single target lands in jit's slot 1
		i.OpCode = 6
		i.Params = append(i.Params, 1)
		i.paramSlotOffset = 1
		i.paramMinCount = 1
		i.paramCountLimit = 1
		if err := i.updateOpcodeForParam(tvm.ParamFormatImmediate, 0); err != nil {
			return err
		}
	case "hlt":
		i.OpCode = 9
	default:
		return fmt.Errorf("unknown instruction '%v'", operation)
	}

	return nil
}

// addParam() adds a parameter to this instruction builder given the string value and where in the instruction it is found.
// Returns an error if the parameter is malformed or beyond the operation's parameter count
func (i *instructionBuilder) addParam(paramStr string, paramIndex int) error {
	if paramIndex < 0 || paramIndex+i.paramSlotOffset >= 3 {
		return fmt.Errorf("'%v' is not a valid parameter index (no less than 0; no operation supports more than three)", paramIndex)
	}

	if i.paramCountLimit > 0 && paramIndex >= i.paramCountLimit {
		return fmt.Errorf("'%v' is an extra parameter (this operation accepts at most %v)", paramStr, i.paramCountLimit)
	}

	numericPattern := regexp.MustCompile(`^\d+$`)
	paramStr = strings.ReplaceAll(paramStr, ",", "")

	paramSlot := paramIndex + i.paramSlotOffset

	var paramFormat tvm.ParamFormat
	if strings.Contains(paramStr, ParamIndicatorMemoryAddress) {
		paramStr = strings.ReplaceAll(paramStr, ParamIndicatorMemoryAddress, "")
		paramFormat = tvm.ParamFormatAddress
	} else if strings.Contains(paramStr, ParamIndicatorImmediate) || numericPattern.MatchString(paramStr) {
		paramStr = strings.ReplaceAll(paramStr, ParamIndicatorImmediate, "")
		paramFormat = tvm.ParamFormatImmediate
	} else if strings.Contains(paramStr, ParamIndicatorReservedRegister) || strings.Contains(paramStr, ParamIndicatorTemporaryRegister) || paramStr == ParamLastAddressRegister {
		registerValue, registerExists := registerValueMap[paramStr]
		if !registerExists {
			return fmt.Errorf("invalid register param '%v'", paramStr)
		}

		paramStr = fmt.Sprintf("%v", registerValue)
		paramFormat = tvm.ParamFormatRegister
	} else {
		return fmt.Errorf("unknown parameter format '%v'", paramStr)
	}

	err := i.updateOpcodeForParam(paramFormat, paramSlot)
	if err != nil {
		return err
	}

	paramVal, err := strconv.Atoi(paramStr)
	if err != nil {
		return err
	}

	for len(i.Params) <= paramSlot {
		i.Params = append(i.Params, 0)
	}

	copy(i.Params[paramSlot+1:], i.Params[paramSlot:])
	i.Params[paramSlot] = paramVal

	return nil
}

// finish() validates the number of user-visible parameters the instruction was given once all
// of them have been added, returning an error if the count is below the operation's minimum
func (i *instructionBuilder) finish(userParamCount int) error {
	if i.paramMinCount > 0 && userParamCount < i.paramMinCount {
		return fmt.Errorf("this operation requires at least %v parameter(s), but %v were given", i.paramMinCount, userParamCount)
	}

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
