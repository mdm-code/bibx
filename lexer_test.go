package parse

import (
	"reflect"
	"testing"
)

var items = []item{
	{itmEntryDelim, `@`},
	{itmEntryType, `article`},
	{itmLeftBrace, `{`},
	{itmCiteKey, `Cohen1963`},
	{itmComma, `,`},
	{itmFieldType, `author`},
	{itmEqSgn, `=`},
	{itmFieldText, `"P. J. Cohen, M. R. Thompson"`},
	{itmComma, `,`},
	{itmFieldType, `title`},
	{itmEqSgn, `=`},
	{itmFieldText, `{The independence of {,} the hypothesis}`},
	{itmComma, `,`},
	{itmFieldType, `journal`},
	{itmEqSgn, `=`},
	{itmFieldText, `"Proceedings of the {Academy} of Sciences"`},
	{itmComma, `,`},
	{itmFieldType, `year`},
	{itmEqSgn, `=`},
	{itmFieldText, `1963`},
	{itmComma, `,`},
	{itmFieldType, `volume`},
	{itmEqSgn, `=`},
	{itmFieldText, `"50"`},
	{itmComma, `,`},
	{itmFieldType, `number`},
	{itmEqSgn, `=`},
	{itmFieldText, `"6"`},
	{itmComma, `,`},
	{itmFieldType, `pages`},
	{itmEqSgn, `=`},
	{itmFieldText, `"1143--1148"`},
	{itmRightBrace, `}`},
}

func TestLexer(t *testing.T) {
	r := newReader(testEntry())
	result := []item{}
	l := newLexer(r)
	itm := l.item()
	for {
		if itm.t == itmEOF || itm.t == itmErr {
			break
		}
		result = append(result, itm)
		itm = l.item()
	}
	if ok := reflect.DeepEqual(items, result); !ok {
		t.Errorf("want %v; have: %v", items, result)
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
			if have := isValidCiteKey(c.testInput); have != c.want {
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
		{"no-quotes", `The University of Arizona`, false},
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
