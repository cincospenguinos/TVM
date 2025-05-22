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
