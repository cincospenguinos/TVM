package tva

import (
	"fmt"
	"testing"

	tvm "tvm/internal/virtual_machine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type executionTestCase struct {
	program         string
	expectedAddress int
	expectedValue   int
	testName        string
}

func TestTsvetokAssembler_HandlesAllRegularInstructionsInMemoryMode(t *testing.T) {
	for _, tc := range []executionTestCase{
		{"hlt", 0, 9, "hlt instruction works"},
		{"add $0, $0, $0\nhlt", 0, 2, "add instruction works"},
		{"mlt $0, $0, $1\nhlt", 1, 4, "mlt instruction works"},
		{"in $0\nhlt", 0, -69, "in instruction works"},
		{"out $0\nhlt", -1, 4, "out instruction works"},
		{"seq $1, $4, $1\nhlt", 1, 0, "seq instruction works"},
		{"jit $0, $4\nadd $7, $0, $0\nhlt", 0, 6, "jit instruction works"},
	} {
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
	for _, tc := range []executionTestCase{
		{"add $2, i12, $0\nhlt", 0, 24, "add instruction works"},
		{"add 5, 2, $0\nhlt", 0, 7, "add instruction with plain immediates works"},
		{"mlt 5, 2, $0\nhlt", 0, 10, "mlt instruction with plain immediates works"},
		{"out i12\nhlt", -1, 12, "out instruction with immediate works"},
		{"seq $1, 1, $0\nhlt", 0, 1, "seq instruction works"},
		{"jit $0, 7\nadd $0, $0, $0\nhlt", 0, 1006, "jit with plain immediate address works"},
		{"jit 0, $4\nadd $0, $0, $0\nhlt", 0, 212, "jit with plain immediate parameter works"},
	} {
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

func TestTsvetokAssembler_HandlesAllRegularInstructionsInRegisterMode(t *testing.T) {
	for _, tc := range []executionTestCase{
		{"add i2, r0, r0\nadd r0, r0, $0\nhlt", 0, 4, "add supports register mode"},
	} {
		t.Run(tc.testName, func (t *testing.T) {
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

func TestTsvetokAssembler_IgnoresCommentsSpacesAndTheLike(t *testing.T) {
	program := `# This is a comment. In Tsvetok Assembly we start comments
	# with the '#' character. There are no multi-line comments;
	# only inline comments
		hlt # this is a comment at the end of the line, which should be ignored
	`

	assembler := NewAssemblerFromString(program)
	intcode, err := assembler.Assemble()
	require.NoError(t, err)
	require.True(t, len(intcode) > 0)
	assert.Equal(t, 9, intcode[0])
}
