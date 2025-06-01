package tvm

import ()

// addOperation adds two numbers together
type addOperation struct {
	*TsvetokVirtualMachine
}

func newAddOperation(t *TsvetokVirtualMachine) addOperation {
	return addOperation{t}
}

func (a addOperation) Execute() error {
	memory := a.getMemory()
	programCounter := a.getProgramCounter()

	leftParam, err := a.getFirstParam()
	if err != nil {
		return err
	}

	rightParam, err := a.getSecondParam()
	if err != nil {
		return err
	}

	outAddr := memory[programCounter+3]

	memory[outAddr] = leftParam.Value + rightParam.Value

	return nil
}

func (a addOperation) GetNextProgramCounter() int { return a.getProgramCounter() + 4 }

func (_ addOperation) Halt() bool { return false }
