package tvm

import (
	"fmt"
)

// InputInterface represents any external service or program that can be called upon for
// an integer
type InputInterface interface {
	// ReceiveInput acquires and returns an integer from an external source
	ReceiveInput() int
}

// OutputInterface represents any external service or program that an integer can be
// emitted
type OutputInterface interface {
	// EmitOutput emits the integer provided to a given target
	EmitOutput(int)
}

// TsvetokVirtualMachine is an implementation of the Tsvetok Virtual Machine Intcode machine (or TVM.)
type TsvetokVirtualMachine struct {
	memory         []int
	programCounter int
	InputInterface
	OutputInterface
}

func NewTsvetokVirtualMachine(program []int) *TsvetokVirtualMachine {
	return &TsvetokVirtualMachine{
		memory:         program,
		programCounter: 0,
	}
}

func (t *TsvetokVirtualMachine) Execute() error {
	for {
		currentOperation := t.getCurrentOperation()
		if currentOperation == nil {
			return fmt.Errorf(`no operation found for opcode "%v"`, t.memory[t.programCounter])
		}

		err := currentOperation.Execute()
		if err != nil {
			return err
		}

		if currentOperation.Halt() {
			break
		}

		t.programCounter = currentOperation.GetNextProgramCounter()
	}

	return nil
}

func (t *TsvetokVirtualMachine) getCurrentOperation() TVMOperation {
	rawOpcode := t.memory[t.programCounter]
	opCode := rawOpcode % 100

	switch opCode {
	case 1:
		return newAddOperation(t)
	case 2:
		return newMultiplyOperation(t)
	case 3:
		return newInputOperation(t)
	case 4:
		return newOutputOperation(t)
	case 5:
		return newSetIfEqualOperation(t)
	case 6:
		return newJumpIfTrueOperation(t)
	case 9:
		return newHaltOperation(t)
	default:
		return nil
	}
}

// getMemory returns the TVM's underlying memory. Writing to this slice is persisted across
// the lifetime of the struct.
func (t *TsvetokVirtualMachine) getMemory() []int {
	// TODO: No more direct memory management. We need to guard against attempting to read at invalid places
	return t.memory
}

// operationParam encapsulates a given parameter, indicating what format type it is, the address in
// the slot requested, and the value found in memory at that address. This struct only describes what
// is currently understood by what exists in memory. Constituent operations will be required to make
// their own decisions further down
type operationParam struct {
	Format  int
	Value   int
}

const (
	ParamFormatAddress   = 0
	ParamFormatImmediate = 1
)

func (t *TsvetokVirtualMachine) getFirstParam() (operationParam, error) {
	memory := t.getMemory()
	programCounter := t.getProgramCounter()

	rawOpcode := t.memory[t.programCounter]
	paramFormat := (rawOpcode / 100) % 10

	if paramFormat > ParamFormatImmediate {
		return operationParam{}, fmt.Errorf("unknown first parameter format '%v'", paramFormat)
	}

	immediate := memory[programCounter+1]
	if paramFormat == ParamFormatImmediate {
		return operationParam{paramFormat, immediate}, nil
	}

	// ParamFormatAddress means we want the value in memory---that is, our immediate value is actually
	// an address
	return operationParam{paramFormat, memory[immediate]}, nil
}

func (t *TsvetokVirtualMachine) getSecondParam() (int, error) {
	memory := t.getMemory()
	programCounter := t.getProgramCounter()

	rawOpcode := t.memory[t.programCounter]
	paramFormat := (rawOpcode / 1000) % 10
	if paramFormat == ParamFormatAddress {
		return memory[memory[programCounter+2]], nil
	}

	if paramFormat == ParamFormatImmediate {
		return memory[programCounter+2], nil
	}

	return 0, fmt.Errorf("unknown second parameter format '%v'", paramFormat)
}

func (t *TsvetokVirtualMachine) getThirdParam() (int, error) {
	memory := t.getMemory()
	programCounter := t.getProgramCounter()

	rawOpcode := t.memory[t.programCounter]
	paramFormat := rawOpcode / 10000
	if paramFormat == ParamFormatAddress {
		return memory[memory[programCounter+3]], nil
	}

	if paramFormat == ParamFormatImmediate {
		return memory[programCounter+3], nil
	}

	return 0, fmt.Errorf("unknown third parameter format '%v'", paramFormat)
}

func (t *TsvetokVirtualMachine) getProgramCounter() int {
	return t.programCounter
}

// CopyMemory returns a copy of the TVM's current memory state. If an internal function
// wishes to write to memory, use getMemory() instead.
func (t *TsvetokVirtualMachine) CopyMemory() []int {
	copiedMemory := make([]int, len(t.memory))
	copy(copiedMemory, t.memory)

	return copiedMemory
}

func (t *TsvetokVirtualMachine) SetInputInterface(i InputInterface) {
	t.InputInterface = i
}

func (t *TsvetokVirtualMachine) SetOutputInterface(o OutputInterface) {
	t.OutputInterface = o
}
