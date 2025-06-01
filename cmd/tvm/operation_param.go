package tvm

import (
	"fmt"
)

const (
	// ParamFormatAddress indicates that the given parameter is to be interpreted in address mode. This means
	// that for a given value '12' in that location, it is to be interpreted as the entry in memory at address
	// '12'
	ParamFormatAddress   = 0

	// ParamFormatImmediate indicates that the given parameter is to be interpreted as an immediate. This means
	// that for the value '7' in the parameter's location, it is to be interpreted simply as the value '7'.
	ParamFormatImmediate = 1
)

// operationParam encapsulates a given parameter, indicating what format type it is, the address in
// the slot requested, and the value found in memory at that address. This struct only describes what
// is currently understood by what exists in memory. Constituent operations will be required to make
// their own decisions further down
type operationParam struct {
	Format  int
	Value   int
}

func newOperationParam(t *TsvetokVirtualMachine, paramFormat, paramAddress int) (operationParam, error) {
	if paramFormat > ParamFormatImmediate || paramFormat < ParamFormatAddress {
		return operationParam{}, fmt.Errorf("unknown parameter format '%v' at address '%v'", paramFormat, paramAddress)
	}

	memory := t.getMemory()
	immediate := memory[paramAddress]
	if paramFormat == ParamFormatImmediate {
		return operationParam{paramFormat, immediate}, nil
	}

	// ParamFormatAddress means we want the value in memory---that is, our immediate value is actually
	// an address
	return operationParam{paramFormat, memory[immediate]}, nil
}
