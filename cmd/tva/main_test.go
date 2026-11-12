package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name              string
		inputFixture      string
		outputRelPath     string
		wantExit          int
		wantStderrPart    string
		wantOutputFixture string
	}{
		{
			name:           "no arguments prints usage",
			wantExit:       2,
			wantStderrPart: "usage",
		},
		{
			name:           "one argument prints usage",
			inputFixture:   "../../test_data/out_5.tva",
			wantExit:       2,
			wantStderrPart: "usage",
		},
		{
			name:           "missing input file fails",
			inputFixture:   "../../test_data/missing.tva",
			outputRelPath:  "out.tvm",
			wantExit:       1,
			wantStderrPart: "missing.tva",
		},
		{
			name:           "unknown instruction fails assembly",
			inputFixture:   "../../test_data/bad_instruction.tva",
			outputRelPath:  "out.tvm",
			wantExit:       1,
			wantStderrPart: "unknown instruction",
		},
		{
			name:              "assembles and writes the binary file",
			inputFixture:      "../../test_data/out_5.tva",
			outputRelPath:     "out.tvm",
			wantExit:          0,
			wantOutputFixture: "../../test_data/out_5.tvm",
		},
		{
			name:              "assembles the label/jump sample to its fixture",
			inputFixture:      "../../test_data/label_jump_test.tva",
			outputRelPath:     "out.tvm",
			wantExit:          0,
			wantOutputFixture: "../../test_data/label_jump_test.tvm",
		},
		{
			name:              "assembles the data labels sample to its fixture",
			inputFixture:      "../../test_data/data_labels.tva",
			outputRelPath:     "out.tvm",
			wantExit:          0,
			wantOutputFixture: "../../test_data/data_labels.tvm",
		},
		{
			name:           "invalid label declaration fails",
			inputFixture:   "../../test_data/bad_label.tva",
			outputRelPath:  "out.tvm",
			wantExit:       1,
			wantStderrPart: "invalid label",
		},
		{
			name:           "unwritable output path fails",
			inputFixture:   "../../test_data/out_5.tva",
			outputRelPath:  filepath.Join("no_such_dir", "out.tvm"),
			wantExit:       1,
			wantStderrPart: "no_such_dir",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			dir := t.TempDir()

			var args []string
			if tc.inputFixture != "" {
				args = append(args, tc.inputFixture)
			}
			if tc.outputRelPath != "" {
				args = append(args, filepath.Join(dir, tc.outputRelPath))
			}

			exit := run(bytes.NewBufferString(""), stdout, stderr, args)

			assert.Equal(t, tc.wantExit, exit)
			if tc.wantStderrPart != "" {
				assert.Contains(t, stderr.String(), tc.wantStderrPart)
			}

			if tc.wantOutputFixture != "" {
				got, err := os.ReadFile(filepath.Join(dir, tc.outputRelPath))
				require.NoError(t, err)

				want, err := os.ReadFile(tc.wantOutputFixture)
				require.NoError(t, err)

				assert.Equal(t, want, got)
			}
		})
	}
}
