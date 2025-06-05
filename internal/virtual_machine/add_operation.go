package virtual_machine

import ()

// addOperation adds two numbers together
type addOperation struct {
	*TsvetokVirtualMachine
}

func newAddOperation(t *TsvetokVirtualMachine) addOperation {
	return addOperation{t}
}

func (a addOperation) Execute() error {
	leftParam, err := a.getFirstParam()
	if err != nil {
		return err
	}

	rightParam, err := a.getSecondParam()
	if err != nil {
		return err
	}

	outAddr, err := a.getThirdParam()
	if err != nil {
		return err
	}

	if outAddr.Format == ParamFormatAddress {
		return a.SetValueInMemory(outAddr.Address, leftParam.Value+rightParam.Value)
	}

	if outAddr.Format == ParamFormatRegister {
		return a.SetValueInRegisterFile(outAddr.Address, leftParam.Value+rightParam.Value)
	}

	return InvalidOutputParamErr{"add"}
}

func (a addOperation) GetNextProgramCounter() int { return a.getProgramCounter() + 4 }

func (_ addOperation) Halt() bool { return false }
