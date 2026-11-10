# AGENTS.md

Tsvetok Virtual Machine (TVM): an educational Go project written for the owner's TDD book. Single Go module named `tvm` (no workspace); the only external dependency is `stretchr/testify`.

## Commands

- Build: `go build ./...`
- Test: `go test ./...` (`cmd/tvm` has no tests)
- Single test: `go test ./internal/virtual_machine/ -run TestName -v`
- There is no CI or linter config. `go vet ./...` and `gofmt -l .` are the quality bar — keep both clean before finishing.

## Layout

- `internal/virtual_machine` — the VM: `TsvetokVirtualMachine`, one file per operation, `tvm_operation.go` (the `TVMOperation` interface + halt), `mocks.go` (I/O mocks).
- `cmd/tva` — the assembler: `TsvetokAssembler`, `instructionBuilder`, `labelExtractor`.
- `cmd/tvm` — the VM executable.
- Tsvetalk (a higher-level language compiling to TVA) is planned; there is no code for it.

## Gotchas

- Import the VM as `tvm/internal/virtual_machine` (module name is `tvm`), not the GitHub path.
- `cmd/tva`'s main is a stub printing hello world; the assembler is consumed as a library: `NewAssemblerFromString(src).Assemble() → ([]int, error)`.
- `cmd/tvm/main.go` is a stub that hardcodes `[]int{9}` (halt); it does not read a TVM file.
- VM memory starts as the `[]int` passed to `NewTsvetokVirtualMachine`. Any read or write to a non-existent (non-negative) address immediately creates it, zero-filled; the memory slice grows exactly to `address+1`. Negative addresses return an error. The register file does *not* expand.
- `in`/`out` return an error if the machine has no input/output interface set. Tests should set `MockInputInterface` / `MockOutputInterface` from `mocks.go` first.
- Intcode is decimal-digit encoded: the lowest two digits are the opcode, and each higher digit is a parameter format — 0=address, 1=immediate, 2=register.
- The register file is fixed at 14 registers: r0–r4 (reserved, preserved across jumps), t0–t7 (temporaries), `la`=13 (write-protected; only `jit` writes it, and only when the jump is taken). Writing it returns `AttemptedLastAddressWriteErr`.
- Labels: the assembler extracts declarations/accessions (`label_extractor.go`) but never resolves them — assembling any label line fails with a strconv error. The one `t.Skip`ed test covers this; don't unskip it.
- Planned TVM binary format (README): little-endian 32-bit words prefixed with the ASCII magic `TVM\0`. File I/O is unimplemented on both the assembler and VM sides.

## Conventions

- Tests are table-driven (`executionTestCase`) using testify `assert`/`require`. Assembler tests assemble a string, execute it on the real VM, and assert on memory/output.
- The owner commits in small, single-purpose steps (see git log) — match that granularity.
- The repo is GPL-3.0 licensed.
