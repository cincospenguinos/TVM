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
		{"mlt $0, $0, $1\nhlt", 1, 4, "mlt instruction works"},
		{"in $0\nhlt", 0, -69, "in instruction works"},
		{"out $0\nhlt", -1, 4, "out instruction works"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			machine := tvm.NewTsvetokVirtualMachine(program)
			
			mockInput := tvm.MockInputInterface{-69}
			machine.SetInputInterface(mockInput)
			
			mockOutput := &tvm.MockOutputInterface{}
			machine.SetOutputInterface(mockOutput)

			preExecutionMemory := machine.CopyMemory()
			require.NoError(t, machine.Execute(), fmt.Sprintf("failed execution (program was %v)", preExecutionMemory))

			memory := machine.CopyMemory()

			if tc.expectedAddress < 0 { // Assert output case
				require.NotNil(t, mockOutput.LastNumberReceived)
				assert.Equal(t, tc.expectedValue, *mockOutput.LastNumberReceived, "output value was not equal")
			} else {
				assert.Equal(t, tc.expectedValue, memory[tc.expectedAddress])
			}
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
		{"mlt 5, 2, $0\nhlt", 0, 10, "mlt instruction with plain immediates works"},
		{"out i12\nhlt", -1, 12, "out instruction with immediate works"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			machine := tvm.NewTsvetokVirtualMachine(program)
			mockOutput := &tvm.MockOutputInterface{}
			machine.SetOutputInterface(mockOutput)

			preExecutionMemory := machine.CopyMemory()
			require.NoError(t, machine.Execute(), fmt.Sprintf("failed execution (program was %v)", preExecutionMemory))

			if tc.expectedAddress < 0 { // Assertion on output
				require.NotNil(t, mockOutput.LastNumberReceived)
				assert.Equal(t, tc.expectedValue, *mockOutput.LastNumberReceived, "output value was not equal")
			} else {
				memory := machine.CopyMemory()
				assert.Equal(t, tc.expectedValue, memory[tc.expectedAddress])
			}
		})
	}
}
