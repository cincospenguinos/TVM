package tvm

import ()

// TVMOperation represents any valid TVM operation
type TVMOperation interface {
	// Execute performs the underlying operation, writing to memory
	Execute() error

	// Returns true if this operation requires halting
	Halt() bool
}

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
