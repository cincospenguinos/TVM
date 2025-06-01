package tva

import (
	"testing"

	tvm "tvm/internal/virtual_machine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTsvetokAssembler_HandlesHalt(t *testing.T) {
	assembler := NewAssemblerFromString("hlt")
	program, err := assembler.Assemble()
	require.NoError(t, err)

	assert.Equal(t, program[0], 9)
}

func TestTsvetokAssembler_HandlesAllRegularInstructions(t *testing.T) {
	type testCase struct {
		program         string
		expectedAddress int
		expectedValue   int
		testName        string
	}

	testCases := []testCase{
		{"hlt", 0, 9, "hlt instruction works"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			machine := tvm.NewTsvetokVirtualMachine(program)
			require.NoError(t, machine.Execute())

			memory := machine.CopyMemory()
			assert.Equal(t, memory[tc.expectedAddress], tc.expectedValue)
		})
	}
}
