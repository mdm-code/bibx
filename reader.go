package parse

import (
	"bufio"
	"io"
)

const (
	charErr charStatus = iota
	charEOF
	charOk
)

// Readable defines the reader interface expected by the lexer.
type readable interface {
	next() char
	revert() error
}

// CharStatus describes the status of the read character.
type charStatus uint8

// Char is a single character returned from the reader.
type char struct {
	t    charStatus
	size int
	val  rune
}

// Reader handles reading a file and exposing character elements.
type reader struct {
	buf *bufio.Reader
	pos int
}

// NewReader instantiates a new reader.
func newReader(r io.Reader) *reader {
	return &reader{bufio.NewReader(r), 0}
}

// Next returns the next available character.
func (r *reader) next() char {
	if c, s, err := r.buf.ReadRune(); err != nil {
		if err == io.EOF {
			return char{t: charEOF, size: s, val: c}
		}
		return char{t: charErr, size: s, val: c}
	} else {
		r.pos += s
		return char{t: charOk, size: s, val: c}
	}
}

// Revert unreads a single rune from the buffer.
func (r *reader) revert() error {
	return r.buf.UnreadRune()
}
