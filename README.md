# Tsvetok Virtual Machine

For use in my TDD book. I realized I needed to write my own thing with my own virtual machine and have it work the way I want it instead of relying on another guy's work (freaking love that guy's work though.)


## TVM

A TVM reads a sequence of 32-bit integers, decodes them, and executes them.

### Register File

* Registers `$r0...$r4` are reserved between jumps
* Registers `$t0...$t8` are not reserved between jumps
* The last-jumped-address register, `$la`, is not reserved between jumps and so must be preserved between them

### File Format

TVM files are binary files with all bytes in little-endian. They begin with the ASCII characters `TVM` after which are sequences of instructions. Four bytes (32 bits) is one word in TVM, and so integers are received and munged in four byte chunks, until the footer `TVM EOF` is found.

### Operations

* Add (opcode `1`)
* Multiply (opcode `2`)
* Input (opcode `3`)
* Output (opcode `4`)
* Set-if-equal (opcode `5`)
* Set-if-not-equal (opcode `6`)
* Jump (opcode `7`)
	* Jump always sets the return register
* Halt (opcode `9`)

## TVA

Tsvetok assembly files are plain text UTF-8 files and have the following features:

* Comments are written with `#` character
* Labels are supported
* The `call` pseudo-instruction is supported, which the final step of assembly (linking) discovers, assembles, and copies into the machine

### TODO

* I think we need an object file type: something that is partially constructed with the exception of the `call` commands that the final assembly step picks up and loads all together.

## Tsvetalk

A higher level language with a grammar we compile down to TVA and the TVM format.

### TODO

* What is the grammar for this? I'm thinking some combination of Mini Java, Ruby, and Lox. I think the only primitive type should be an integer, and everything else is in its standard library. I think we should do a multi-pass compiler thing so we don't have to do header/source file splits (although I do really like that about C.)
