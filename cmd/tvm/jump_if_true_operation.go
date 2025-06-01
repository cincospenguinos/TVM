package tvm

import ()

// jumpIfTrueOperation will set the program counter to the second parameter if the value provided is true
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
