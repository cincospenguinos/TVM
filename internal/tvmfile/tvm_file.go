// Package tvmfile reads and writes TVM binary files. A TVM file begins with the four
// ASCII magic bytes `TVM\0`, followed by a sequence of little-endian signed 32-bit words
package tvmfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// magic is the four-byte ASCII prefix that identifies a TVM binary file
var magic = []byte("TVM\x00")

// Decode decodes a TVM binary blob into the sequence of words it contains. The blob
// must begin with the TVM magic bytes, and the remaining bytes must form whole
// little-endian signed 32-bit words
func Decode(data []byte) ([]int, error) {
	if !bytes.HasPrefix(data, magic) {
		return nil, fmt.Errorf("invalid magic bytes: expected 'TVM\\0'")
	}

	rest := data[len(magic):]
	if len(rest)%4 != 0 {
		return nil, fmt.Errorf("truncated file: %v bytes after the magic do not form whole 4-byte words", len(rest))
	}

	words := make([]int, 0, len(rest)/4)
	for i := 0; i < len(rest); i += 4 {
		words = append(words, int(int32(binary.LittleEndian.Uint32(rest[i:i+4]))))
	}

	return words, nil
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
