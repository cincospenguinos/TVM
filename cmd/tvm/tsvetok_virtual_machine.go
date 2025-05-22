package tvm

import (
	"fmt"
)

type TsvetokVirtualMachine struct {
	memory []int
	programCounter int
}

type addOperation struct {
	machine *TsvetokVirtualMachine
}

func NewTsvetokVirtualMachine(program []int) *TsvetokVirtualMachine {
	return &TsvetokVirtualMachine{
		memory: program,
		programCounter: 0,
	}
}

func (t *TsvetokVirtualMachine) Execute() error {
	for {
		currentOperationStruct := t.getCurrentOperation()
		if currentOperationStruct == nil {
			return fmt.Errorf(`no operation found for opcode "%v"`, t.memory[t.programCounter])
		}

		err := currentOperationStruct.Execute()
		if err != nil {
			return err
		}

		if currentOperationStruct.Halt() {
			break
		}

		t.programCounter += 4
	}

	return nil
}

func (t *TsvetokVirtualMachine) getCurrentOperation() TVMOperation {
	opCode := t.memory[t.programCounter]
	if opCode == 1 {
		return newAddOperation(t)
	}

	if opCode == 2 {
		return newMultiplyOperation(t)
	}

	if opCode == 9 {
		return newHaltOperation(t)
	}

	return nil
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
