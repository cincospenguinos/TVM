package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name           string
		args           []string
		stdin          string
		wantExit       int
		wantStdout     string
		wantStderrPart string
	}{
		{
			name:           "no arguments prints usage",
			args:           []string{},
			wantExit:       2,
			wantStderrPart: "usage",
		},
		{
			name:           "extra arguments print usage",
			args:           []string{"a.tvm", "b.tvm"},
			wantExit:       2,
			wantStderrPart: "usage",
		},
		{
			name:           "missing file fails",
			args:           []string{"../../test_data/missing.tvm"},
			wantExit:       1,
			wantStderrPart: "missing.tvm",
		},
		{
			name:           "bad magic fails",
			args:           []string{"../../test_data/bad_magic.tvm"},
			wantExit:       1,
			wantStderrPart: "magic",
		},
		{
			name:       "halt runs quietly",
			args:       []string{"../../test_data/halt.tvm"},
			wantExit:   0,
			wantStdout: "",
		},
		{
			name:       "add beyond the program runs quietly",
			args:       []string{"../../test_data/add_5_5.tvm"},
			wantExit:   0,
			wantStdout: "",
		},
		{
			name:       "out emits to stdout",
			args:       []string{"../../test_data/out_5.tvm"},
			wantExit:   0,
			wantStdout: "5\n",
		},
		{
			name:       "in echoes through out",
			args:       []string{"../../test_data/in_out.tvm"},
			stdin:      "42\n",
			wantExit:   0,
			wantStdout: "42\n",
		},
		{
			name:       "in returns 0 at end of input",
			args:       []string{"../../test_data/in_out.tvm"},
			stdin:      "",
			wantExit:   0,
			wantStdout: "0\n",
		},
		{
			name:       "in returns 0 on unparseable input",
			args:       []string{"../../test_data/in_out.tvm"},
			stdin:      "abc\n",
			wantExit:   0,
			wantStdout: "0\n",
		},
		{
			name:       "the label/jump sample prints the un-overwritten jmp word",
			args:       []string{"../../test_data/label_jump_test.tvm"},
			wantExit:   0,
			wantStdout: "1106\n",
		},
		{
			name:       "the data labels sample prints the first stored word",
			args:       []string{"../../test_data/data_labels.tvm"},
			wantExit:   0,
			wantStdout: "12\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			exit := run(bytes.NewBufferString(tc.stdin), stdout, stderr, tc.args)

			assert.Equal(t, tc.wantExit, exit)
			assert.Equal(t, tc.wantStdout, stdout.String())
			if tc.wantStderrPart != "" {
				assert.Contains(t, stderr.String(), tc.wantStderrPart)
			}
		})
	}
}
