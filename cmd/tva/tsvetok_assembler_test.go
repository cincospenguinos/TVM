package tva

import (
	"fmt"
	"testing"

	tvm "tvm/internal/virtual_machine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTsvetokAssembler_HandlesAllRegularInstructionsInMemoryMode(t *testing.T) {
	type testCase struct {
		program         string
		expectedAddress int
		expectedValue   int
		testName        string
	}

	testCases := []testCase{
		{"hlt", 0, 9, "hlt instruction works"},
		{"add $0, $0, $0\nhlt", 0, 2, "add instruction works"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			machine := tvm.NewTsvetokVirtualMachine(program)
			preExecutionMemory := machine.CopyMemory()
			require.NoError(t, machine.Execute(), fmt.Sprintf("failed execution (program was %v)", preExecutionMemory))

			memory := machine.CopyMemory()
			assert.Equal(t, memory[tc.expectedAddress], tc.expectedValue)
		})
	}
}

func TestTsvetokAssembler_HandlesAllRegularInstructionsInImmediateMode(t *testing.T) {
	type testCase struct {
		program         string
		expectedAddress int
		expectedValue   int
		testName        string
	}

	testCases := []testCase{
		{"add $2, i12, $0\nhlt", 0, 24, "add instruction works"},
		{"add 5, 2, $0\nhlt", 0, 7, "add instruction with plain immediates works"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			machine := tvm.NewTsvetokVirtualMachine(program)
			preExecutionMemory := machine.CopyMemory()
			require.NoError(t, machine.Execute(), fmt.Sprintf("failed execution (program was %v)", preExecutionMemory))

			memory := machine.CopyMemory()
			assert.Equal(t, memory[tc.expectedAddress], tc.expectedValue)
		})
	}
}
