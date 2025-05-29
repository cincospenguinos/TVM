package tvm

import (
	"fmt"
)

type InputInterface interface {
	// ReceiveInput acquires and returns an integer from an external source
	ReceiveInput() int
}

type OutputInterface interface {
	// EmitOutput emits the integer provided to a given target
	EmitOutput(int)
}

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
	opCode := t.memory[t.programCounter]
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
	case 9:
		return newHaltOperation(t)
	default:
		return nil
	}
}

// getMemory returns the TVM's underlying memory. Writing to this slice is persisted across
// the lifetime of the struct.
func (t *TsvetokVirtualMachine) getMemory() []int {
	return t.memory
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
