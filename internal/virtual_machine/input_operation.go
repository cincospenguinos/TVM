package virtual_machine

import ()

// inputOperation multiplies two numbers together
type inputOperation struct {
	*TsvetokVirtualMachine
}

func newInputOperation(t *TsvetokVirtualMachine) inputOperation {
	return inputOperation{t}
}

func (m inputOperation) Execute() error {
	address, err := m.getFirstParam()
	if err != nil {
		return err
	}

	number := m.ReceiveInput()

	if address.Format == ParamFormatAddress {
		return m.SetValueInMemory(address.Address, number)
	}

	if address.Format == ParamFormatRegister {
		return m.SetValueInRegisterFile(address.Address, number)
	}

	return InvalidOutputParamErr{"in"}
}

func (m inputOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 2 }

func (_ inputOperation) Halt() bool { return false }
