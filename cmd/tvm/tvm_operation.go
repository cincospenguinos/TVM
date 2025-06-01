package tvm

import ()

// TVMOperation represents any valid TVM operation
type TVMOperation interface {
	// Execute performs the underlying operation. The operation may write back to memory
	Execute() error

	// GetNextProgramCounter returns the program counter of the next operation after this one
	GetNextProgramCounter() int

	// Halt returns true if this operation requires halting
	Halt() bool
}

// haltOperation does nothing and informs the machine that it is time to halt
type haltOperation struct {
	*TsvetokVirtualMachine
}

func newHaltOperation(t *TsvetokVirtualMachine) haltOperation {
	return haltOperation{t}
}

func (_ haltOperation) Execute() error {
	return nil
}

func (h haltOperation) GetNextProgramCounter() int { return h.getProgramCounter() }

func (_ haltOperation) Halt() bool { return true }

// multiplyOperation multiplies two nubmers together
type multiplyOperation struct {
	*TsvetokVirtualMachine
}

func newMultiplyOperation(t *TsvetokVirtualMachine) multiplyOperation {
	return multiplyOperation{t}
}

func (m multiplyOperation) Execute() error {
	memory := m.getMemory()
	programCounter := m.getProgramCounter()

	leftParam, err := m.getFirstParam()
	if err != nil {
		return err
	}

	rightParam, err := m.getSecondParam()
	if err != nil {
		return err
	}

	outAddr := memory[programCounter+3]

	memory[outAddr] = leftParam.Value * rightParam

	return nil
}

func (m multiplyOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 4 }

func (_ multiplyOperation) Halt() bool { return false }

// inputOperation multiplies two nubmers together
type inputOperation struct {
	*TsvetokVirtualMachine
}

func newInputOperation(t *TsvetokVirtualMachine) inputOperation {
	return inputOperation{t}
}

func (m inputOperation) Execute() error {
	memory := m.getMemory()
	number := m.ReceiveInput()
	address := memory[m.getProgramCounter()+1]
	memory[address] = number

	return nil
}

func (m inputOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 2 }

func (_ inputOperation) Halt() bool { return false }

// outputOperation multiplies two nubmers together
type outputOperation struct {
	*TsvetokVirtualMachine
}

func newOutputOperation(t *TsvetokVirtualMachine) outputOperation {
	return outputOperation{t}
}

func (m outputOperation) Execute() error {
	param, err := m.getFirstParam()
	if err != nil {
		return err
	}

	m.EmitOutput(param.Value)

	return nil
}

func (m outputOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 2 }

func (_ outputOperation) Halt() bool { return false }

type setIfEqualOperation struct {
	*TsvetokVirtualMachine
}

func newSetIfEqualOperation(t *TsvetokVirtualMachine) setIfEqualOperation {
	return setIfEqualOperation{t}
}

func (s setIfEqualOperation) Execute() error {
	memory := s.getMemory()

	leftParam, err := s.getFirstParam()
	if err != nil {
		return err
	}

	rightParam, err := s.getSecondParam()
	if err != nil {
		return err
	}

	outputAddr := memory[s.getProgramCounter()+3]

	if leftParam.Value == rightParam {
		memory[outputAddr] = 1
	} else {
		memory[outputAddr] = 0
	}

	return nil
}

func (s setIfEqualOperation) GetNextProgramCounter() int { return s.getProgramCounter() + 4 }

func (s setIfEqualOperation) Halt() bool { return false }

type jumpIfTrueOperation struct {
	*TsvetokVirtualMachine
	nextProgramCounter int
}

func newJumpIfTrueOperation(t *TsvetokVirtualMachine) *jumpIfTrueOperation {
	return &jumpIfTrueOperation{t, -1}
}

func (s *jumpIfTrueOperation) Execute() error {
	firstParam, err := s.getFirstParam()
	if err != nil {
		return err
	}

	secondParam, err := s.getSecondParam()
	if err != nil {
		return err
	}

	if firstParam.Value != 0 {
		s.nextProgramCounter = secondParam
		return nil
	}

	s.nextProgramCounter = s.getProgramCounter() + 3

	return nil
}

func (s *jumpIfTrueOperation) GetNextProgramCounter() int { return s.nextProgramCounter }

func (s *jumpIfTrueOperation) Halt() bool { return false }
