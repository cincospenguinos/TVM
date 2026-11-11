# Plan: Labels and `jmp` in the TVA Assembler

Status: ready for owner review (2026-11-11). Owner guidance from the planning session is incorporated below.

## Goal

Let TVA source use labels: declare one on its own line, and jump to "the instruction on the line immediately following the label" via a jump instruction or pseudo-instruction. The sample program `test_data/label_jump_test.tva` must assemble and run as its comments describe: `jmp i.main` skips the `add`, so `out $0` prints the un-overwritten `jmp` word.

## Owner guidance (2026-11-11)

1. `jmp` is a pseudo-instruction that expands to `jit 1, <target>` (constant-true condition). Include it in the implementation sequence.
2. `$.<label>` references must be **accepted** by the assembler without any semantic assertion — the assembler does not judge whether the resulting program makes sense. (A runtime demo of `$.<label>` is inherently contrived, since the word after a label is an opcode word; test it at the assembler level only.)
3. Declaring data at a label (e.g. `.my-data 69 420`, eventually ASCII strings) is a future feature and well out of scope. In this round, a label line with anything after the label is a **clear assembly error**.
4. Register-form label references (`r.<label>` / `t.<label>`) are **not** a feature of this language. Labels are referenced only via `i.` and `$.`.
5. The legacy `t.Skip`ed label test is unskipped as the **final** step of the implementation, as proof the feature works; the test is tightened accordingly and AGENTS.md is updated to reflect the change.

## Current state (analysis, 2026-11-11)

| Piece | State |
|---|---|
| `cmd/tva/label_extractor.go` | Half-built. Any line starting with `.` is treated as a declaration, and the name is taken with `strings.ReplaceAll(line, ".", "")` — no name validation, so `.bad label` silently "declares" `bad label`. Accession scan is unanchored (`[\$i]\.` matches anywhere in a chunk, so `ti.foo` registers) and never checks that referenced labels were declared. Contains a `line, _ :=` shadowing bug. |
| `cmd/tva/tsvetok_assembler.go` | Two loops: extract labels, then build instructions. A label line falls through to `setOperation` and dies with `unknown instruction '.main'`; a label param dies in `addParam` with a strconv error (`strings.ReplaceAll(param, "i", "")` mangles the `i` inside `main` too). Matches the AGENTS.md gotcha ("assembling any label line fails with a strconv error"). |
| `cmd/tva/instruction_builder.go` | `setOperation` knows `add mlt in out seq jit hlt` — **no `jmp`** (assembling the sample fails with `unknown instruction 'jmp'`, verified). |
| `internal/virtual_machine` | **No changes needed.** `jit` (opcode 6) already does everything a jump needs: condition = param 0, target = param 1, sets `la` when taken. Immediate/address parameter formats already implement the `i.` / `$.` semantics. |
| `TestTsvetokAssembler_AcceptsLabelsAsLocationsForJumping` | `t.Skip`ed; AGENTS.md says "don't unskip it". It covers exactly this feature and will pass once labels work. Unskipped + tightened in the final step per owner guidance #5. |
| `README.md` | TVA section already claims "Labels are supported" and has an open checkbox `[ ] Labels for jumping are supported` (plus a separate checkbox for data preservation, which stays open). |
| `test_data/label_jump_test.tva` | Added by the owner, currently untracked. Committed alongside the step that first references it. |

Baseline is green: `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` all clean.

## Design

### Label syntax

A valid label is a `.` followed by one or more of `[A-Za-z0-9_-]` (letters, digits, dashes, underscores; no dots or other punctuation). This covers all of the owner's valid examples (`.label`, `.data`, `.1234`, `._1234`, `.my-cool-label`, `.my_cool_label`, `.__foo-bar123__`) and rejects the invalid ones (`.bad label`, `.$foo`). Two regexes, both living in `label_extractor.go` as the single source of truth:

- declaration line (the whole trimmed line): `^\.[A-Za-z0-9_-]+$`
- parameter reference (after comma-strip): `^([i$])\.[A-Za-z0-9_-]+$` — capture group 1 is the format indicator

### Semantics

- **Declaration**: a label must occupy its **entire line**. Its value is the **current output word index** at the point of declaration — for a standalone line that is exactly the index of the first word of the next instruction, i.e. "the index of the instruction immediately following the label".
- **Reference** (in any parameter slot): both forms resolve to the same base index; only the parameter format differs:
  - `i.<label>` → **immediate** format, value = label index (use the index directly)
  - `$.<label>` → **address** format, value = label index (the VM reads the *contents* at that index)
- **Forward references work**: `jmp i.main` before `.main` is fine, because resolution happens after the whole file is scanned.
- `la` behaves as it does for `jit`: set to the instruction after the jump when the jump is taken (always, since `jmp`'s condition is constant-true).

### Assembly pipeline (classic two-pass, extended to three loops)

1. **Extract** (existing loop): `labelExtractor.ExtractLabels(line)` per line — validates declaration syntax + uniqueness, validates reference syntax, records declarations and accessions with their lines.
2. **Validate**: new `labelExtractor.ValidateAccessions()` — every referenced label must have been declared anywhere in the file.
3. **Index pass** (new): walk lines in order tracking `currentIndex`. A declaration line registers `labelIndexes[name] = currentIndex` and adds zero words. An instruction line adds its word count: `len(chunks)` + parameters injected by pseudo-ops — `jmp` injects one (its constant-true condition), so `jmp <target>` contributes `len(chunks) + 1` words. A small `instructionWords(chunks)` helper holds this; it is the only place that knows about pseudo-op expansion cost, and it extends trivially when `jif` et al. land.
4. **Build** (existing loop): declaration lines are skipped. For each parameter: strip commas, and if the param is a label reference, **rewrite it textually** to a plain parameter using the resolved index (`i.main` → `i7`, `$.main` → `$7`), then call the existing `addParam`. The `instructionBuilder` therefore needs **no label awareness at all** — it only ever sees ordinary immediates/addresses (the rewrite keeps a valid label reference pure-digit, so the existing `ReplaceAll(param, "i", "")` handling is safe).

The build pass defensively errors if it ever encounters a reference whose name failed to resolve (cannot happen once the extract+validate passes run, but it guards against silent mis-assembly).

### `jmp` pseudo-instruction

The VM has no unconditional-jump opcode, so the assembler expands `jmp <target>` → `jit 1 <target>`:

- `instructionBuilder` gains a `jmp` case in `setOperation`: opcode `6`, param 0 pre-populated with immediate `1`, and a slot offset so the user's single target lands in jit's slot 1.
- `jmp` accepts any target form: `jmp 7`, `jmp $0`, `jmp r0`, `jmp i.<label>`, `$.`-form, and accepts **exactly one** parameter (a second target parameter is rejected — jit would only read slot 1 and the extra word would corrupt the next-instruction pointer).
- `{"." + "label", false}` in the builder unit tests still passes: a `.label` handed directly to `setOperation` remains an unknown instruction; label *lines* are intercepted before reaching the builder.

### Error policy

| Condition | Behavior |
|---|---|
| invalid declaration syntax (`.bad label`, `.$foo`, `.`) | error, line number included. `.bad label` (valid label + extra content) gets "must be alone on its line"; bad characters get "invalid label" |
| invalid reference syntax (`i.$foo`) | error, line number included |
| reference to a never-declared label | error naming the label and the referencing line (first occurrence) |
| duplicate declaration | error (existing behavior, message kept) |
| declared but unused label | silently allowed |
| `$.<label>` pointing somewhere semantically odd | **accepted** — the assembler makes no semantic assertions (owner guidance #2) |

## Expected behavior of the sample

`test_data/label_jump_test.tva` assembles to:

```
idx  0     1   2     3      4   5   6    7       8   9
     1106  1   7   11101   10  10  0   10004   0   9
      jmp i.main           add 10,10,$0   out $0    hlt    (.main = index 7)
```

Execution: `jit` at 0 takes (cond 1) → PC 7, `la` 3 → `out $0` prints `memory[0]` = **1106** → `hlt`. The `add` never executes, so `memory[0]` stays 1106. Expected end-to-end: stdout `1106\n`, exit 0, `memory[0] == 1106`.

New fixture `test_data/label_jump_test.tvm`: the TVM encoding of the intcode above (magic `TVM\0` + 10 little-endian signed 32-bit words, 44 bytes; `*.tvm binary` already in `.gitattributes`). Generated via `tvmfile.WriteFile`.

## Code changes by file

- **`cmd/tva/label_extractor.go`** — rewrite around the two regexes: declaration validation (name = `line[1:]`, shadowing bug fixed), anchored reference parsing via a shared `parseLabelReference` helper, first-occurrence accession records, new `ValidateAccessions()`, updated doc comment (drop the stale `.my-variable 0` same-line example).
- **`cmd/tva/tsvetok_assembler.go`** — `Assemble()` becomes: lines → extract → `ValidateAccessions` → index pass (new `instructionWords` helper) → build pass (skip declarations; rewrite references to `<indicator><index>` before `addParam`) → concat.
- **`cmd/tva/instruction_builder.go`** — add the `jmp` case (slot offset, pre-set condition, one-param limit). No label handling.
- **`internal/virtual_machine/**`** — untouched.
- **Fixtures** — add `test_data/label_jump_test.tvm`; add a small `test_data/bad_label.tva` (`.bad label` + `hlt`) for the `run()` error-path case, mirroring the existing `bad_instruction.tva` precedent.
- **Tests** — `label_extractor_test.go` (valid/invalid names, invalid refs, undeclared refs), `instruction_builder_test.go` (jmp intcode + one-param limit), `tsvetok_assembler_test.go` (resolution intcode cases, run-on-VM fixture test, legacy test unskipped+tightened last), `cmd/tva/main_test.go` (fixture round-trip, bad-label error), `cmd/tvm/main_test.go` (run the new `.tvm` fixture → `1106\n`).
- **`AGENTS.md`** — replace the stale label gotcha with the real semantics + `jmp`; the "don't unskip" note is removed (final step).
- **`README.md`** — check off `Labels for jumping`, add `jmp` to the TVA TODO list, specify the `i.` / `$.` reference syntax in the features bullet (final step).

## Implementation steps

Test-first; each step is one commit, red tests first then green. (Commit messages per the git-commit skill.)

1. **Validate label declaration syntax** — tests: extractor accepts every valid name form, rejects `.bad label` / `.$foo` / `.` with errors, still rejects duplicates. Impl: `labelDeclaration` regex + `ExtractLabels` declaration branch rewrite.
2. **Validate label references** — tests: invalid reference syntax errors; `Assemble` errors on a referenced-but-undeclared label (with line number); a forward reference (use before declaration) passes extraction. Impl: anchored `parseLabelReference` in the accession scan + `ValidateAccessions()` called in `Assemble`.
3. **`jmp` pseudo-instruction** — tests: `setOperation("jmp")` valid (`.label` still invalid); `jmp 7` → `[1106, 1, 7]`, `jmp $0` → `[106, 1, 0]`, `jmp r0` → `[2106, 1, 0]`; a second `jmp` param is rejected. Impl: builder `jmp` case with slot offset and one-param limit.
4. **Resolve label references in the assembler** — tests (assemble → assert exact intcode): standalone label emits no words; backward ref (`add… / .end / jit 1, i.end / hlt` → `[11101, 1, 1, 0, 1106, 1, 4, 9]`); forward ref (same order reversed → index 7); `$.end` → address format (`[106, 1, 7, …]`); index accumulation across mixed word counts; the sample program's exact intcode `[1106, 1, 7, 11101, 10, 10, 0, 10004, 0, 9]`. Impl: skip declarations in the build loop, index pass with `instructionWords` (jmp-aware), textual reference rewrite.
5. **End-to-end fixtures and run tests** — add `test_data/label_jump_test.tvm` (and commit the owner's untracked `label_jump_test.tva`) + `bad_label.tva`; tva `run()` cases: sample → byte-for-byte fixture match, bad label → exit 1 with "invalid label" on stderr; tvm `run()` case: sample fixture → stdout `1106\n`, exit 0; assembler test: read the real `.tva`, assemble, run on the VM with mocks, assert output 1106 and `memory[0] == 1106`.
6. **Final: unskip the legacy label test** — unskip `TestTsvetokAssembler_AcceptsLabelsAsLocationsForJumping`, tighten its assertion from `NotEqual(0, memory[0])` to `Equal(1106, memory[0])` (and assert the assembled intcode), per owner guidance #5. Same commit updates **AGENTS.md** (label gotcha → real semantics + `jmp`; remove "don't unskip") and **README.md** (check off `Labels for jumping`, add `jmp`, specify `i.` / `$.` syntax).

Verification gate after each step: `go build ./... && go vet ./... && gofmt -l . && go test ./...` all clean.

## Out of scope (future work)

- Declaring data at a label (`.my-data 69 420`, eventually ASCII) — tracked as its own README TODO ("Labels for data preservation"). Until it lands, same-line label + content is a clear error.
- Register-form label references — not a feature of this language (owner decision).
- `jif`, `sub`, `mov`, `nil` pseudo-instructions and `call`/linking — separate README TODOs; only constraint here is that `instructionWords` extends to them.
- Pre-existing latent bug (flagged, not fixed): any op given more params than it reads (e.g. `jit a, b, c`) emits words the VM skips over, corrupting the next instruction. `jmp`'s one-param limit avoids introducing new instances; a general fix belongs in its own step.
