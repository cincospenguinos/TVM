package tvm

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
	memory := s.getMemory()

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

	if outputAddr.Format != ParamFormatAddress {
		return InvalidOutputParamErr{"seq"}
	}

	if leftParam.Value == rightParam.Value {
		memory[outputAddr.Address] = 1
	} else {
		memory[outputAddr.Address] = 0
	}

	return nil
}

func (s setIfEqualOperation) GetNextProgramCounter() int { return s.getProgramCounter() + 4 }

func (s setIfEqualOperation) Halt() bool { return false }
