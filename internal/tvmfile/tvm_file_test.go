package tvmfile

import (
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
