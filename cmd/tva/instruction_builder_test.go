package tva

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
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
