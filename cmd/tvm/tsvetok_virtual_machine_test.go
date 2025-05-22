package tvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTsvetokVirtualMachine_HaltsProperly(t *testing.T) {
	err := NewTsvetokVirtualMachine([]int{9}).Execute()
	require.NoError(t, err)
}

func TestTsvetokVirtualMachine_AddsProperlyInMemoryMode(t *testing.T) {
	program := []int{1, 0, 0, 0, 9}
	machine := NewTsvetokVirtualMachine(program)
	err := machine.Execute()
	require.NoError(t, err)

	result := machine.GetMemory()
	assert.Equal(t, 2, result[0])
}
