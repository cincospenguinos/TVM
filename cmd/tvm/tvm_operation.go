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
		s.nextProgramCounter = secondParam.Value
		return nil
	}

	s.nextProgramCounter = s.getProgramCounter() + 3

	return nil
}

func (s *jumpIfTrueOperation) GetNextProgramCounter() int { return s.nextProgramCounter }

func (s *jumpIfTrueOperation) Halt() bool { return false }
