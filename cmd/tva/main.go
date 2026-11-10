package main

import (
	"fmt"
	"io"
	"os"

	"tvm/internal/tvmfile"
)

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]))
}

// run assembles the TVA source file at args[0] and writes the resulting TVM
// binary to args[1]. It returns the process exit code: 2 for usage errors, 1
// for read, assembly, or write errors, and 0 on success
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
