package tvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockInputInterface struct {
	NumberToReturn int
}

func (m MockInputInterface) ReceiveInput() int {
	return m.NumberToReturn
}

type MockOutputInterface struct {
	LastNumberReceived *int
}

func (m *MockOutputInterface) EmitOutput(number int) {
	m.LastNumberReceived = &number
}

func TestTsvetokVirtualMachine_HaltsProperly(t *testing.T) {
	err := NewTsvetokVirtualMachine([]int{9}).Execute()
	require.NoError(t, err)
}

func TestTsvetokVirtualMachine_AddsProperlyInMemoryMode(t *testing.T) {
	program := []int{1, 0, 0, 0, 9}
	machine := NewTsvetokVirtualMachine(program)
	err := machine.Execute()
	require.NoError(t, err)

	result := machine.CopyMemory()
	assert.Equal(t, 2, result[0])
}

func TestTsvetokVirtualMachine_MultipliesProperlyInMemoryMode(t *testing.T) {
	program := []int{2, 0, 0, 0, 9}
	machine := NewTsvetokVirtualMachine(program)
	err := machine.Execute()
	require.NoError(t, err)

	result := machine.CopyMemory()
	assert.Equal(t, 4, result[0])
}

func TestTsvetokVirtualMachine_GivesErrorForInvalidOpcode(t *testing.T) {
	machine := NewTsvetokVirtualMachine([]int{-1234})
	require.Error(t, machine.Execute())
}

func TestTsvetokVirtualMachine_HandlesInputCorrectly(t *testing.T) {
	program := []int{3, 0, 9}
	machine := NewTsvetokVirtualMachine(program)
	machine.SetInputInterface(MockInputInterface{-1})
	require.NoError(t, machine.Execute())

	result := machine.CopyMemory()
	assert.Equal(t, -1, result[0])
}

func TestTsvetokVirtualMachine_HandlesOutputCorrectly(t *testing.T) {
	mockOutput := &MockOutputInterface{}
	machine := NewTsvetokVirtualMachine([]int{4, 0, 9})
	machine.SetOutputInterface(mockOutput)

	require.NoError(t, machine.Execute())
	require.NotNil(t, mockOutput.LastNumberReceived)
	assert.Equal(t, 4, *mockOutput.LastNumberReceived)
}

func TestTsvetokVirtualMachine_SetIfEqualSetsIfEqualInMemoryMode(t *testing.T) {
	program := []int{5, 0, 0, 0, 9}
	machine := NewTsvetokVirtualMachine(program)
	require.NoError(t, machine.Execute())

	result := machine.CopyMemory()
	assert.Equal(t, 1, result[0])
}

func TestTsvetokVirtualMachine_SetIfEqualSetsToFalseIfNotEqualInMemoryMode(t *testing.T) {
	program := []int{5, 0, 1, 0, 9}
	machine := NewTsvetokVirtualMachine(program)
	require.NoError(t, machine.Execute())

	result := machine.CopyMemory()
	assert.Equal(t, 0, result[0])
}

func TestTsvetokVirtualMachine_JumpIfTrueDoesItsNamesake(t *testing.T) {
	program := []int{6, 0, 8, 1, 0, 0, 0, 9, 7}
	machine := NewTsvetokVirtualMachine(program)
	require.NoError(t, machine.Execute())

	result := machine.CopyMemory()
	assert.Equal(t, 6, result[0])
}

func TestTsvetokVirtualMachine_JumpIfTrueDoesNotJumpIfFalse(t *testing.T) {
	program := []int{6, 4, 8, 1, 0, 0, 0, 9, 7}
	machine := NewTsvetokVirtualMachine(program)
	require.NoError(t, machine.Execute())

	result := machine.CopyMemory()
	assert.Equal(t, 12, result[0])
}

func TestTsvetokVirtualMachine_AllInputParamsSupportImmediateMode(t *testing.T) {
	type testCase struct {
		program          []int
		expectedAddress  int
		expectedValue    int
		testName         string
	}
	testCases := []testCase {
		{[]int{101, 10, 2, 0, 9}, 0, 12,      "add first param immediate"},
		{[]int{1001, 4, 10, 0, 9}, 0, 19,     "add second param immediate"},
		{[]int{1102, 1, 0, 0, 9}, 0, 0,       "mlt both params immediate"},
		{[]int{104, 100, 9}, -1, 100, "out one immediate parameter"},
		{[]int{1105, 1105, 1105, 0, 9}, 0, 1, "seq first param immediate"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func (t *testing.T) {
			mockOutput := &MockOutputInterface{}
			machine := NewTsvetokVirtualMachine(tc.program)
			machine.SetOutputInterface(mockOutput)

			require.NoError(t, machine.Execute())

			memory := machine.CopyMemory()

			if memory[0] % 10 == 4 { // We're testing output
				assert.Equal(t, tc.expectedValue, *mockOutput.LastNumberReceived)
			} else {
				assert.Equal(t, tc.expectedValue, memory[tc.expectedAddress])
			}

		})
	}
}
