package virtual_machine

import (
	"fmt"
)

// outputOperation emits the value at the location provided by the first parameter to the
// machine's output interface
type outputOperation struct {
	*TsvetokVirtualMachine
}

func newOutputOperation(t *TsvetokVirtualMachine) outputOperation {
	return outputOperation{t}
}

func (m outputOperation) Execute() error {
	if m.OutputInterface == nil {
		return fmt.Errorf("no output interface set on the machine (use SetOutputInterface)")
	}

	param, err := m.getFirstParam()
	if err != nil {
		return err
	}

	m.EmitOutput(param.Value)

	return nil
}

func (m outputOperation) GetNextProgramCounter() int { return m.getProgramCounter() + 2 }

func (_ outputOperation) Halt() bool { return false }
