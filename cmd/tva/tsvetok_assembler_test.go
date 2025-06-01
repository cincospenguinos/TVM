package tva

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTsvetokAssembler_HandlesHalt(t *testing.T) {
	assembler := NewAssemblerFromString("hlt")
	program, err := assembler.Assemble()
	
	require.NoError(t, err)
	assert.Equal(t, program[0], 9)
}
