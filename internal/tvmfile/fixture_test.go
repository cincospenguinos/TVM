package tvmfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixturePath resolves a file from the repository's test_data directory
func fixturePath(name string) string {
	return filepath.Join("..", "..", "test_data", name)
}

func TestDecodeFixtures(t *testing.T) {
	for _, tc := range []struct {
		name        string
		fixture     string
		wantWords   []int
		wantErrPart string
	}{
		{name: "halt", fixture: "halt.tvm", wantWords: []int{9}},
		{name: "add_5_5", fixture: "add_5_5.tvm", wantWords: []int{1101, 5, 5, 10, 9}},
		{name: "out_5", fixture: "out_5.tvm", wantWords: []int{104, 5, 9}},
		{name: "in_out", fixture: "in_out.tvm", wantWords: []int{203, 0, 204, 0, 9}},
		{name: "empty", fixture: "empty.tvm", wantWords: []int{}},
		{name: "bad_magic", fixture: "bad_magic.tvm", wantErrPart: "magic"},
		{name: "truncated_word", fixture: "truncated_word.tvm", wantErrPart: "truncated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(fixturePath(tc.fixture))
			require.NoError(t, err)

			words, err := Decode(data)

			if tc.wantErrPart != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrPart)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantWords, words)
		})
	}
}

func TestEncodeMatchesFixtures(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fixture string
		program []int
	}{
		{name: "halt", fixture: "halt.tvm", program: []int{9}},
		{name: "add_5_5", fixture: "add_5_5.tvm", program: []int{1101, 5, 5, 10, 9}},
		{name: "out_5", fixture: "out_5.tvm", program: []int{104, 5, 9}},
		{name: "in_out", fixture: "in_out.tvm", program: []int{203, 0, 204, 0, 9}},
		{name: "empty", fixture: "empty.tvm", program: []int{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wantBytes, err := os.ReadFile(fixturePath(tc.fixture))
			require.NoError(t, err)

			gotBytes, err := Encode(tc.program)
			require.NoError(t, err)
			assert.Equal(t, wantBytes, gotBytes)
		})
	}
}
