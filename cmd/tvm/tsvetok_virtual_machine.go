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

	// Returns true if this operation requires halting
	Halt() bool
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

func (a addOperation) Halt() bool { return false }

type multiplyOperation struct {
	machine *TsvetokVirtualMachine
}

func newMultiplyOperation(t *TsvetokVirtualMachine) multiplyOperation {
	return multiplyOperation{t}
}

func (m multiplyOperation) Execute() error {
	memory := m.machine.getMemory()
	programCounter := m.machine.getProgramCounter()

	leftAddr := memory[programCounter + 1]
	rightAddr := memory[programCounter + 2]
	outAddr := memory[programCounter + 3]

	memory[outAddr] = memory[leftAddr] * memory[rightAddr]

	return nil
}

func (a multiplyOperation) Halt() bool { return false }

type haltOperation struct {
	machine *TsvetokVirtualMachine
}

func newHaltOperation(t *TsvetokVirtualMachine) haltOperation {
	return haltOperation{t}
}

func (m haltOperation) Execute() error {
	return nil
}

func (a haltOperation) Halt() bool { return true }

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

func (t *TsvetokVirtualMachine) getCurrentOperation() TsvetokVirtualMachineOperation {
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
