# Tsvetok Virtual Machine

For use in my TDD book. I realized I needed to write my own thing with my own virtual machine and have it work the way I want it instead of relying on another guy's work (freaking love that guy's work though.)


## TVM

A TVM reads a sequence of 32-bit integers, decodes them, and executes them.

### Operations

* Add (opcode `1`)
* Multiply (opcode `2`)
* Input (opcode `3`)
* Output (opcode `4`)
* Set-if-equal (opcode `5`)
* Jump if true (opcode `6`)
	* Jump always sets the return register if it takes the jump
* Set-if-less-than (opcode `7`)
	* Turns out I need it
* Halt (opcode `9`)

### Register File

* Registers `$r0...$r4` are reserved between jumps
* Registers `$t0...$t8` are not reserved between jumps
* The last-jumped-address register, `$la`, is not reserved between jumps and so must be preserved between them. This register also cannot be written to.

The registers are enumerated as follows:

```
r0...r4 -> 0, 1, 2, 3, 4
t0...t7 -> 5, 6, 7, 8, 9, 10, 11, 12
la       -> 13
```

### Operation Types

The first digit indicates the first operand's type, the second the second, and so on for as many operands exist. There are three types:

* Memory type, indicated by `0`
* Immediate type, indicated by `1`
* Register type, indicated by `2`

Memory type means "this operand is an address in memory." Immediate type means "this operand is an integer value to be read as an integer value." Register type means "this operand's value refers to a register in the register file."

### File Format

TVM files are binary files with all bytes in little-endian. They begin with the ASCII characters `TVM` after which are sequences of instructions. Four bytes (32 bits) is one word in TVM, and so integers are received and munged in four byte chunks, until the file ends

### TODO

- [x] Halt instruction
- [x] Add instruction
- [x] Multiply instruction
- [x] Input instruction
- [x] Output instruction
- [x] Sane defaults for Input/Output operations somewhere
- [x] All operations support address mode
- [x] All operations support immediate mode
- [x] All operations support register mode
	* Actually I'm not sure I want to support register mode yet
- [ ] Set-less-than instruction
- [ ] Any memory address that does not exist will immediately exist upon lookup or writing
	* If we expand memory to fill the space, we set everything inside to 0
- [ ] Read a TVM binary file and executes it
- [ ] Auto expands memory when attempting to access a valid location
	* If it's past the length of memory, then we can expand it. It would be a nice quality of life feature
- [ ] Do we want to allow the program counter to be a read-only register by programmers? It'd be a nice quality of life thing (`out pc` could act like a print statement)

## TVA

Tsvetok assembly files are plain text UTF-8 files and have the following features:

* Comments are written with `#` character
* Labels are supported
* The `call` pseudo-instruction is supported, which the final step of assembly (linking) discovers, assembles, and copies into the machine

### TODO

- [x] `hlt` is supported
- [x] `add` is supported
- [x] `mlt` is supported
- [x] `in` is supported
- [x] `out` is supported
- [x] `seq` is supported
- [x] `jit` is supported
- [ ] Labels for jumping are supported
- [ ] Labels for data preservation are supported
- [x] All operations support immediates
- [x] All operations support registers
- [ ] `jif` pseudo-instruction is supported
- [ ] `sub` pseudo-instruction is supported
- [ ] `nil` psuedo-instruction is supported
	* This sets the underlying value to simply 0 unconditionally
- [ ] `mov` pseudo-instruction is supported
	* This copies the source value at the destination (length of three)
- [x] Comments are removed and ignored
- [ ] Writes to a TVM binary file with correct syntax
- [ ] Do we want to do validation in the assembler? I think we do. If there's a semantic error with the execution of the underlying program, the programmer really ought to know.

## Tsvetalk

A higher level language with a grammar we compile down to TVA and the TVM format.

### TODO

* What is the grammar for this? I'm thinking some combination of Mini Java, Ruby, and Lox. I think the only primitive type should be an integer, and everything else is in its standard library. I think we should do a copypasta import kind of thing, but maybe with a multi-pass compiler (I like `#include` directives, personally)

## PROJECT TODOS

- [ ] GitLab/Travis CI pipeline for building/testing/deploying?
- [ ] golangci-lint to check things like errors and stuff
