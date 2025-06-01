package tvm

import (
	"fmt"
)

// addOperation adds two numbers together
type addOperation struct {
	*TsvetokVirtualMachine
}

func newAddOperation(t *TsvetokVirtualMachine) addOperation {
	return addOperation{t}
}

func (a addOperation) Execute() error {
	memory := a.getMemory()

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

	if outAddr.Format != ParamFormatAddress {
		return fmt.Errorf("output parameter for add is not in address format")
	}

	memory[outAddr.Address] = leftParam.Value + rightParam.Value

	return nil
}

func (a addOperation) GetNextProgramCounter() int { return a.getProgramCounter() + 4 }

func (_ addOperation) Halt() bool { return false }
