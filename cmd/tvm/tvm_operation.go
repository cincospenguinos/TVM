package tvm

import ()

// TVMOperation represents any valid TVM operation
type TVMOperation interface {
	// Execute performs the underlying operation, writing to memory
	Execute() error

	// GetNextProgramCounter returns the program counter of the next operation after this one
	GetNextProgramCounter() int

	// Halt returns true if this operation requires halting
	Halt() bool
}

// haltOperation does nothing and informs the machine that it is time to halt
type haltOperation struct {
	machine *TsvetokVirtualMachine
}

func newHaltOperation(t *TsvetokVirtualMachine) haltOperation {
	return haltOperation{t}
}

func (_ haltOperation) Execute() error {
	return nil
}

func (h haltOperation) GetNextProgramCounter() int { return h.machine.getProgramCounter() }

func (_ haltOperation) Halt() bool { return true }

// addOperation adds two numbers together
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

func (a addOperation) GetNextProgramCounter() int { return a.machine.getProgramCounter() + 4 }

func (_ addOperation) Halt() bool { return false }

// multiplyOperation multiplies two nubmers together
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

func (m multiplyOperation) GetNextProgramCounter() int { return m.machine.getProgramCounter() + 4 }

func (_ multiplyOperation) Halt() bool { return false }
