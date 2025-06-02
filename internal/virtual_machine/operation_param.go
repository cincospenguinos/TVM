package virtual_machine

import (
	"fmt"
)

const (
	// ParamFormatAddress indicates that the given parameter is to be interpreted in address mode. This means
	// that for a given value '12' in that location, it is to be interpreted as the entry in memory at address
	// '12'
	ParamFormatAddress = 0

	// ParamFormatImmediate indicates that the given parameter is to be interpreted as an immediate. This means
	// that for the value '7' in the parameter's location, it is to be interpreted simply as the value '7'.
	ParamFormatImmediate = 1

	// ParamFormatRegister indicates that the given parameter is into the register file. The register file
	// is addressed in the same manner as memory but is limited to its space and cannot be expanded (see
	// TsvetokVirtualMachine docs)
	ParamFormatRegister = 2
)

// operationParam encapsulates a given parameter, indicating what format type it is, the address in
// the slot requested, and the value found in memory at that address. This struct only describes what
// is currently understood by what exists in memory. Constituent operations will be required to make
// their own decisions further down
type operationParam struct {
	// Format is the parameter's given format (i.e. Address, Immediate, etc.)
	Format int

	// Address is the integer found in memory at the parameter's location. For example: if this parameter is the
	// first parameter, then the Address is the value in memory found at the memory entry one past the current
	// program counter
	Address int

	// Value is the integer found in memory at the parameter's Address (see above.) This integer is equal to
	// address if the parameter is in immediate mode
	Value int
}

func newOperationParam(t *TsvetokVirtualMachine, paramFormat, paramAddress int) (operationParam, error) {
	if paramFormat != ParamFormatImmediate && paramFormat != ParamFormatAddress && paramFormat != ParamFormatRegister {
		return operationParam{}, fmt.Errorf("unknown parameter format '%v' at address '%v'", paramFormat, paramAddress)
	}

	immediate, err := t.GetValueInMemory(paramAddress)
	if err != nil {
		return operationParam{}, err
	}

	if paramFormat == ParamFormatImmediate {
		return operationParam{paramFormat, immediate, immediate}, nil
	}

	if paramFormat == ParamFormatRegister {
		registerValue, err := t.GetValueInRegisterFile(immediate)
		if err != nil {
			return operationParam{}, err
		}

		return operationParam{paramFormat, immediate, registerValue}, nil
	}

	value, err := t.GetValueInMemory(immediate)
	if err != nil {
		return operationParam{}, err
	}

	return operationParam{paramFormat, immediate, value}, nil
}
