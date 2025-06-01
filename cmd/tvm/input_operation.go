package tvm

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

	if address.Format != ParamFormatAddress {
		return InvalidOutputParamErr{"in"}
	}

	memory := m.getMemory()
	number := m.ReceiveInput()
	memory[address.Address] = number

	return nil
}

func (m inputOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 2 }

func (_ inputOperation) Halt() bool { return false }
