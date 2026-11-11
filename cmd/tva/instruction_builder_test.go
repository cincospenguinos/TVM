package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstructionBuilder_AcceptsValidOpCodesAndRejectsInvalidOnes(t *testing.T) {
	type builderOpCodeTestCase struct {
		opCode  string
		isValid bool
	}

	for _, tc := range []builderOpCodeTestCase{
		{"add", true},
		{"mlt", true},
		{"seq", true},
		{"jit", true},
		{"jmp", true},
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

func TestInstructionBuilder_JmpExpandsToJitWithConstantTrueCondition(t *testing.T) {
	for _, tc := range []struct {
		targetParam string
		expected    []int
	}{
		{"7", []int{1106, 1, 7}},
		{"$0", []int{106, 1, 0}},
		{"r0", []int{2106, 1, 0}},
	} {
		t.Run(tc.targetParam, func(t *testing.T) {
			builder := &instructionBuilder{}
			require.NoError(t, builder.setOperation("jmp"))
			require.NoError(t, builder.addParam(tc.targetParam, 0))

			assert.Equal(t, tc.expected, builder.toIntcode())
		})
	}
}

func TestInstructionBuilder_JmpTakesExactlyOneParameter(t *testing.T) {
	t.Run("a second target parameter is rejected", func(t *testing.T) {
		builder := &instructionBuilder{}
		require.NoError(t, builder.setOperation("jmp"))
		require.NoError(t, builder.addParam("7", 0))

		assert.Error(t, builder.addParam("8", 1), "jit only reads slot 1, so the extra word would corrupt the next instruction pointer")
	})

	t.Run("a missing target parameter is rejected", func(t *testing.T) {
		builder := &instructionBuilder{}
		require.NoError(t, builder.setOperation("jmp"))

		assert.Error(t, builder.finish(0), "jmp has no target to jump to")
	})

	t.Run("a single parameter is fine", func(t *testing.T) {
		builder := &instructionBuilder{}
		require.NoError(t, builder.setOperation("jmp"))
		require.NoError(t, builder.addParam("7", 0))

		assert.NoError(t, builder.finish(1))
	})
}
