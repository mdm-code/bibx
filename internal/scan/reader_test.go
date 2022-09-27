package scan

import (
	"strings"
	"testing"
)

func testTexEntry() *strings.Reader {
	r := strings.NewReader(texEntry)
	return r
}

func testTexString() *strings.Reader {
	r := strings.NewReader(texStrings)
	return r
}

func testTexPreamble() *strings.Reader {
	r := strings.NewReader(texPreamble)
	return r
}

func TestCharReader(t *testing.T) {
	r := NewReader(testTexEntry())
	result := []byte{}
outer:
	for {
		char := r.Next()
		if char.t == charErr || char.t == charEOF {
			break outer
		}
		result = append(result, byte(char.val))
	}
	if string(result) != texEntry {
		t.Errorf("want %s; have %s", string(result), texEntry)
	}
}
