package tvm

import ()

// outputOperation multiplies two nubmers together
type outputOperation struct {
	*TsvetokVirtualMachine
}

func newOutputOperation(t *TsvetokVirtualMachine) outputOperation {
	return outputOperation{t}
}

func (m outputOperation) Execute() error {
	param, err := m.getFirstParam()
	if err != nil {
		return err
	}

	m.EmitOutput(param.Value)

	return nil
}

func (m outputOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 2 }

func (_ outputOperation) Halt() bool { return false }
