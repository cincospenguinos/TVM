// Package tvmfile reads and writes TVM binary files. A TVM file begins with the four
// ASCII magic bytes `TVM\0`, followed by a sequence of little-endian signed 32-bit words
package tvmfile

import (
	"fmt"
)

// Decode decodes a TVM binary blob into the sequence of words it contains
func Decode(data []byte) ([]int, error) {
	return nil, fmt.Errorf("not implemented")
}

// Encode encodes a sequence of words as a TVM binary blob
func Encode(program []int) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadFile reads and decodes the TVM binary file at the given path
func ReadFile(path string) ([]int, error) {
	return nil, fmt.Errorf("not implemented")
}

// WriteFile encodes the given program and writes it to the given path
func WriteFile(path string, program []int) error {
	return fmt.Errorf("not implemented")
}
