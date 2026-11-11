package main

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

			mockInput := tvm.MockInputInterface{NumberToReturn: -69}
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
		{"mlt i2, i3, r0\nadd r0, r0, $0\nhlt", 0, 12, "mlt supports register mode"},
		{"in t0\nadd t0, 0, $0\nhlt", 0, -3, "in supports register mode (temporary register)"},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			machine := tvm.NewTsvetokVirtualMachine(program)
			mockOutput := &tvm.MockOutputInterface{}
			machine.SetOutputInterface(mockOutput)

			mockInput := &tvm.MockInputInterface{NumberToReturn: -3}
			machine.SetInputInterface(mockInput)

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

func TestTsvetokAssembler_StoringFarBeyondTheProgramExpandsMemory(t *testing.T) {
	assembler := NewAssemblerFromString("add 5, 5, $100\nhlt")
	program, err := assembler.Assemble()
	require.NoError(t, err)
	assert.Equal(t, []int{1101, 5, 5, 100, 9}, program)

	machine := tvm.NewTsvetokVirtualMachine(program)
	require.NoError(t, machine.Execute())

	memory := machine.CopyMemory()
	assert.Len(t, memory, 101)
	assert.Equal(t, 10, memory[100])
}

func TestTsvetokAssembler_FailsWhenReferencingAnUndeclaredLabel(t *testing.T) {
	assembler := NewAssemblerFromString("add i.nope, 0, $0\nhlt")

	_, err := assembler.Assemble()

	require.Error(t, err, "referencing a label that was never declared should fail")
	assert.Contains(t, err.Error(), "nope")
	assert.Contains(t, err.Error(), "1") // the line of the reference
}

func TestTsvetokAssembler_ResolvedLabelReferences(t *testing.T) {
	for _, tc := range []struct {
		program  string
		expected []int
		name     string
	}{
		{
			program:  "hlt\n.main",
			expected: []int{9},
			name:     "a standalone label emits no words",
		},
		{
			program:  "add 1, 1, $0\n.end\njit 1, i.end\nhlt",
			expected: []int{1101, 1, 1, 0, 1106, 1, 4, 9},
			name:     "a backward reference resolves to the index after the instruction before the label",
		},
		{
			program:  "jit 1, i.end\nadd 1, 1, $0\n.end\nhlt",
			expected: []int{1106, 1, 7, 1101, 1, 1, 0, 9},
			name:     "a forward reference resolves to the same index",
		},
		{
			program:  "jit 1, $.end\nadd 1, 1, $0\n.end\nhlt",
			expected: []int{106, 1, 7, 1101, 1, 1, 0, 9},
			name:     "a $. form reference uses address format",
		},
		{
			// add=4 words, hlt=1 word, out=2 words, in=2 words: .mid lands at 4, .end at 7
			program:  "add 1, 1, $0\n.mid\nhlt\nout $.end\n.end\nin i.mid",
			expected: []int{1101, 1, 1, 0, 9, 4, 7, 103, 4},
			name:     "label indexes accumulate across mixed instruction word counts",
		},
		{
			program: "  jmp i.main # jump to .main, skipping the add\n" +
				"  add 10, 10, $0 # overwriting the jmp word, which should not execute\n" +
				".main\n" +
				"  out $0\n" +
				"  hlt\n",
			expected: []int{1106, 1, 7, 1101, 10, 10, 0, 4, 0, 9},
			name:     "the sample label/jump program assembles to its expected intcode",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			assert.Equal(t, tc.expected, program)
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

func TestTsvetokAssembler_AcceptsLabelsAsLocationsForJumping(t *testing.T) {
	t.Skip("TODO: This is an integration heavy test")
	assemblyStr := `
		jit 1, i.end
		add 0, 0, $0
	.end
		hlt
	`

	assembler := NewAssemblerFromString(assemblyStr)
	program, err := assembler.Assemble()
	require.NoError(t, err)

	machine := tvm.NewTsvetokVirtualMachine(program)
	err = machine.Execute()
	require.NoError(t, err)

	memory := machine.CopyMemory()
	assert.NotEqual(t, 0, memory[0])
}
