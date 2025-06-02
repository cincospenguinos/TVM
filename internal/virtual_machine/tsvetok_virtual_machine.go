package virtual_machine

import (
	"fmt"
)

const (
	RegisterLastAddress = 13
)

// TsvetokVirtualMachine is an implementation of the Tsvetok Virtual Machine Intcode machine (or TVM.)
type TsvetokVirtualMachine struct {
	memory         []int
	registerFile   []int
	programCounter int
	InputInterface
	OutputInterface
}

func NewTsvetokVirtualMachine(program []int) *TsvetokVirtualMachine {
	return &TsvetokVirtualMachine{
		memory:         program,
		registerFile:   make([]int, 14),
		programCounter: 0,
	}
}

func (t *TsvetokVirtualMachine) Execute() error {
	for {
		currentOperation := t.getCurrentOperation()
		if currentOperation == nil {
			return fmt.Errorf(`no operation found for opcode "%v"`, t.memory[t.programCounter])
		}

		err := currentOperation.Execute()
		if err != nil {
			return err
		}

		if currentOperation.Halt() {
			break
		}

		t.programCounter = currentOperation.GetNextProgramCounter()
	}

	return nil
}

func (t *TsvetokVirtualMachine) getCurrentOperation() TVMOperation {
	rawOpcode := t.memory[t.programCounter]
	opCode := rawOpcode % 100

	switch opCode {
	case 1:
		return newAddOperation(t)
	case 2:
		return newMultiplyOperation(t)
	case 3:
		return newInputOperation(t)
	case 4:
		return newOutputOperation(t)
	case 5:
		return newSetIfEqualOperation(t)
	case 6:
		return newJumpIfTrueOperation(t)
	case 9:
		return newHaltOperation(t)
	default:
		return nil
	}
}

// getMemory returns the TVM's underlying memory. Writing to this slice is persisted across
// the lifetime of the struct.
func (t *TsvetokVirtualMachine) getMemory() []int {
	return t.memory
}

func (t *TsvetokVirtualMachine) GetValueInMemory(address int) (int, error) {
	if address >= 0 && address < len(t.memory) {
		return t.memory[address], nil
	}

	return -1, fmt.Errorf("cannot lookup memory at address '%v' (memory is of size '%v')", address, len(t.memory))
}

func (t *TsvetokVirtualMachine) SetValueInMemory(address, value int) error {
	if address >= 0 && address < len(t.memory) {
		t.memory[address] = value
		return nil
	}

	return fmt.Errorf("cannot write to memory at address '%v' (memory is of size '%v')", address, len(t.memory))
}

func (t *TsvetokVirtualMachine) GetValueInRegisterFile(address int) (int, error) {
	if address >= 0 && address < len(t.registerFile) {
		return t.registerFile[address], nil
	}

	return 0, fmt.Errorf("cannot lookup register at address '%v' (register file is of size '%v')", address, len(t.registerFile))
}

func (t *TsvetokVirtualMachine) SetValueInRegisterFile(address, value int) error {
	if address == RegisterLastAddress {
		return AttemptedLastAddressWriteErr{}
	}

	if address >= 0 && address < len(t.registerFile) {
		t.registerFile[address] = value
		return nil
	}

	return fmt.Errorf("cannot write to register file at address '%v' (register file is of size '%v')", address, len(t.registerFile))
}

func (t *TsvetokVirtualMachine) getFirstParam() (operationParam, error) {
	rawOpcode := t.memory[t.programCounter]
	paramFormat := (rawOpcode / 100) % 10

	return newOperationParam(t, paramFormat, t.programCounter+1)
}

func (t *TsvetokVirtualMachine) getSecondParam() (operationParam, error) {
	rawOpcode := t.memory[t.programCounter]
	paramFormat := (rawOpcode / 1000) % 10

	return newOperationParam(t, paramFormat, t.programCounter+2)
}

func (t *TsvetokVirtualMachine) getThirdParam() (operationParam, error) {
	rawOpcode := t.memory[t.programCounter]
	paramFormat := rawOpcode / 10000

	return newOperationParam(t, paramFormat, t.programCounter+3)
}

func (t *TsvetokVirtualMachine) getProgramCounter() int {
	return t.programCounter
}

// CopyMemory returns a copy of the TVM's current memory state. If an internal function
// wishes to write to memory, use getMemory() instead.
func (t *TsvetokVirtualMachine) CopyMemory() []int {
	copiedMemory := make([]int, len(t.memory))
	copy(copiedMemory, t.memory)

	return copiedMemory
}

func (t *TsvetokVirtualMachine) SetInputInterface(i InputInterface) {
	t.InputInterface = i
}

func (t *TsvetokVirtualMachine) SetOutputInterface(o OutputInterface) {
	t.OutputInterface = o
}
