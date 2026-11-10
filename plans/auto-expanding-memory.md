# Auto-Expanding Memory

**Status:** Done (all steps committed)
**Source:** README TODO — "Any memory address that does not exist will immediately exist upon lookup or writing. If we expand memory to fill the space, we set everything inside to 0."

## Spec

Memory addresses exist, infinite and zero-valued. Specifically:

- **Writes expand.** `SetValueInMemory(100, 42)` on a 5-word program succeeds; memory grows to 101 words; indices 5–99 are `0`; index 100 is `42`.
- **Reads expand.** `GetValueInMemory(100)` on a 5-word program returns `(0, nil)` and grows memory identically. (Decided: matches the README TODO wording and the classic AoC Day 2 convention — one consistent rule for both accessors.)
- **Negative addresses still error.** Expansion only grows *upward*; there is no negative memory. This behavior is unchanged and must be pinned by tests first.
- **Expansion is exact.** After an access to address `n` where `n >= len(memory)`, `len(memory) == n + 1`. No over-allocation — keeps the behavior simple and fully testable. (Amortized growth is a performance non-goal for an educational VM.)
- **Everything newly created is `0`.** Guaranteed by construction (`make([]int, n)` zero-fills).

## Baseline

Verified green before starting (commit `b51bfca`): `go build ./...`, `go test ./...`, `go vet ./...`, `gofmt -l .` (no output).

## What changes, and where

| File | Change |
|---|---|
| `internal/virtual_machine/tsvetok_virtual_machine.go` | New unexported helper `expandMemory(address int)`; rework `GetValueInMemory` and `SetValueInMemory` to reject only negative addresses, then `expandMemory` before accessing. Error messages reworded for the remaining negative-only case (no test asserts the old strings — verified). |
| `internal/virtual_machine/tsvetok_virtual_machine_test.go` | All new tests (Steps 0–5). |
| `cmd/tva/tsvetok_assembler_test.go` | One seam test (Step 5b). |
| `AGENTS.md` | Rewrite the stale gotcha "memory does not auto-expand (open README TODO)" to describe the new semantics. |
| `README.md` | Check off the TODO on line 63. |

**What does *not* change:** all seven operation files, `operation_param.go`, `cmd/tvm`, `mocks.go`, the register file (fixed at 14 — `operation_param.go` explicitly says it cannot be expanded), and `CopyMemory()` (it naturally reflects growth). Every operation already funnels memory access through the two accessor methods — parameter reads via `newOperationParam → GetValueInMemory`, result writes via `SetValueInMemory` — so centralizing expansion in those two methods gives every operation the feature for free. Verified: no existing test (VM or assembler) depends on the old out-of-bounds error behavior, so nothing regresses.

## Non-goals (flagged, not fixed)

- **Program counter past end of memory** still panics in `getCurrentOperation` (it reads `t.memory[pc]` directly). A malformed-program condition, distinct from data-address expansion. Pre-existing; out of scope.
- **Over-allocation / amortized growth.** We expand exactly to `address+1`.
- **Register file expansion.** Stays fixed at 14.
- **File I/O.** The TVM binary format TODO is separate.

## Steps

Each step is a small, single-purpose commit matching the repo's existing commit granularity (tests committed separately from the implementation that makes them pass, red before green).

### Step 0 — Characterize the invariant that must survive

The rework must not let negative addresses through. Pin the current behavior first (this test is green immediately — it's a guardrail, not red-first).

- [x] Add to `internal/virtual_machine/tsvetok_virtual_machine_test.go`:

```go
func TestTsvetokVirtualMachine_NegativeMemoryAddressesError(t *testing.T) {
	for _, tc := range []executionTestCase{
		{program: []int{9, 1, 2, 3, 4}, testName: "non-empty program"},
		{program: []int{}, testName: "empty program"},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			machine := NewTsvetokVirtualMachine(tc.program)

			err := machine.SetValueInMemory(-1, 42)
			require.Error(t, err, "writing to a negative address must error")

			_, err = machine.GetValueInMemory(-1)
			require.Error(t, err, "reading a negative address must error")
		})
	}
}
```

- [x] Commit: `Add tests for negative memory addresses`

### Step 1 — Writes expand (RED)

- [x] Add:

```go
func TestTsvetokVirtualMachine_ExpandsMemoryOnOutOfBoundsWrite(t *testing.T) {
	machine := NewTsvetokVirtualMachine([]int{9, 1, 2, 3, 4})

	require.NoError(t, machine.SetValueInMemory(100, 42))

	memory := machine.CopyMemory()
	assert.Len(t, memory, 101)
	assert.Equal(t, 42, memory[100])

	// the gap must be zero-filled
	gapVal, err := machine.GetValueInMemory(5)
	require.NoError(t, err)
	assert.Equal(t, 0, gapVal)

	gapVal, err = machine.GetValueInMemory(99)
	require.NoError(t, err)
	assert.Equal(t, 0, gapVal)

	// the original program words must be untouched
	assert.Equal(t, []int{9, 1, 2, 3, 4}, memory[:5])
}

func TestTsvetokVirtualMachine_WritesToExistingMemoryDoNotGrowIt(t *testing.T) {
	machine := NewTsvetokVirtualMachine([]int{9, 1, 2, 3, 4})

	require.NoError(t, machine.SetValueInMemory(4, 99))
	assert.Len(t, machine.CopyMemory(), 5, "writing to the last index must not grow memory")

	require.NoError(t, machine.SetValueInMemory(5, 10))
	memory := machine.CopyMemory()
	assert.Len(t, memory, 6, "writing one past the end must grow by exactly one word")
	assert.Equal(t, 10, memory[5])
}
```

- [x] Confirm red: `go test ./internal/virtual_machine/ -run 'ExpandsMemoryOnOutOfBoundsWrite|WritesToExistingMemoryDoNotGrowIt' -v` fails (the OOB write currently returns an error).
- [x] Commit: `Add tests for out-of-bounds memory writes`

### Step 2 — Writes expand (GREEN)

- [x] In `internal/virtual_machine/tsvetok_virtual_machine.go`, add the helper and rework the write path:

```go
// expandMemory grows t.memory so the given address is in bounds, zero-filling all newly
// created entries. It is a no-op if the address already exists. It does not validate the
// address; callers are responsible for rejecting negative addresses
func (t *TsvetokVirtualMachine) expandMemory(address int) {
	if address < len(t.memory) {
		return
	}

	expanded := make([]int, address+1)
	copy(expanded, t.memory)
	t.memory = expanded
}

func (t *TsvetokVirtualMachine) SetValueInMemory(address, value int) error {
	if address < 0 {
		return fmt.Errorf("cannot write to memory at negative address '%v'", address)
	}

	t.expandMemory(address)
	t.memory[address] = value
	return nil
}
```

  (The error message is reworded from "…(memory is of size …)" because the size-based failure no longer exists. No test asserts the old text.)

- [x] Green: `go test ./internal/virtual_machine/ -v` — the Step 1 tests pass; the Step 0 guardrail still passes; full `go test ./...` green.
- [x] Commit: `Expand memory on out-of-bounds write`

### Step 3 — Reads expand (RED)

- [x] Add:

```go
func TestTsvetokVirtualMachine_ExpandsMemoryOnOutOfBoundsRead(t *testing.T) {
	machine := NewTsvetokVirtualMachine([]int{9, 1, 2, 3, 4})

	// an in-range read is unchanged and must not grow memory
	val, err := machine.GetValueInMemory(3)
	require.NoError(t, err)
	assert.Equal(t, 3, val)
	assert.Len(t, machine.CopyMemory(), 5)

	// an out-of-bounds read returns 0 and materializes the address
	val, err = machine.GetValueInMemory(100)
	require.NoError(t, err)
	assert.Equal(t, 0, val)
	assert.Len(t, machine.CopyMemory(), 101)
}
```

- [x] Confirm red: `go test ./internal/virtual_machine/ -run ExpandsMemoryOnOutOfBoundsRead -v` fails (the OOB read currently returns an error).
- [x] Commit: `Add tests for out-of-bounds memory reads`

### Step 4 — Reads expand (GREEN)

- [x] Rework the read path in `internal/virtual_machine/tsvetok_virtual_machine.go`:

```go
func (t *TsvetokVirtualMachine) GetValueInMemory(address int) (int, error) {
	if address < 0 {
		return -1, fmt.Errorf("cannot lookup memory at negative address '%v'", address)
	}

	t.expandMemory(address)
	return t.memory[address], nil
}
```

- [x] Green: full `go test ./...` + `go vet ./...` + `gofmt -l .` clean.
- [x] Commit: `Expand memory on out-of-bounds read`

### Step 5 — Your exact scenario, end-to-end (GREEN — anchors the feature)

- [x] Add to `internal/virtual_machine/tsvetok_virtual_machine_test.go`:

```go
func TestTsvetokVirtualMachine_StoresToMemoryFarBeyondTheProgram(t *testing.T) {
	// add 5, 5, $100 / hlt — a five-word program that stores at index 100
	machine := NewTsvetokVirtualMachine([]int{1101, 5, 5, 100, 9})
	require.NoError(t, machine.Execute())

	memory := machine.CopyMemory()
	assert.Len(t, memory, 101)
	assert.Equal(t, 10, memory[100])
	assert.Equal(t, 1101, memory[0], "the program words must be untouched")

	gapVal, err := machine.GetValueInMemory(50)
	require.NoError(t, err)
	assert.Equal(t, 0, gapVal, "the gap must be zero-filled")
}
```

- [x] Commit: `Add end-to-end test for storing far beyond the program`

#### Step 5b — Assembler seam

Covers assembler → VM integration. Verified against `instruction_builder.go`: bare digits are immediates, `$` is address mode, so this encodes to `[1101, 5, 5, 100, 9]`.

- [x] Add to `cmd/tva/tsvetok_assembler_test.go`:

```go
func TestTsvetokAssembler_StoringFarBeyondTheProgramExpandsMemory(t *testing.T) {
	assembler := NewAssemblerFromString("add 5, 5, $100\nhlt")
	program, err := assembler.Assemble()
	require.NoError(t, err)
	assert.Equal(t, []int{1101, 5, 5, 100, 9}, program)

	machine := tvm.NewTsvetokVirtualMachine(program)
	require.NoError(t, machine.Execute())

	memory := machine.CopyMemory()
	assert.Len(t, memory, 101)
	assert.Equal(t, 10, memory[100])
}
```

- [x] Commit: `Add assembler test for storing far beyond the program`

### Step 6 — Docs

- [x] Rewrite the AGENTS.md gotcha. Current:

  > VM memory is exactly the `[]int` passed to `NewTsvetokVirtualMachine`. Out-of-bounds read/write returns an error; memory does not auto-expand (open README TODO).

  New:

  > VM memory starts as the `[]int` passed to `NewTsvetokVirtualMachine`. Any read or write to a non-existent (non-negative) address immediately creates it, zero-filled; the memory slice grows exactly to `address+1`. Negative addresses return an error. The register file does *not* expand.

- [x] Check off the README TODO (line 63): `- [ ]` → `- [x]` (keep the sub-bullet; it now describes implemented behavior).
- [x] Commit: `Update README and AGENTS.md for auto-expanding memory`

## Final gate

After each step (and one last time at the end):

```sh
go build ./...
go test ./...
go vet ./...
gofmt -l .   # must print nothing
```

## Walkthrough

Work through Steps 0 → 6 in order, checking off each box as its commit lands. The implementation delta on `tsvetok_virtual_machine.go` is one ~10-line helper plus a 3-line rework of each accessor — the rest of the feature arrives for free through the two choke points.
