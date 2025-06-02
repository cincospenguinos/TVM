package virtual_machine

import ()

// setIfEqualOperation sets a given memory address to true (1) if the two numbers provided are equal. If they are not
// equal then the provided address is set to false (0)
type setIfEqualOperation struct {
	*TsvetokVirtualMachine
}

func newSetIfEqualOperation(t *TsvetokVirtualMachine) setIfEqualOperation {
	return setIfEqualOperation{t}
}

func (s setIfEqualOperation) Execute() error {
	leftParam, err := s.getFirstParam()
	if err != nil {
		return err
	}

	rightParam, err := s.getSecondParam()
	if err != nil {
		return err
	}

	outputAddr, err := s.getThirdParam()
	if err != nil {
		return err
	}

	outputVal := 0
	if leftParam.Value == rightParam.Value {
		outputVal = 1
	}

	if outputAddr.Format == ParamFormatAddress {
		return s.SetValueInMemory(outputAddr.Address, outputVal)
	}

	if outputAddr.Format == ParamFormatRegister {
		return s.SetValueInRegisterFile(outputAddr.Address, outputVal)
	}

	return InvalidOutputParamErr{"seq"}
}

func (s setIfEqualOperation) GetNextProgramCounter() int { return s.getProgramCounter() + 4 }

func (s setIfEqualOperation) Halt() bool { return false }
