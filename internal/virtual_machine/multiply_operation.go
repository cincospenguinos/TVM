package virtual_machine

import ()

// multiplyOperation multiplies two nubmers together
type multiplyOperation struct {
	*TsvetokVirtualMachine
}

func newMultiplyOperation(t *TsvetokVirtualMachine) multiplyOperation {
	return multiplyOperation{t}
}

func (m multiplyOperation) Execute() error {
	leftParam, err := m.getFirstParam()
	if err != nil {
		return err
	}

	rightParam, err := m.getSecondParam()
	if err != nil {
		return err
	}

	outAddr, err := m.getThirdParam()
	if err != nil {
		return err
	}

	if outAddr.Format != ParamFormatAddress {
		return InvalidOutputParamErr{"mlt"}
	}

	return m.SetValueInMemory(outAddr.Address, leftParam.Value * rightParam.Value)
}

func (m multiplyOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 4 }

func (_ multiplyOperation) Halt() bool { return false }
