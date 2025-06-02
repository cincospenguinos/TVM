package virtual_machine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type executionTestCase struct {
	program          []int
	expectedAddress  int
	expectedRegister int
	expectedValue    int
	testName         string
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
	for _, tc := range []executionTestCase {
		{program: []int{101, 10, 2, 0, 9}, expectedAddress: 0, expectedValue: 12, testName: "add first param immediate"},
		{program: []int{1001, 4, 10, 0, 9}, expectedAddress: 0, expectedValue: 19, testName: "add second param immediate"},
		{program: []int{1102, 1, 0, 0, 9}, expectedAddress: 0, expectedValue: 0, testName: "mlt both params immediate"},
		{program: []int{104, 100, 9}, expectedAddress: -1, expectedValue: 100, testName: "out one immediate parameter"},
		{program: []int{1105, 1105, 1105, 0, 9}, expectedAddress: 0, expectedValue: 1, testName: "seq first param immediate"},
		{program: []int{106, 1, 8, 0, 0, 0, 1, 9, 7}, expectedAddress: 1, expectedValue: 1, testName: "jit first param immediate"},
		{program: []int{1006, 1, 7, 0, 0, 0, 1, 9, 7}, expectedAddress: 1, expectedValue: 1, testName: "jit second param immediate"},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			mockOutput := &MockOutputInterface{}
			machine := NewTsvetokVirtualMachine(tc.program)
			machine.SetOutputInterface(mockOutput)

			require.NoError(t, machine.Execute())

			memory := machine.CopyMemory()
			if memory[0]%10 == 4 { // We're testing output
				assert.Equal(t, tc.expectedValue, *mockOutput.LastNumberReceived)
			} else {
				assert.Equal(t, tc.expectedValue, memory[tc.expectedAddress])
			}
		})
	}
}

func TestTsvetokVirtualMachine_NoOutputParamSupportsImmediateMode(t *testing.T) {
	for _, tc := range []executionTestCase {
		{program: []int{10001, 0, 0, 12, 9}, testName: "add output param immediate"},
		{program: []int{10002, 0, 0, -69, 9}, testName: "mlt output param immediate"},
		{program: []int{103, -1, 9}, testName: "in output param immediate"},
		{program: []int{10005, 0, 0, 420, 9}, testName: "seq output param immediate"},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			machine := NewTsvetokVirtualMachine(tc.program)
			err := machine.Execute()
			require.Error(t, err)

			_, isInvalidParamErr := err.(InvalidOutputParamErr)
			require.True(t, isInvalidParamErr, "error provided must be InvalidOutputParamErr")
		})
	}
}

func TestTsvetokVirtualMachine_RegisterModeIsSupportedEverywhere(t *testing.T) {
	for _, tc := range []executionTestCase {
		{program: []int{21201, 0, 1, 0, 9}, expectedRegister: 0, expectedValue: 1, testName: "add registers most params"},
	} {
		mockOutput := &MockOutputInterface{}
		machine := NewTsvetokVirtualMachine(tc.program)
		machine.SetOutputInterface(mockOutput)

		require.NoError(t, machine.Execute())

		actualValue, err := machine.GetValueInRegisterFile(tc.expectedRegister)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedValue, actualValue)
	}
}
