package tva

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstructionBuilder_AcceptsValidOpCodesAndRejectsInvalidOnes(t *testing.T) {
	type builderOpCodeTestCase struct{
		opCode  string
		isValid bool
	}

	for _, tc := range []builderOpCodeTestCase {
		{"add", true},
		{"mlt", true},
		{"seq", true},
		{"jit", true},
		{"in", true},
		{"out", true},
		{"hlt", true},
		{"nope", false},
		{".label", false},
	} {
		builder := &instructionBuilder{}
		err := builder.setOperation(tc.opCode)

		if tc.isValid {
			assert.Nil(t, err, fmt.Sprintf("'%v' is meant to be a valid opcode", tc.opCode))
			continue
		}

		assert.NotNil(t, err, fmt.Sprintf("'%v' is not meant to be a valid opcode", tc.opCode))
	}
}

func TestInstructionBuilder_RejectsInvalidParamIndexes(t *testing.T) {
	builder := &instructionBuilder{}

	require.Error(t, builder.addParam("0", 3))
	require.Error(t, builder.addParam("0", -1))
}

func TestInstructionBuilder_AddParamRespectsIndex(t *testing.T) {
	builder := &instructionBuilder{}
	builder.setOperation("add")

	require.NoError(t, builder.addParam("i2", 0))
	require.NoError(t, builder.addParam("$0", 2))
	require.NoError(t, builder.addParam("r1", 1))

	instruction := builder.toIntcode()
	assert.Equal(t, []int{2101, 2, 1, 0}, instruction)
}
