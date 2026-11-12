package main

import (
	"fmt"
	"os"
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

func TestTsvetokAssembler_ResolvedDataLabels(t *testing.T) {
	for _, tc := range []struct {
		program  string
		expected []int
		name     string
	}{
		{
			program:  ".data 12",
			expected: []int{12},
			name:     "a single literal",
		},
		{
			program:  ".data i12",
			expected: []int{12},
			name:     "the immediate prefix is equivalent to the plain number",
		},
		{
			program:  ".data -1",
			expected: []int{-1},
			name:     "a negative literal",
		},
		{
			program:  ".data i-12",
			expected: []int{-12},
			name:     "a negative immediate",
		},
		{
			program:  ".data 1 2 3",
			expected: []int{1, 2, 3},
			name:     "multiple values, in order",
		},
		{
			program:  ".data-12 12",
			expected: []int{12},
			name:     "the name is the entire first chunk, dashes included",
		},
		{
			program:  ".data 12\nhlt",
			expected: []int{12, 9},
			name:     "data words occupy their place in the stream",
		},
		{
			// Finicky, as intended: $. copies the word at the label's index, which is the
			// hlt instruction's opcode word
			program:  ".data $.foo\n.foo\nhlt",
			expected: []int{9, 9},
			name:     "a backward '$.' captures the word at the label's index",
		},
		{
			program:  ".data $.foo\nadd 1 1 $0\n.foo\nhlt",
			expected: []int{9, 1101, 1, 1, 0, 9},
			name:     "a forward '$.' is finalized from its target word",
		},
		{
			program:  ".one $.two\n.two 42\nhlt",
			expected: []int{42, 42, 9},
			name:     "a '$.' chain resolves to the first finalized word",
		},
		{
			program:  ".self\n.data i.self",
			expected: []int{0},
			name:     "an 'i.' self-reference stores the label's own index (no dereference, no cycle)",
		},
		{
			// bar = 5, biz = 6, baz = 8; 'out 7' = 104 7; $.biz captured the out opcode word
			program:  ".crazy-data__69420 i.bar 12 3 $.biz i.baz\n.bar\nhlt\n.biz\nout 7\n.baz",
			expected: []int{5, 12, 3, 104, 8, 9, 104, 7},
			name:     "arbitrary mixes of the forms, as in the owner's example",
		},
		{
			program:  "add 1, 1, $0\n.data 1 2\n.end\nout i.end\nhlt",
			expected: []int{1101, 1, 1, 0, 1, 2, 104, 6, 9},
			name:     "a data label shifts the indexes of labels declared after it",
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

func TestTsvetokAssembler_FailsOnInvalidDataLabels(t *testing.T) {
	for _, tc := range []struct {
		program  string
		errParts []string
		name     string
	}{
		{
			program:  ".data foo",
			errParts: []string{"invalid label"},
			name:     "a bare word is not a data value",
		},
		{
			program:  ".data 1.5",
			errParts: []string{"invalid label"},
			name:     "a float is not a data value",
		},
		{
			program:  ".data $.nope",
			errParts: []string{"nope"},
			name:     "a reference to a never-declared label",
		},
		{
			program:  ".foo\n.foo 12",
			errParts: []string{"already declared"},
			name:     "a duplicate declaration, either form",
		},
		{
			program:  ".a $.b\n.b $.a",
			errParts: []string{"cycle"},
			name:     "a '$.' cycle",
		},
		{
			program:  ".self\n.data $.self",
			errParts: []string{"cycle"},
			name:     "a '$.' self-reference is a cycle",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			_, err := assembler.Assemble()

			require.Error(t, err, "the program should not assemble")
			for _, part := range tc.errParts {
				assert.Contains(t, err.Error(), part, "the error should explain the failure")
			}
		})
	}
}

func TestTsvetokAssembler_RunsWithForwardReferencesToDataLabels(t *testing.T) {
	for _, tc := range []executionTestCase{
		{
			program:         "out $.data\nhlt\n.data 12",
			expectedAddress: -1,
			expectedValue:   12,
			testName:        "'out $.data' prints the first word stored at the data label",
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			assembler := NewAssemblerFromString(tc.program)
			program, err := assembler.Assemble()
			require.NoError(t, err)

			machine := tvm.NewTsvetokVirtualMachine(program)
			mockOutput := &tvm.MockOutputInterface{}
			machine.SetOutputInterface(mockOutput)

			require.NoError(t, machine.Execute(), fmt.Sprintf("failed execution (program was %v)", program))

			require.NotNil(t, mockOutput.LastNumberReceived)
			assert.Equal(t, tc.expectedValue, *mockOutput.LastNumberReceived)
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

func TestTsvetokAssembler_RunsTheLabelJumpSample(t *testing.T) {
	source, err := os.ReadFile("../../test_data/label_jump_test.tva")
	require.NoError(t, err)

	assembler := NewAssemblerFromString(string(source))
	program, err := assembler.Assemble()
	require.NoError(t, err)

	machine := tvm.NewTsvetokVirtualMachine(program)

	mockInput := tvm.MockInputInterface{NumberToReturn: -69}
	machine.SetInputInterface(mockInput)

	mockOutput := &tvm.MockOutputInterface{}
	machine.SetOutputInterface(mockOutput)

	require.NoError(t, machine.Execute())

	// The jmp skips the add, so memory[0] is never overwritten and still holds the
	// jmp's opcode word, which the out then prints
	require.NotNil(t, mockOutput.LastNumberReceived)
	assert.Equal(t, 1106, *mockOutput.LastNumberReceived)

	memory := machine.CopyMemory()
	assert.Equal(t, 1106, memory[0])
}

func TestTsvetokAssembler_RunsTheDataLabelsSample(t *testing.T) {
	source, err := os.ReadFile("../../test_data/data_labels.tva")
	require.NoError(t, err)

	assembler := NewAssemblerFromString(string(source))
	program, err := assembler.Assemble()
	require.NoError(t, err)

	// out $.value, hlt, the data 12 34, and a copy of the first data word
	assert.Equal(t, []int{4, 3, 9, 12, 34, 12}, program)

	machine := tvm.NewTsvetokVirtualMachine(program)

	mockOutput := &tvm.MockOutputInterface{}
	machine.SetOutputInterface(mockOutput)

	require.NoError(t, machine.Execute(), fmt.Sprintf("failed execution (program was %v)", program))

	// The out reads memory[3] (the first word stored at .value); the data itself is
	// never executed, because the hlt comes first
	require.NotNil(t, mockOutput.LastNumberReceived)
	assert.Equal(t, 12, *mockOutput.LastNumberReceived)

	memory := machine.CopyMemory()
	assert.Equal(t, []int{4, 3, 9, 12, 34, 12}, memory)
}

func TestTsvetokAssembler_AcceptsLabelsAsLocationsForJumping(t *testing.T) {
	assemblyStr := `
		jit 1, i.end
		add 0, 0, $0
	.end
		hlt
	`

	assembler := NewAssemblerFromString(assemblyStr)
	program, err := assembler.Assemble()
	require.NoError(t, err)

	// .end resolves to index 7 (the hlt): the jit jumps over the add, which would
	// otherwise have overwritten memory[0] with 0
	assert.Equal(t, []int{1106, 1, 7, 1101, 0, 0, 0, 9}, program)

	machine := tvm.NewTsvetokVirtualMachine(program)
	err = machine.Execute()
	require.NoError(t, err)

	memory := machine.CopyMemory()
	assert.Equal(t, 1106, memory[0])
}
