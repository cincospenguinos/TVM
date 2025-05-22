package tvm

import (
	"fmt"
)

type TsvetokVirtualMachine struct {
	memory []int
	programCounter int
}

// TsvetokVirtualMachineOperation represents any valid TVM operation
type TsvetokVirtualMachineOperation interface {

	// Execute performs the underlying operation, writing to memory
	Execute() error
}

type addOperation struct {
	machine *TsvetokVirtualMachine
}

func newAddOperation(t *TsvetokVirtualMachine) addOperation {
	return addOperation{t}
}

func (a addOperation) Execute() error {
	memory := a.machine.getMemory()
	programCounter := a.machine.getProgramCounter()

	leftAddr := memory[programCounter + 1]
	rightAddr := memory[programCounter + 2]
	outAddr := memory[programCounter + 3]

	memory[outAddr] = memory[leftAddr] + memory[rightAddr]

	return nil
}

func NewTsvetokVirtualMachine(program []int) *TsvetokVirtualMachine {
	return &TsvetokVirtualMachine{
		memory: program,
		programCounter: 0,
	}
}

func (t *TsvetokVirtualMachine) Execute() error {
	for {
		if currentOperationStruct := t.getCurrentOperation(); currentOperationStruct != nil {
			err := currentOperationStruct.Execute()
			if err != nil {
				return err
			}

			t.programCounter += 4

			continue
		}

		currentOperation := t.memory[t.programCounter]

		if currentOperation == 1 { // ADD
			leftAddr := t.memory[t.programCounter + 1]
			rightAddr := t.memory[t.programCounter + 2]
			outAddr := t.memory[t.programCounter + 3]

			t.memory[outAddr] = t.memory[leftAddr] + t.memory[rightAddr]
			t.programCounter += 4
		} else if currentOperation == 2 { // MULTIPLY
			leftAddr := t.memory[t.programCounter + 1]
			rightAddr := t.memory[t.programCounter + 2]
			outAddr := t.memory[t.programCounter + 3]

			t.memory[outAddr] = t.memory[leftAddr] * t.memory[rightAddr]
			t.programCounter += 4
		} else if currentOperation == 9 { // HALT
			break
		} else {
			return fmt.Errorf(`no operation matches opcode "%v"`, currentOperation)
		}
	}

	return nil
}

func (t *TsvetokVirtualMachine) getCurrentOperation() TsvetokVirtualMachineOperation {
	opCode := t.memory[t.programCounter]
	if opCode == 1 {
		return newAddOperation(t)
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
