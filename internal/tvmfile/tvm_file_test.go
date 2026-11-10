package tvmfile

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMagic is the four-byte TVM magic, used to build the in-memory blobs under test
var testMagic = []byte("TVM\x00")

func TestDecode(t *testing.T) {
	for _, tc := range []struct {
		name        string
		data        []byte
		wantWords   []int
		wantErrPart string
	}{
		{
			name:      "single word",
			data:      append(append([]byte{}, testMagic...), 9, 0, 0, 0),
			wantWords: []int{9},
		},
		{
			name:        "bad magic",
			data:        []byte("XVM\x00\x09\x00\x00\x00"),
			wantErrPart: "magic",
		},
		{
			name:        "shorter than the magic",
			data:        []byte("TV"),
			wantErrPart: "magic",
		},
		{
			name:        "truncated word",
			data:        append(append([]byte{}, testMagic...), 1, 0, 0),
			wantErrPart: "truncated",
		},
		{
			name:      "magic only is an empty program",
			data:      append([]byte{}, testMagic...),
			wantWords: []int{},
		},
		{
			name:      "negative word sign-extends",
			data:      append(append([]byte{}, testMagic...), 0xFF, 0xFF, 0xFF, 0xFF),
			wantWords: []int{-1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			words, err := Decode(tc.data)

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

func TestEncode(t *testing.T) {
	for _, tc := range []struct {
		name        string
		program     []int
		wantBytes   []byte
		wantErrPart string
	}{
		{
			name:      "single word",
			program:   []int{9},
			wantBytes: append(append([]byte{}, testMagic...), 9, 0, 0, 0),
		},
		{
			name:      "empty program is just the magic",
			program:   []int{},
			wantBytes: append([]byte{}, testMagic...),
		},
		{
			name:      "negative word is two's complement",
			program:   []int{-1},
			wantBytes: append(append([]byte{}, testMagic...), 0xFF, 0xFF, 0xFF, 0xFF),
		},
		{
			name:      "maximum word",
			program:   []int{math.MaxInt32},
			wantBytes: append(append([]byte{}, testMagic...), 0xFF, 0xFF, 0xFF, 0x7F),
		},
		{
			name:      "minimum word",
			program:   []int{math.MinInt32},
			wantBytes: append(append([]byte{}, testMagic...), 0x00, 0x00, 0x00, 0x80),
		},
		{
			name:        "word above the maximum errors",
			program:     []int{1 << 31},
			wantErrPart: "32-bit",
		},
		{
			name:        "word below the minimum errors",
			program:     []int{math.MinInt32 - 1},
			wantErrPart: "32-bit",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Encode(tc.program)

			if tc.wantErrPart != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrPart)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantBytes, data)
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	for _, program := range [][]int{
		{},
		{9},
		{1101, 5, 5, 10, 9},
		{-1, 0, 1, math.MaxInt32, math.MinInt32},
	} {
		t.Run(fmt.Sprintf("%v", program), func(t *testing.T) {
			data, err := Encode(program)
			require.NoError(t, err)

			words, err := Decode(data)
			require.NoError(t, err)
			assert.Equal(t, program, words)
		})
	}
}

func TestReadFile(t *testing.T) {
	for _, tc := range []struct {
		name        string
		fileName    string
		fileBytes   []byte
		wantWords   []int
		wantErrPart string
	}{
		{
			name:      "reads a program",
			fileName:  "program.tvm",
			fileBytes: append(append([]byte{}, testMagic...), 9, 0, 0, 0),
			wantWords: []int{9},
		},
		{
			name:        "bad magic errors",
			fileName:    "program.tvm",
			fileBytes:   []byte("XVM\x00\x09\x00\x00\x00"),
			wantErrPart: "magic",
		},
		{
			name:        "missing file errors",
			fileName:    "missing.tvm",
			wantErrPart: "missing.tvm",
		},
		{
			name:      "magic only is an empty program",
			fileName:  "program.tvm",
			fileBytes: append([]byte{}, testMagic...),
			wantWords: []int{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tc.fileName)
			if tc.fileBytes != nil {
				require.NoError(t, os.WriteFile(path, tc.fileBytes, 0o644))
			}

			words, err := ReadFile(path)

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

func TestWriteFile(t *testing.T) {
	t.Run("writes the magic and the words", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "program.tvm")
		wantBytes := append(append([]byte{}, testMagic...), 9, 0, 0, 0)

		require.NoError(t, WriteFile(path, []int{9}))

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, wantBytes, data)
	})

	t.Run("round trips through ReadFile", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "program.tvm")
		program := []int{1101, 5, 5, 10, 9}

		require.NoError(t, WriteFile(path, program))

		words, err := ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, program, words)
	})

	t.Run("an overflowing word errors without writing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "program.tvm")

		require.Error(t, WriteFile(path, []int{1 << 31}))
		_, statErr := os.Stat(path)
		assert.Error(t, statErr)
	})
}
