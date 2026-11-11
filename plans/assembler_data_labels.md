# Plan: Data Preservation Labels in the TVA Assembler

Status: approved by the owner 2026-11-11, with the three rulings below incorporated. Ready for implementation on branch `feature/implement-data-labels`.

## Goal

Let a label declaration line also carry data values: `.data 12 34 i.other $.copy`. The label is declared, indexed, and validated exactly as a regular label, and each value token becomes one word in the program's word stream at the label's index. Data words are respected everywhere labels are respected, because they are ordinary words — the VM and the file format are untouched. Closes the README TODO "Labels for data preservation are supported", which `plans/assembler_labels.md` deferred as future work.

## Owner guidance (2026-11-11)

1. `.data i12` is equivalent to `.data 12`. In the TVA instruction-set agreement, a literal integer value is shorthand for a prepended `i` indicating literal or "immediate" (as in `add i12, i12, $0`); the `i` prefix must therefore be accepted in any spot where an immediate may be declared, including data values.
2. Data values may be **negative**: signed integers are delineated by the `-` character (`.data -1` places −1 at the spot indicated by the label).
3. The sample fixture (`.tva` + generated `.tvm` + round-trip and run tests) is a **required final step**, placed after the rest of the implementation, demonstrating the feature's function end-to-end.

The original spec from the planning session, all of which this plan honors: `.data 12`, `.data-12 12`, and `.data 1 2 3` are accepted (name = first token; values follow); `.data foo` (bare word) is **not** accepted; `.data $.foo` **is** accepted, storing the value found at the address of label `foo` as the word at `data`'s index (finicky, as the owner noted: if `foo` precedes an instruction, that value is the instruction's encoded opcode word); `.data i.foo` stores the label's index; and arbitrary mixes of the forms, e.g. `.crazy-data__69420 i.bar 12 3 $.biz i.baz`, emit each value as an individual word at that point in the program.

## Current state (analysis, 2026-11-11)

| Piece | State |
|---|---|
| `cmd/tva/label_extractor.go` | Complete for regular labels: whole-line declarations only (`labelDeclarationPattern` matches `^\.[A-Za-z0-9_-]+$`); any content after a label is a clear error (per the prior plan's guidance #3). Data values must land here as a second declaration form. |
| `cmd/tva/label_extractor_test.go` | Three rows of `TestLabelExtractor_RejectsInvalidLabelDeclarations` (`.bad label`, `.foo bar baz`, `.foo bar`) expect "must be alone on its line" — those exact shapes become *data-label attempts* under the new grammar, so the expected wording must change. The `.foo.`, `.`, `.$foo` rows are untouched. |
| `cmd/tva/tsvetok_assembler.go` | Three passes: extract → validate → index (`instructionWords`, instruction-only by design) → build (skips whole-line declarations; rewrites `i.`/`$.` references textually before `addParam`). Data labels need branches in the index and build passes plus a new resolution step. |
| `cmd/tva/instruction_builder.go` | `addParam` is the reference for immediate semantics: `i12` and `i-12` parse; a bare `-12` is rejected ("unknown parameter format"); `-i12` parses only by virtue of an unanchored `ReplaceAll` quirk (verified). Data values follow the same agreement with anchored regexes (see value tokens). |
| `internal/virtual_machine/**`, `internal/tvmfile` | **No changes needed.** Data words are ordinary words in memory; the VM does not care how they got there. |
| `test_data/bad_label.tva` + `cmd/tva/main_test.go` | The `.bad label` fixture asserts the substring "invalid label" on stderr; keeping that prefix in the new bad-value wording leaves both unchanged. |
| `README.md` | Open checkbox `[ ] Labels for data preservation are supported` and the features bullet both need updating (final step). |

Baseline is green: `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` all clean.

## Design

### Line syntax

A trimmed line beginning with `.` is classified (one source of truth, shared by the extract, index, and build passes):

| Line shape | Kind |
|---|---|
| entire line matches `^\.[A-Za-z0-9_-]+$` | regular label declaration (unchanged; emits 0 words) |
| **first chunk** matches `^\.[A-Za-z0-9_-]+$` with one or more further chunks | **data label**: name = first chunk sans dot; each remaining chunk is a data value |
| otherwise (`.foo.`, `.`, `.$foo`) | invalid label — existing errors, unchanged |

The name shape exists once (`labelNamePattern`, factored out of the existing patterns); "entire line" (the existing whole-line declaration) and "first chunk" (the new `isDataLabel`) are the two anchorings of it. A label name is only the first chunk — exactly as `.data-12 12` declares the name `data-12`, not `data`.

### Value tokens

Each data value token must be exactly one of (all anchored):

| Token | Form | Word emitted |
|---|---|---|
| `-?\d+` | integer literal, sign optional (guidance #2) | the value itself |
| `i-?\d+` | immediate shorthand (guidance #1) | the value itself (identical to the literal form) |
| `i.<label>` | immediate reference (reuses `labelReferencePattern`) | the label's index |
| `$.<label>` | address reference (reuses `labelReferencePattern`) | the word found at the label's index (assembly-time dereference) |
| anything else | bare word, `1.5`, `12,`, `$5`, `-i12`, … | assembly error |

The `i` prefix disambiguates on its next character: `.` begins a label reference, a digit or `-` begins an immediate literal. Label names may begin with digits (`.1234` is valid), but the dot keeps the two forms disjoint. There is no comma tolerance: a trailing comma makes the token invalid, unlike instruction parameters.

### Semantics

- **Label value**: the same convention as a regular label — the word index of the word immediately following the label. For a data label the data words occupy that slot: `.data 12 34` at word index M → `data` = M, `memory[M]` = 12, `memory[M+1]` = 34.
- **Literals and `i.` references** emit their resolved value directly at build time.
- **`$.<label>`**: a data word is a bare integer — it cannot carry a parameter format, so the VM can never defer the dereference. The assembler resolves it: the word is the word stored at the label's index. If that label precedes an instruction, the value captured is the instruction's encoded opcode word (the acknowledged finickiness).
- **Forward references work** for all forms, matching how regular labels behave. A forward `$.` target's word is not yet emitted when the data label is built, so `$.` values emit a placeholder (0) plus a pending `(slot → label)` record, finalized in a new resolution pass.
- **Cycles are an error**: if following a `$.` chain revisits a still-pending word (`.a $.b` + `.b $.a`, or a self-reference), assembly fails with a cycle error naming the labels involved. The asymmetry is intentional: `.data i.data` (store one's own index) is fine; `.data $.data` (dereference one's own word) is a cycle.
- **No semantic policing of use** (consistent with the prior feature's guidance that the assembler makes no semantic assertions): jumping into data words, or executing them by falling through, is the programmer's problem — exactly as with any odd `$.` reference.
- **Noted asymmetry** (not fixed here): in *instruction* parameters a bare `-12` is rejected (only `i-12` works), while data values accept the bare negative per guidance #2. Unifying the two is a separate cleanup; the data value parser is deliberately the stricter of the two (anchored regexes; the `ReplaceAll` quirk in `addParam` is left as-is).

### Assembly pipeline (changes to `Assemble()`)

The existing three loops stay; each gains a data-label branch, plus a new resolution step:

1. **Extract**: `declareLabel` accepts the data form — parses the line via the new data-label parser, records the declaration (duplicate detection unchanged), and records every `i.`/`$.` reference in the value tokens for `ValidateAccessions`, so `.data $.nope` fails with the usual "referenced on line N but never declared". `ValidateAccessions` itself is unchanged.
2. **Index pass**: a data label line registers `labelIndexes[name] = currentIndex` and advances `currentIndex` by its **token count** — not via `instructionWords` (which would count the name chunk itself as a word and be off by one; it stays instruction-only, per its doc comment).
3. **Build**: data label lines dispatch to the new data-label emitter instead of `instructionBuilder`: literals and `i.` references emit their resolved word; `$.` references emit 0 and record a pending `(slot → label)` entry. The `instructionBuilder` remains label-unaware, as before.
4. **Resolution (new)**: after the word list is complete, for each pending slot follow the chain — slot → `labelIndexes[label]` → … — while the position is still pending, copy the value from the first finalized word reached, and drop the pending entry. Revisiting a position within one walk is a cycle error. Backward references resolve through this same code path with no special case; a program with no `$.` data values skips the pass trivially.

### Error policy

| Condition | Behavior |
|---|---|
| valid name + number / `i`-number / `i.<label>` / `$.<label>` | accepted |
| valid name + any other token (`.data foo`, `.data 1.5`, `.data $5`, `.data 12,`) | error prefixed `invalid label '<name>':` (keeps the `bad_label.tva` fixture test passing) explaining the accepted forms; line number included |
| valid name + non-space junk (`.foo.`) | unchanged: "must be alone on its line" (no token boundary, so it is not a data-label attempt) |
| invalid name (`.`, `.$foo`) | unchanged |
| duplicate declaration, either form | unchanged ("already declared") |
| `i.`/`$.` reference to a never-declared label in a value | unchanged validation ("referenced … never declared") |
| `$.` cycle (incl. self-reference) | new error naming the labels in the cycle |
| data label declared but unused / `$.` capturing an opcode word | silently accepted (no semantic assertions) |

## Expected behavior (worked examples)

- `.data 12` → `[12]`; `.data i12` → `[12]`; `.data -1` → `[-1]`; `.data i-12` → `[-12]`; `.data 1 2 3` → `[1, 2, 3]`.
- Finicky `$.` (backward): `.data $.foo` / `.foo` / `hlt` → `[9, 9]` (the data word is a copy of the `hlt` word at `foo`'s index).
- Forward `$.`: `.data $.foo` / `add 1 1 $0` / `.foo` / `hlt` → `[9, 1101, 1, 1, 0, 9]` (word 0 = word 4).
- Chain through another data label: `.one $.two` / `.two 42` / `hlt` → `[42, 42, 9]`.
- The owner's mixed example (`bar` = 5, `biz` = 6, `baz` = 8; `out 7` = `104 7`):

  ```
  .crazy-data__69420 i.bar 12 3 $.biz i.baz
  .bar
  hlt
  .biz
  out 7
  .baz
  ```

  → `[5, 12, 3, 104, 8, 9, 104, 7]` — `$.biz` captured the `out` **opcode word** 104, and the trailing `.baz` points past the end, as labels already may.
- Index accounting: a data label between instructions shifts subsequent label indexes (`.data 1 2` between an `add` and `.end` moves `end` by two; `out i.end` prints the shifted index).
- Cycle pair `.a $.b` / `.b $.a` and self-reference `.self` / `.data $.self` → cycle error; `.self` / `.data i.self` → `[0]` (no dereference, no cycle).

### Sample fixture (required, final step)

`test_data/data_labels.tva` — data labels placed after the `hlt` so the machine never executes them; the `$.` data value demonstrates the assembly-time copy:

```
# data_labels.tva
# Prints the first word stored at a data label, without ever executing the data.
out $.value
hlt
.value 12 34
.copy $.value
```

Assembles to `[4, 3, 9, 12, 34, 12]` — `value` = 3 (words `12 34` at 3–4), `copy` = 5 (word = `memory[3]` = 12). Execution: `out` in address mode reads `memory[3]` = **12** → prints `12`, then `hlt`. The generated `test_data/data_labels.tvm` is magic + 6 little-endian signed 32-bit words (28 bytes) and serves as byte-for-byte ground truth for the tva round-trip test.

## Code changes by file

- **`cmd/tva/data_label.go`** (new) — `dataLabel`/`dataValue` types; `parseDataValue` (the four token forms above, anchored) and `parseDataLabel`; the `isDataLabel` line classifier; word emission with pending-reference output; `resolveDataReferences` (chain walk + cycle detection).
- **`cmd/tva/label_extractor.go`** — factor the name shape into `labelNamePattern`; `declareLabel` branches on whole-line vs. name+values, delegating value parsing to `data_label.go` and recording value references for `ValidateAccessions`; doc comments updated (same-line content becomes the documented data-label form).
- **`cmd/tva/tsvetok_assembler.go`** — index pass: data labels contribute their token count; build pass: data lines dispatch to the data-label emitter, and the loop appends each statement's words directly (instead of holding `[]*instructionBuilder` to concatenate at the end) so the resolution pass can operate on the final word list; the resolution step runs before `Assemble` returns.
- **`cmd/tva/data_label_test.go`** (new) — value parser table (all four accepted forms; rejections: bare word, `1.5`, `$5`, `12,`, `-i12`, `i.`), multi-token parsing, word emission against a label table, resolution pass (backward / forward / chain / cycle / self-cycle).
- **`cmd/tva/label_extractor_test.go`** — extract-table rows for data lines (name declared; `i.`/`$.` references recorded); the three reworded rows in `TestLabelExtractor_RejectsInvalidLabelDeclarations`; duplicate data label; invalid-value rejection (message starts "invalid label"); undeclared reference in a value.
- **`cmd/tva/tsvetok_assembler_test.go`** — exact-intcode resolution tests (all worked examples above, including the owner's mixed example), the index-shift test, error tests (bare word, undeclared, duplicate, cycle, self-cycle), and run-on-VM forward-reference tests (`out $.data` prints the stored value).
- **`cmd/tva/main_test.go`** — the `bad_label` row stays unchanged; new round-trip row: sample `.tva` → byte-for-byte match with `data_labels.tvm`.
- **`cmd/tvm/main_test.go`** — new row: run `data_labels.tvm` → stdout `12\n`, exit 0.
- **Fixtures** — `test_data/data_labels.tva` + `test_data/data_labels.tvm` (generated via the tva executable; `*.tvm` is already marked binary in `.gitattributes`).
- **`AGENTS.md`** — the labels gotcha gains the data-label semantics (final step).
- **`README.md`** — check off "Labels for data preservation"; the features bullet specifies the value forms (final step).

## Implementation steps

Test-first; each step is one commit, red tests first then green, on branch `feature/implement-data-labels`:

1. **Parse data values** — `data_label.go` + `data_label_test.go` (value parser, name+values parsing, word count, emission, resolution with cycle detection). No behavior change: the extractor still rejects `.data 12` outright, so the suite stays green.
2. **Accept data labels in the extract pass** — red: `.data 12` must declare `data`; value references must feed `ValidateAccessions`; the three stale extractor rows are reworded. Green: the `declareLabel` branch + `labelNamePattern` factor.
3. **Emit and resolve data words in the assembler** — red: `.data 12` assembles to `[12]` (currently dies `unknown instruction '.data'`); all owner examples as exact intcode; the index-shift case; forward `$.`; cycle/self-cycle errors. Green: index-pass word count, build dispatch, placeholders + resolution pass. The meatiest step; if it proves unwieldy it splits into 3a (literals + `i.`) / 3b (`$.` + resolution) — but the value grammar is complete from step 1, so no intermediate "not yet supported" state should ever be visible.
4. **Documentation** — `AGENTS.md` labels gotcha (data-label syntax; label = first data word; `$.` resolved at assembly time; cycle error) and `README.md` (checkbox + features bullet).
5. **Sample fixture and end-to-end demonstration (required, final)** — add `test_data/data_labels.tva` + generated `data_labels.tvm`; tva `run()` round-trip row (byte-for-byte match); tvm `run()` row (stdout `12\n`, exit 0); assembler test reading the real `.tva` → assemble → run on the VM with mocks (output 12, six-word memory). Mirrors the labels feature' fixture finale.

Verification gate after each step: `go build ./... && go vet ./... && gofmt -l . && go test ./...` all clean.

## Out of scope (future work)

- **ASCII strings / other non-integer data** — the prior plan deferred "eventually ASCII strings". The value-token grammar here leaves room for a new token form (e.g. a quoted string emitting its digits as words).
- **`$<address>`-form data values** — a raw-address, label-less assembly-time dereference is not part of the declared grammar and stays rejected.
- **Jumping into data / executing data words** — no semantic policing, per the owner's standing guidance; a demonstration of that hazard belongs with a VM or tutorial step, not the assembler.
- **Register-form references; `jif` / `sub` / `mov` / `nil`; `call`/linking** — separate README TODOs; the only constraint here is that `instructionWords` and the reference machinery already anticipate them.
- **Unifying signed-immediate parsing in `addParam`** — the noted asymmetry (bare `-12` rejected in instruction parameters, accepted in data values) is a separate cleanup.
- **VM and file format** — untouched.
