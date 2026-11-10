package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"tvm/internal/tvmfile"
	tvm "tvm/internal/virtual_machine"
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

	program, err := tvmfile.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	machine := tvm.NewTsvetokVirtualMachine(program)
	machine.SetInputInterface(&stdinInput{scanner: bufio.NewScanner(stdin)})
	machine.SetOutputInterface(&stdoutOutput{writer: stdout})

	if err := machine.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}
