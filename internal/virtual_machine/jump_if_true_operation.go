package virtual_machine

import ()

// jumpIfTrueOperation will set the program counter to the second parameter if the value provided is true. It
// will also set the last-address register (a typically write-protected register) to what would have been
// the next instruction if the jump was not taken
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

	s.nextProgramCounter = s.getProgramCounter() + 3
	if firstParam.Value != 0 {
		s.registerFile[RegisterLastAddress] = s.nextProgramCounter
		s.nextProgramCounter = secondParam.Value
	}

	return nil
}

func (s *jumpIfTrueOperation) GetNextProgramCounter() int { return s.nextProgramCounter }

func (s *jumpIfTrueOperation) Halt() bool { return false }
