// Package tvmfile reads and writes TVM binary files. A TVM file begins with the four
// ASCII magic bytes `TVM\0`, followed by a sequence of little-endian signed 32-bit words
package tvmfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
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

// Encode encodes a sequence of words as a TVM binary blob: the magic bytes followed
// by each word as a little-endian signed 32-bit integer. Words outside the signed
// 32-bit range return an error
func Encode(program []int) ([]byte, error) {
	out := make([]byte, len(magic)+4*len(program))
	copy(out, magic)

	for i, word := range program {
		if word < math.MinInt32 || word > math.MaxInt32 {
			return nil, fmt.Errorf("word %v (%v) does not fit in a signed 32-bit integer", i, word)
		}

		binary.LittleEndian.PutUint32(out[len(magic)+4*i:], uint32(int32(word)))
	}

	return out, nil
}

// ReadFile reads the TVM binary file at the given path and decodes it
func ReadFile(path string) ([]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	words, err := Decode(data)
	if err != nil {
		return nil, fmt.Errorf("reading %v: %w", path, err)
	}

	return words, nil
}

// WriteFile encodes the given program and writes it to the given path
func WriteFile(path string, program []int) error {
	data, err := Encode(program)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %v: %w", path, err)
	}

	return nil
}
