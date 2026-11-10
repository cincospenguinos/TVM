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
