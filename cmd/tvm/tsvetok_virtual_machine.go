package tvm

import ()

type TsvetokVirtualMachine struct {
	memory []int
	programCounter int
}

func NewTsvetokVirtualMachine(program []int) *TsvetokVirtualMachine {
	return &TsvetokVirtualMachine{
		memory: program,
		programCounter: 0,
	}
}

func (t *TsvetokVirtualMachine) Execute() error {
	for {
		currentOperation := t.memory[t.programCounter]

		if currentOperation == 1 { // ADD
			leftAddr := t.memory[t.programCounter + 1]
			rightAddr := t.memory[t.programCounter + 2]
			outAddr := t.memory[t.programCounter + 3]

			t.memory[outAddr] = t.memory[leftAddr] + t.memory[rightAddr]
			t.programCounter += 4
		} else if currentOperation == 9 { // HALT
			break
		}
	}

	return nil
}

func (t *TsvetokVirtualMachine) GetMemory() []int {
	return t.memory
}
