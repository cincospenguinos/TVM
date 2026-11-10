# Binary File I/O

**Status:** Not started
**Source:** README "File Format" section — *"TVM files are binary files with all bytes in little-endian. They begin with the ASCII characters `TVM\0` after which are sequences of instructions. Four bytes (32 bits) is one word in TVM, and so integers are received and munged in four byte chunks, until the file ends"* — plus the two TODOs it leaves open: line 64 (TVM) *Read a TVM binary file and execute it* and line 95 (TVA) *Writes to a TVM binary file with correct syntax*.

## Decisions (from plan review)

1. **The tva file-write side is included** in this feature (README line 95) — it is the write half of "file I/O."
2. **All work lands on the current branch** (`feature/let-s-see-what-a-local-llm-can-do`). Longer-term branch procedures (merging the auto-expanding memory feature, etc.) are deferred to a separate discussion.
3. **Words are signed 32-bit.** `Encode` rejects values outside the `int32` range; `Decode` sign-extends via `int32`. Confirmed.
4. **`in` returns 0 on EOF or unparseable input** — the `InputInterface` cannot carry an error, so this is the documented lenient default. Confirmed.
5. **Codec before fixtures.** The `internal/tvmfile` package is built and tested (with in-memory bytes) before any fixture files exist.
6. **Nine fixtures** per the table below — tentative; revisit if the set proves awkward.

## Spec

A TVM binary file is: the four ASCII magic bytes `TVM\0`, followed by zero or more little-endian **signed** 32-bit words.

- **Decode** — first 4 bytes must be exactly `TVM\0` (else error mentioning "magic"); bytes after the magic must be a multiple of 4 (else error mentioning "truncated"); a magic-only file is valid and decodes to an empty program; a 0-byte file fails the magic check.
- **Encode** — magic first, then each word as a 4-byte little-endian two's-complement value; a word outside `[math.MinInt32, math.MaxInt32]` is an error.
- **Where it lives** — new package `internal/tvmfile` (pure; no dependency on `virtual_machine`, so `cmd/tvm` and `cmd/tva` can both import it without cycles): exported `Decode([]byte) ([]int, error)`, `Encode([]int) ([]byte, error)`, `ReadFile(path) ([]int, error)`, `WriteFile(path, []int) error`.
- **Fixture convention** — `test_data/` at the repo root holds every test input, treated as committed, read-only ground truth: tests *read* from it; anything a test *writes* goes to `t.TempDir()` so the repo never accrues artifacts. A `.gitattributes` line (`*.tvm binary`) pins the binary files against CRLF mangling.
- **Ground truth** — the binary fixtures are hand-authored byte sequences, independent of `Encode`; the encode tests assert byte-for-byte equality *against the fixtures*, so round-trip tests are corroborating evidence, not the primary one.

## Baseline

Current HEAD `127a2e8` on `feature/let-s-see-what-a-local-llm-can-do`. Gate verified green at the end of the auto-expanding memory feature: `go build ./...`, `go test ./...`, `go vet ./...`, `gofmt -l .` (no output). Re-verify before starting.

## What changes, and where

| File | Change |
|---|---|
| `internal/tvmfile/tvm_file.go` (new) | `magic` var; `Decode`, `Encode`, `ReadFile`, `WriteFile` |
| `internal/tvmfile/tvm_file_test.go` (new) | In-memory tests: `TestDecode`, `TestEncode`, `TestEncodeDecodeRoundTrip`, `TestReadFile`, `TestWriteFile` |
| `internal/tvmfile/fixture_test.go` (new) | Fixture-based verification: `TestDecodeFixtures`, `TestEncodeMatchesFixtures` |
| `.gitattributes` (new) | `*.tvm binary` |
| `test_data/` (new; 9 files) | Hand-authored fixtures (table below) |
| `cmd/tvm/main.go` | Refactored into `run(stdin, stdout, stderr, args) int`; reads + executes a `.tvm` file; the hardcoded `[]int{9}` stub dies |
| `cmd/tvm/io.go` (new) | `stdinInput` / `stdoutOutput` adapters implementing the machine's I/O interfaces |
| `cmd/tvm/main_test.go` (new) | `TestRun` table over `run` |
| `cmd/tva/main.go` | Refactored into `run(...)`; reads a `.tva` source, assembles, writes a `.tvm` file; the hello-world stub dies |
| `cmd/tva/main_test.go` (new) | `TestRun` table over `run` |
| `README.md` | Check off TODOs on lines 64 and 95 (File Format section stays — it is already accurate). Note: CRLF line endings in this file. |
| `AGENTS.md` | Rewrite the "file I/O is unimplemented" gotcha |

**What does *not* change:** everything in `internal/virtual_machine` (the auto-expanding-memory feature is done and frozen), the assembler core (`tsvetok_assembler.go`, `instruction_builder.go`, `label_extractor.go`), the register file, `mocks.go`.

## Fixtures (`test_data/`)

| File | Words | Authoring bytes (hex; `54 56 4D 00` is the magic) |
|---|---|---|
| `halt.tvm` | `9` | `54 56 4D 00` + `09 00 00 00` |
| `add_5_5.tvm` | `1101 5 5 10 9` (`add 5, 5, $10` — stores one past the program, exercising auto-expansion) | magic + `4D 04 00 00` `05 00 00 00` `05 00 00 00` `0A 00 00 00` `09 00 00 00` |
| `out_5.tvm` | `104 5 9` (`out 5`) | magic + `68 00 00 00` `05 00 00 00` `09 00 00 00` |
| `in_out.tvm` | `203 0 204 0 9` (`in r0; out r0; hlt`) | magic + `CB 00 00 00` `00 00 00 00` `CC 00 00 00` `00 00 00 00` `09 00 00 00` |
| `empty.tvm` | (none) | `54 56 4D 00` |
| `bad_magic.tvm` | `9` | `58 56 4D 00` + `09 00 00 00` |
| `truncated_word.tvm` | (3 stray bytes) | `54 56 4D 00` + `01 00 00` |
| `out_5.tva` | text: `out 5` / `hlt` | — |
| `bad_instruction.tva` | text: `frobnicate 1` | — |

The `.tvm` bodies are generated with `printf` octal escapes (e.g. `1101 = 0x044D → \115\004\000\000`) and verified against the hex table with `xxd`.

## Non-goals (flagged, not fixed)

- **Executing an empty program still panics** — `getCurrentOperation` reads `t.memory[pc]` directly (the pre-existing PC-past-end issue, flagged in the memory feature). Consequence: `empty.tvm` is decoded/read in tests but **never executed** in the `cmd/tvm` tests.
- **Assembler label resolution** — `cmd/tva` propagates the existing label errors as-is; labels remain a separate feature.
- **Format evolution** — no version byte, no compression. A file may legitimately contain data words, not just instructions; the VM does not care.
- **CLI frills** — no program-from-stdin, globbing, or inspection flags for either executable.

## Steps

### Phase 1 — The `tvmfile` codec (before any fixtures, per decision 5)

Each step is a small, single-purpose commit: red tests first, then the implementation that makes them pass. The package's test file grows incrementally; its import block ends as `("fmt", "math", "os", "path/filepath", "testing", testify assert/require)`.

#### Step 1 — Package skeleton

- [ ] Create `internal/tvmfile/tvm_file.go` with the package doc comment and four stubs (the `magic` var lands with Step 3, not here):

```go
// Package tvmfile reads and writes TVM binary files. A TVM file begins with the four
// ASCII magic bytes `TVM\0`, followed by a sequence of little-endian signed 32-bit words
package tvmfile

import (
	"fmt"
)

// Decode decodes a TVM binary blob into the sequence of words it contains
func Decode(data []byte) ([]int, error) {
	return nil, fmt.Errorf("not implemented")
}

// Encode encodes a sequence of words as a TVM binary blob
func Encode(program []int) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadFile reads and decodes the TVM binary file at the given path
func ReadFile(path string) ([]int, error) {
	return nil, fmt.Errorf("not implemented")
}

// WriteFile encodes the given program and writes it to the given path
func WriteFile(path string, program []int) error {
	return fmt.Errorf("not implemented")
}
```

- [ ] Gate green (`go build ./...` etc.).
- [ ] Commit: `Add tvmfile package skeleton`

#### Step 2 — Decode (RED)

- [ ] Create `internal/tvmfile/tvm_file_test.go` with an in-memory `TestDecode` (no fixtures yet — the blobs are built from literals, so the test's ground truth is independent of `Encode`):

```go
var testMagic = []byte("TVM\0")

func TestDecode(t *testing.T) {
	for _, tc := range []struct {
		name        string
		data        []byte
		wantWords   []int
		wantErrPart string
	}{
		{
			name:        "single word",
			data:        append(append([]byte{}, testMagic...), 9, 0, 0, 0),
			wantWords:   []int{9},
		},
		{
			name:        "bad magic",
			data:        []byte("XVM\0\x09\x00\x00\x00"),
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
			name:        "negative word sign-extends",
			data:        append(append([]byte{}, testMagic...), 0xFF, 0xFF, 0xFF, 0xFF),
			wantWords:   []int{-1},
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
```

- [ ] Confirm red: `go test ./internal/tvmfile/ -run TestDecode -v` — every subtest fails with "not implemented."
- [ ] Commit: `Add tests for decoding TVM binaries`

#### Step 3 — Decode (GREEN)

- [ ] Implement in `tvm_file.go` (add `bytes` and `encoding/binary` imports):

```go
// magic is the four-byte ASCII prefix that identifies a TVM binary file
var magic = []byte("TVM\0")

// Decode decodes a TVM binary blob into the sequence of words it contains. The blob
// must begin with the TVM magic bytes, and the remaining bytes must form whole
// little-endian signed 32-bit words
func Decode(data []byte) ([]int, error) {
	if !bytes.HasPrefix(data, magic) {
		return nil, fmt.Errorf("invalid magic bytes: expected 'TVM\\0'")
	}

	rest := data[len(magic):]
	if len(rest)%4 != 0 {
		return nil, fmt.Errorf("truncated file: %v bytes after the magic do not form whole 4-byte words", len(rest))
	}

	words := make([]int, 0, len(rest)/4)
	for i := 0; i < len(rest); i += 4 {
		words = append(words, int(int32(binary.LittleEndian.Uint32(rest[i:i+4]))))
	}

	return words, nil
}
```

- [ ] Green: `go test ./internal/tvmfile/ -v` + full gate.
- [ ] Commit: `Decode TVM binary files` (body bullets: magic prefix check; multiple-of-4 check; sign-extended little-endian word loop)

#### Step 4 — Encode (RED)

- [ ] Add to `tvm_file_test.go` (the byte literals are the ground truth; the round-trip test is corroborating):

```go
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
			wantBytes: append(append([]byte{}, testMagic...), 0xFF, 0xFF, 0x7F, 0x7F),
		},
		{
			name:      "minimum word",
			program:   []int{math.MinInt32},
			wantBytes: append(append([]byte{}, testMagic...), 0x00, 0x00, 0x80, 0x7F),
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
```

- [ ] Confirm red: `go test ./internal/tvmfile/ -run 'TestEncode|RoundTrip' -v` — all fail with "not implemented."
- [ ] Commit: `Add tests for encoding TVM binaries`

#### Step 5 — Encode (GREEN)

- [ ] Implement in `tvm_file.go` (add the `math` import):

```go
// Encode encodes a sequence of words as a TVM binary blob: the magic bytes followed
// by each word as a little-endian signed 32-bit integer. Words outside the signed
// 32-bit range return an error
func Encode(program []int) ([]byte, error) {
	out := make([]byte, len(magic)+4*len(program))
	copy(out, magic)

	for i, word := range program {
		if word < math.MinInt32 || word > math.MaxInt32 {
			return nil, fmt.Errorf("word %v (%v) does not fit in a signed 32-bit integer", i, word)
		}

		binary.LittleEndian.PutUint32(out[len(magic)+4*i:], uint32(int32(word)))
	}

	return out, nil
}
```

- [ ] Green: `go test ./internal/tvmfile/ -v` + full gate.
- [ ] Commit: `Encode TVM binary files`

#### Step 6 — File I/O (RED)

- [ ] Add to `tvm_file_test.go` (files go to `t.TempDir()`; the bytes written during setup are literals, never `Encode` output — the disk round-trip is the point):

```go
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
			fileBytes:   []byte("XVM\0\x09\x00\x00\x00"),
			wantErrPart: "magic",
		},
		{
			name:         "missing file errors",
			fileName:     "missing.tvm",
			wantErrPart:  "missing.tvm",
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
```

- [ ] Confirm red: `go test ./internal/tvmfile/ -run 'TestReadFile|TestWriteFile' -v`.
- [ ] Commit: `Add tests for reading and writing TVM files`

#### Step 7 — File I/O (GREEN)

- [ ] Implement in `tvm_file.go` (add the `os` import):

```go
// ReadFile reads the TVM binary file at the given path and decodes it
func ReadFile(path string) ([]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	words, err := Decode(data)
	if err != nil {
		return nil, fmt.Errorf("reading %v: %w", path, err)
	}

	return words, nil
}

// WriteFile encodes the given program and writes it to the given path
func WriteFile(path string, program []int) error {
	data, err := Encode(program)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %v: %w", path, err)
	}

	return nil
}
```

- [ ] Green: `go test ./internal/tvmfile/ -v` + full gate.
- [ ] Commit: `Read and write TVM files`

### Phase 2 — Fixtures (ground truth, after the codec)

#### Step 8 — `.gitattributes`

- [ ] Create `.gitattributes` at the repo root with the single line `*.tvm binary` (pins the binary fixtures against CRLF normalization on mixed-platform checkouts).
- [ ] Commit: `Mark TVM binary files as binary for git`

#### Step 9 — The fixture files

- [ ] Generate the seven `.tvm` fixtures (octal-escape equivalents of the hex table above):

```sh
mkdir -p test_data

printf 'TVM\0\011\000\000\000' > test_data/halt.tvm
printf 'TVM\0\115\004\000\000\005\000\000\000\005\000\000\000\012\000\000\000\011\000\000\000' > test_data/add_5_5.tvm
printf 'TVM\0\150\000\000\000\005\000\000\000\011\000\000\000' > test_data/out_5.tvm
printf 'TVM\0\313\000\000\000\000\000\000\000\314\000\000\000\000\000\000\000\011\000\000\000' > test_data/in_out.tvm
printf 'TVM\0' > test_data/empty.tvm
printf 'XVM\0\011\000\000\000' > test_data/bad_magic.tvm
printf 'TVM\0\001\000\000' > test_data/truncated_word.tvm
```

  and the two `.tva` text files: `test_data/out_5.tva` = `out 5` / `hlt` and `test_data/bad_instruction.tva` = `frobnicate 1` (one line each, trailing newline).
- [ ] Verify: `xxd test_data/*.tvm` matches the hex table byte-for-byte.
- [ ] Commit: `Add test fixtures for TVM binary file I/O`

#### Step 10 — Fixture verification tests (green-from-the-start guardrail)

- [ ] Create `internal/tvmfile/fixture_test.go`:

```go
// fixturePath resolves a file from the repository's test_data directory
func fixturePath(name string) string {
	return filepath.Join("..", "test_data", name)
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
```

  These pass immediately — their job is to catch a botched byte in Step 9's `printf` commands (if one fails, the fixture is wrong, not the codec; fix the fixture).
- [ ] Confirm green: `go test ./internal/tvmfile/ -run Fixture -v` + full gate.
- [ ] Commit: `Add fixture verification tests for the TVM file format`

### Phase 3 — `cmd/tvm`: read a TVM file and execute it (README line 64)

#### Step 11 — Refactor into a testable `run`

- [ ] Rewrite `cmd/tvm/main.go` (imports land in two steps: `fmt`/`io`/`os` here; `tvm`/`tvmfile`/`bufio` arrive in Step 13):

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]))
}

// run executes the TVM program file at args[0]. It returns the process exit code:
// 2 for usage errors, 1 for load or execution errors, and 0 for a clean halt
func run(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: tvm <program.tvm>")
		return 2
	}

	return 1 // TODO: load and execute the program file
}
```

- [ ] Gate green. Commit: `Refactor the tvm executable into a testable run function` (note in the body that the hardcoded `[]int{9}` placeholder is removed)

#### Step 12 — `TestRun` (RED)

- [ ] Create `cmd/tvm/main_test.go`:

```go
func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name             string
		args             []string
		stdin            string
		wantExit         int
		wantStdout       string
		wantStderrPart   string
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
```

  (The usage cases pass at this point — they were implemented in Step 11; every file-backed case is red: the stub returns 1 and prints nothing.)
- [ ] Confirm red: `go test ./cmd/tvm/ -run TestRun -v`.
- [ ] Commit: `Add tests for executing TVM files`

#### Step 13 — Execute (GREEN)

- [ ] Complete `run` in `main.go` (add the `bufio`, `tvm/internal/tvmfile`, and `tvm/internal/virtual_machine` imports):

```go
func run(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: tvm <program.tvm>")
		return 2
	}

	program, err := tvmfile.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	machine := tvm.NewTsvetokVirtualMachine(program)
	machine.SetInputInterface(&stdinInput{scanner: bufio.NewScanner(stdin)})
	machine.SetOutputInterface(&stdoutOutput{writer: stdout})

	if err := machine.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}
```

- [ ] Create `cmd/tvm/io.go`:

```go
// stdinInput implements the machine's input interface by reading integers from a
// reader. End-of-file and unparseable input both yield 0, since the interface
// cannot carry an error
type stdinInput struct {
	scanner *bufio.Scanner
}

func (s *stdinInput) ReceiveInput() int {
	if !s.scanner.Scan() {
		return 0
	}

	number, err := strconv.Atoi(s.scanner.Text())
	if err != nil {
		return 0
	}

	return number
}

// stdoutOutput implements the machine's output interface by writing each integer
// to a writer, one per line
type stdoutOutput struct {
	writer io.Writer
}

func (o *stdoutOutput) EmitOutput(number int) {
	fmt.Fprintln(o.writer, number)
}
```

- [ ] Green: `go test ./cmd/tvm/ -v` + full gate.
- [ ] Commit: `Execute TVM files in the tvm executable` (body bullets: `tvmfile.ReadFile` + real stdin/stdout adapters; usage errors exit 2, load/execution errors exit 1)

### Phase 4 — `cmd/tva`: write a TVM binary file (README line 95)

#### Step 14 — Refactor into a testable `run`

- [ ] Rewrite `cmd/tva/main.go` (replacing the hello-world stub):

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]))
}

// run assembles the TVA source file at args[0] and writes the resulting TVM
// binary to args[1]. It returns 2 for usage errors, 1 for read, assembly, or
// write errors, and 0 on success
func run(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: tva <input.tva> <output.tvm>")
		return 2
	}

	return 1 // TODO: assemble and write the program file
}
```

- [ ] Gate green. Commit: `Refactor the tva executable into a testable run function`

#### Step 15 — `TestRun` (RED)

- [ ] Create `cmd/tva/main_test.go` (inputs are the committed fixtures used in place; outputs go to `t.TempDir()`):

```go
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
			wantExit:       1,
			wantStderrPart: "missing.tva",
		},
		{
			name:             "unknown instruction fails assembly",
			inputFixture:     "../../test_data/bad_instruction.tva",
			outputRelPath:    "out.tvm",
			wantExit:         1,
			wantStderrPart:   "unknown instruction",
		},
		{
			name:              "assembles and writes the binary file",
			inputFixture:      "../../test_data/out_5.tva",
			outputRelPath:     "out.tvm",
			wantExit:          0,
			wantOutputFixture: "../../test_data/out_5.tvm",
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
```

  The `assembles and writes` case is the seam anchor: the hand-authored `out_5.tva` must assemble to bytes identical to the hand-authored `out_5.tvm`.
- [ ] Confirm red: `go test ./cmd/tva/ -run TestRun -v`.
- [ ] Commit: `Add tests for writing TVM files`

#### Step 16 — Assemble and write (GREEN)

- [ ] Complete `run` in `main.go` (add the `tvm/internal/tvmfile` import):

```go
func run(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: tva <input.tva> <output.tvm>")
		return 2
	}

	source, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	program, err := NewAssemblerFromString(string(source)).Assemble()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if err := tvmfile.WriteFile(args[1], program); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}
```

- [ ] Green: `go test ./cmd/tva/ -v` + full gate.
- [ ] Commit: `Write TVM binary files in the tva executable`

### Phase 5 — Docs

#### Step 17 — README + AGENTS.md

- [ ] README: check off line 64 (`- [ ]` → `- [x]` *Read a TVM binary file and executes it*) and line 95 (`- [ ]` → `- [x]` *Writes to a TVM binary file with correct syntax*). Keep both sub-phrasings as-is. (CRLF file — stage surgically if the working tree ever has unrelated README edits.)
- [ ] AGENTS.md: rewrite the gotcha

  > Planned TVM binary format (README): little-endian 32-bit words prefixed with the ASCII magic `TVM\0`. File I/O is unimplemented on both the assembler and VM sides.

  to

  > TVM binary files: `internal/tvmfile` encodes/decodes them — ASCII magic `TVM\0` followed by little-endian signed 32-bit words. `cmd/tvm` runs `.tvm` files and `cmd/tva` writes them. Binary test fixtures live in `test_data/` (read-only ground truth; test output goes to temp dirs).

- [ ] Commit: `Update README and AGENTS.md for binary file I/O`

## Final gate

After Step 17, and one last time at the end:

```sh
go build ./...
go test -race ./...
go vet ./...
gofmt -l .   # must print nothing
```

Plus a manual smoke from the repo root:

```sh
go run ./cmd/tva test_data/out_5.tva /tmp/demo.tvm
go run ./cmd/tvm /tmp/demo.tvm   # must print 5
```

## Walkthrough

Work through Steps 1 → 17 in order, checking off each box as its commit lands. The doc is an interim tracking file: once every step has landed, remove it in a final commit (per the convention set by the auto-expanding memory plan).

The shape of the feature: a self-contained codec package (Phase 1) whose behavior is anchored first by literal bytes and then by the committed fixtures (Phase 2), with each executable a thin `run()` over it (Phases 3–4). If the fixture set (decision 6) proves awkward as we go, adjust it in Step 9 before the Phase 3 tests start depending on it.
