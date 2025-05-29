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

func TestTsvetokVirtualMachine_SetIfEqualDoesNotSetIfNotEqualInMemoryMode(t *testing.T) {
	program := []int{5, 0, 1, 0, 9}
	machine := NewTsvetokVirtualMachine(program)
	require.NoError(t, machine.Execute())

	result := machine.CopyMemory()
	assert.Equal(t, 5, result[0])
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
