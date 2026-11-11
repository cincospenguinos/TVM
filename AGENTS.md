# AGENTS.md

Tsvetok Virtual Machine (TVM): an educational Go project written for the owner's TDD book. Single Go module named `tvm` (no workspace); the only external dependency is `stretchr/testify`.

## Commands

- Build: `go build ./...`
- Test: `go test ./...`
- Single test: `go test ./internal/virtual_machine/ -run TestName -v`
- There is no CI or linter config. `go vet ./...` and `gofmt -l .` are the quality bar — keep both clean before finishing.

## Layout

- `internal/virtual_machine` — the VM: `TsvetokVirtualMachine`, one file per operation, `tvm_operation.go` (the `TVMOperation` interface + halt), `mocks.go` (I/O mocks).
- `cmd/tva` — the assembler: `TsvetokAssembler`, `instructionBuilder`, `labelExtractor`.
- `cmd/tvm` — the VM executable.
- Tsvetalk (a higher-level language compiling to TVA) is planned; there is no code for it.

## Gotchas

- Import the VM as `tvm/internal/virtual_machine` (module name is `tvm`), not the GitHub path.
- `cmd/tva`'s main assembles a `.tva` source file (one arg) into a `.tvm` binary at the output path (two args) via `NewAssemblerFromString(src).Assemble() → ([]int, error)`. The assembler lives in the main package, so it can't be imported from other packages.
- `cmd/tvm`'s main runs one `.tvm` file (one arg) with stdin/stdout wired up as the machine's I/O (`io.go`); stdin that is empty or unparseable yields 0.
- Both mains return 2 for usage errors and 1 for any load/assembly/execution/write failure.
- VM memory starts as the `[]int` passed to `NewTsvetokVirtualMachine`. Any read or write to a non-existent (non-negative) address immediately creates it, zero-filled; the memory slice grows exactly to `address+1`. Negative addresses return an error. The register file does *not* expand.
- `in`/`out` return an error if the machine has no input/output interface set. Tests should set `MockInputInterface` / `MockOutputInterface` from `mocks.go` first.
- Intcode is decimal-digit encoded: the lowest two digits are the opcode, and each higher digit is a parameter format — 0=address, 1=immediate, 2=register.
- The register file is fixed at 14 registers: r0–r4 (reserved, preserved across jumps), t0–t7 (temporaries), `la`=13 (write-protected; only `jit` writes it, and only when the jump is taken). Writing it returns `AttemptedLastAddressWriteErr`.
- Labels: the assembler extracts declarations/accessions (`label_extractor.go`) but never resolves them — assembling any label line fails with a strconv error. The one `t.Skip`ed test covers this; don't unskip it.
- TVM binary files: `internal/tvmfile` encodes/decodes them — ASCII magic `TVM\0` followed by little-endian signed 32-bit words. Binary test fixtures live in `test_data/` (read-only ground truth; test output goes to temp dirs).

## Conventions

- Tests are table-driven (`executionTestCase`) using testify `assert`/`require`. Assembler tests assemble a string, execute it on the real VM, and assert on memory/output.
- The owner commits in small, single-purpose steps (see git log) — match that granularity.
- The repo is GPL-3.0 licensed.
