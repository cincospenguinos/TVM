package virtual_machine

import (
	"fmt"
)

// inputOperation reads a number from the machine's input interface and stores it at the
// location provided by the first parameter
type inputOperation struct {
	*TsvetokVirtualMachine
}

func newInputOperation(t *TsvetokVirtualMachine) inputOperation {
	return inputOperation{t}
}

func (m inputOperation) Execute() error {
	if m.InputInterface == nil {
		return fmt.Errorf("no input interface set on the machine (use SetInputInterface)")
	}

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
