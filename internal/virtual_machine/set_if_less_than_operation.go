package virtual_machine

import ()

// setIfLessThanOperation multiplies two numbers together
type setIfLessThanOperation struct {
	*TsvetokVirtualMachine
}

func newSetIfLessThanOperation(t *TsvetokVirtualMachine) setIfLessThanOperation {
	return setIfLessThanOperation{t}
}

func (m setIfLessThanOperation) Execute() error {
	lhs, err := m.getFirstParam()
	if err != nil {
		return err
	}

	rhs, err := m.getSecondParam()
	if err != nil {
		return err
	}

	outAddr, err := m.getThirdParam()
	if err != nil {
		return err
	}

	value := 0
	if lhs.Value < rhs.Value {
		value = 1
	}

	if outAddr.Format == ParamFormatAddress {
		return m.SetValueInMemory(outAddr.Address, value)
	}

	if outAddr.Format == ParamFormatRegister {
		return m.SetValueInRegisterFile(outAddr.Address, value)
	}

	return InvalidOutputParamErr{"slt"}
}

func (m setIfLessThanOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 4 }

func (_ setIfLessThanOperation) Halt() bool { return false }

