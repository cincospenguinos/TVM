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
// binary to args[1]. It returns the process exit code: 2 for usage errors, 1
// for read, assembly, or write errors, and 0 on success
func run(stdin io.Reader, stdout, stderr io.Writer, args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: tva <input.tva> <output.tvm>")
		return 2
	}

	return 1 // TODO: assemble and write the program file
}
