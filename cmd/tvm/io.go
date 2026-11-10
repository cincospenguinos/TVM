package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// stdinInput implements the machine's input interface by reading integers from a
// reader, one per line. End-of-file and unparseable input both yield 0, since the
// interface cannot carry an error
type stdinInput struct {
	scanner *bufio.Scanner
}

func (s *stdinInput) ReceiveInput() int {
	if !s.scanner.Scan() {
		return 0
	}

	number, err := strconv.Atoi(s.scanner.Text())
	if err != nil {
		return 0
	}

	return number
}

// stdoutOutput implements the machine's output interface by writing each integer
// to a writer, one per line
type stdoutOutput struct {
	writer io.Writer
}

func (o *stdoutOutput) EmitOutput(number int) {
	fmt.Fprintln(o.writer, number)
}
