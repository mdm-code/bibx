package parse

import (
	"strings"
	"testing"
)

var text = `
@article{Cohen1963,
  author   = "P. J. Cohen, M. R. Thompson",
  title    = {The independence of {,} the hypothesis},
  journal  = "Proceedings of the {Academy} of Sciences",
  year     = 1963,
  volume   = "50",
  number   = "6",
  pages    = "1143--1148"
}
`

func testEntry() *strings.Reader {
	r := strings.NewReader(text)
	return r
}

func TestCharReader(t *testing.T) {
	r := newReader(testEntry())
	result := []byte{}
outer:
	for {
		char := r.next()
		if char.t == charErr || char.t == charEOF {
			break outer
		}
		result = append(result, byte(char.val))
	}
	if string(result) != text {
		t.Errorf("want %s; have %s", string(result), text)
	}
}
