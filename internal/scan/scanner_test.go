package scan

import (
	"reflect"
	"testing"
)

var texEntry = `
% The author never intended to write this book.
@article(Cohen1963,
  % this is a comment.
  % the next line is just to test this.
  author   = "P. J. C{\"o}hen, M. R. Thompson",
  title    = {The independence of {,} the hypothesis},
  journal  = "Proceedings of the $\eq{2}$ {Academy} of Sciences",
  year     = 1963, % this is a comment.
  volume   = "50",
  number   = "6",
  pages    = "1143--1148" % this is a comment.
  % this is a comment.
)
`

var texPreamble = `
@PREAMBLE{ "\@ifundefined{url}{\def\url#1{\texttt{#1}}}{}" }
`

var texStrings = `
@string{goossens = "Goossens, Michel"}
`

var entryItems = []Item{
	{ItemComment, `% The author never intended to write this book.`},
	{ItemEntryDelim, `@`},
	{ItemEntry, `article`},
	{ItemLeftDelim, `(`},
	{ItemCiteKey, `Cohen1963`},
	{ItemComma, `,`},
	{ItemComment, `this is a comment.`},
	{ItemComment, `the next line is just to test this.`},
	{ItemFieldType, `author`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `"P. J. C{\"o}hen, M. R. Thompson"`},
	{ItemComma, `,`},
	{ItemFieldType, `title`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `{The independence of {,} the hypothesis}`},
	{ItemComma, `,`},
	{ItemFieldType, `journal`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `"Proceedings of the $\eq{2}$ {Academy} of Sciences"`},
	{ItemComma, `,`},
	{ItemFieldType, `year`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `1963`},
	{ItemComma, `,`},
	{ItemComment, `this is a comment.`},
	{ItemFieldType, `volume`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `"50"`},
	{ItemComma, `,`},
	{ItemFieldType, `number`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `"6"`},
	{ItemComma, `,`},
	{ItemFieldType, `pages`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `"1143--1148"`},
	{ItemComment, `this is a comment.`},
	{ItemComment, `this is a comment.`},
	{ItemRightDelim, `)`},
}

var preambleItems = []Item{
	{ItemEntryDelim, `@`},
	{ItemPreamble, `PREAMBLE`},
	{ItemLeftDelim, `{`},
	{ItemFieldText, `"\@ifundefined{url}{\def\url#1{\texttt{#1}}}{}"`},
	{ItemRightDelim, `}`},
}

var stringItems = []Item{
	{ItemEntryDelim, `@`},
	{ItemAbbrev, `string`},
	{ItemLeftDelim, `{`},
	{ItemFieldType, `goossens`},
	{ItemEqSgn, `=`},
	{ItemFieldText, `"Goossens, Michel"`},
	{ItemRightDelim, `}`},
}

func TestLexerPreamble(t *testing.T) {
	r := NewReader(testTexPreamble())
	result := []Item{}
	l := NewScanner(r)
	itm := l.Next()
	for {
		if itm.T == ItemEOF || itm.T == ItemErr {
			break
		}
		result = append(result, itm)
		itm = l.Next()
	}
	if ok := reflect.DeepEqual(preambleItems, result); !ok {
		t.Errorf("want %v; have: %v", entryItems, result)
	}
}

func TestLexerEntry(t *testing.T) {
	r := NewReader(testTexEntry())
	result := []Item{}
	l := NewScanner(r)
	itm := l.Next()
	for {
		if itm.T == ItemEOF || itm.T == ItemErr {
			break
		}
		result = append(result, itm)
		itm = l.Next()
	}
	if ok := reflect.DeepEqual(entryItems, result); !ok {
		t.Errorf("want %v; have: %v", entryItems, result)
	}
}

func TextLexerString(t *testing.T) {
	r := NewReader(testTexString())
	result := []Item{}
	l := NewScanner(r)
	itm := l.Next()
	for {
		if itm.T == ItemEOF || itm.T == ItemErr {
			break
		}
		result = append(result, itm)
		itm = l.Next()
	}
	if ok := reflect.DeepEqual(preambleItems, result); !ok {
		t.Errorf("want %v; have: %v", entryItems, result)
	}
}

func TestIsContinuous(t *testing.T) {
	cases := []struct {
		name      string
		testInput string
		want      bool
	}{
		{"space", "Cohen 1963", false},
		{"newline", "John\nDoe", false},
		{"tab", "M\tJames1992", false},
		{"trailing", "Trimm1999  ", false},
		{"ok", "Trudgill1936", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if have := isContinuous(c.testInput); have != c.want {
				t.Errorf("for %s :: have %t; want %t", c.testInput, have, c.want)
			}
		})
	}
}

func TestValidCiteKey(t *testing.T) {
	cases := []struct {
		name      string
		testInput string
		want      bool
	}{
		{"basic", "companion", true},
		{"alphanumeric", "Chomsky1965", true},
		{"complex", "book:N_Chomsky1968", true},
		{"space", "N Chomsky 1965", false},
		{"failing", "book = NC1963", false},
		{"empty", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if have := IsValidName(c.testInput); have != c.want {
				t.Errorf("for %s :: have: %t; want: %t", c.testInput, have, c.want)
			}
		})
	}
}

func TestValidInteger(t *testing.T) {
	cases := []struct {
		name      string
		testInput string
		want      bool
	}{
		{"date", "1984", true},
		{"page", "50", true},
		{"number", "6", true},
		{"pages", "12--25", false},
		{"chapter", "3.", false},
		{"string", "C. J. Thompson", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if have := isValidInt(c.testInput); have != c.want {
				t.Errorf("for %s :: have: %t; want: %t", c.testInput, have, c.want)
			}
		})
	}
}

func TestIsLetter(t *testing.T) {
	cases := []struct {
		name      string
		testInput string
		want      bool
	}{
		{"article", "article", true},
		{"BOOK", "book", true},
		{"punctuation", "article-12", false},
		{"digits", "book198", false},
		{"whitespace", "in collection", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if have := isLetter(c.testInput); have != c.want {
				t.Errorf("for %s :: have: %t; want: %t", c.testInput, have, c.want)
			}
		})
	}
}

func TestIsProperQuoted(t *testing.T) {
	cases := []struct {
		name      string
		testInput string
		want      bool
	}{
		{"simple-quotes", `"Brooks, Michael and Russel, Robert"`, true},
		{"simple-brackets", "{The independence of the hypothesis}", true},
		{"elaborate-brackets", `{The {Death} of an "Author"}`, true},
		{"elaborate-quote", `"The {D}eath of an {"}Author{"}"`, true},
		{"quote-pages", `"1234--5843"`, true},
		{"simple-missing", `"Pale {F}ire`, false},
		{"elaborate-missing", `{Pale "{Fire"}`, false},
		{"escaped-quotation-mark", `{C{\"o}hen}`, true},
		{"empty", ``, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if have := isProperQuoted(c.testInput); have != c.want {
				t.Errorf("for %s :: have: %t; want %t", c.testInput, have, c.want)
			}
		})
	}
}
