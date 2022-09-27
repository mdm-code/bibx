package parse

import (
	"strings"
	"testing"

	"github.com/mdm-code/bibx/internal/scan"
)

var haveEntryOne = `
% This is an example of a book entry type.
@book{bookExample,
  author    = {Peter Babington},
  title     = {The title of the work},
  publisher = {The name of the publisher},
  year      = 1993,
  volume    = 4,
  series    = 10,
  address   = {The address},
  edition   = 3,
  month     = 7,
  note      = {An optional note}
}
`

var wantEntryOne = &EntryDecl{
	Name:    "book",
	CiteKey: "bookExample",
	Comments: &CommentGroupExpr{
		Values: []*CommentExpr{
			{"% This is an example of a book entry type."},
		},
	},
	Fields: []*FieldStmt{
		{Key: "author", Value: "{Peter Babington}"},
		{Key: "title", Value: "{The title of the work}"},
		{Key: "publisher", Value: "{The name of the publisher}"},
		{Key: "year", Value: "1993"},
		{Key: "volume", Value: "4"},
		{Key: "series", Value: "10"},
		{Key: "address", Value: "{The address}"},
		{Key: "edition", Value: "3"},
		{Key: "month", Value: "7"},
		{Key: "note", Value: "{An optional note}"},
	},
}

var haveEntryTwo = `
% This is an example of a misc entry type.
@misc{miscExample,
  author       = {Peter Isley},
  title        = {The title of the work},
  howpublished = {How it was published},
  month        = 7,
  year         = 1993,
  note         = {An optional note}
}
`

var wantEntryTwo = &EntryDecl{
	Name:    "misc",
	CiteKey: "miscExample",
	Comments: &CommentGroupExpr{
		Values: []*CommentExpr{
			{"% This is an example of a misc entry type."},
		},
	},
	Fields: []*FieldStmt{
		{Key: "author", Value: "{Peter Isley}"},
		{Key: "title", Value: "{The title of the work}"},
		{Key: "howpublished", Value: "{How it was published}"},
		{Key: "month", Value: "7"},
		{Key: "year", Value: "1993"},
		{Key: "note", Value: "{An optional note}"},
	},
}

var haveAbbrev = `
% This is a comment on the abbreviation.
@string{btx = "{\textsc{Bib}\TeX}" }
`

var wantAbbrev = &AbbrevDecl{
	Comments: &CommentGroupExpr{
		Values: []*CommentExpr{
			{"% This is a comment on the abbreviation."},
		},
	},
	Field: &FieldStmt{Key: "btx", Value: `"{\textsc{Bib}\TeX}"`},
}

var havePreamble = `
% This is a comment on the preamble.
@PREAMBLE{"\makeatletter"}
`

var wantPreamble = &PreambleDecl{
	Comments: &CommentGroupExpr{
		Values: []*CommentExpr{
			{Value: "% This is a comment on the preamble."},
		},
	},
	Value: `"\makeatletter"`,
}

func TestParsedDecl(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   Node
	}{
		{
			name:   "first entry declaration",
			source: haveEntryOne,
			want:   wantEntryOne,
		},
		{
			name:   "second entry declaration",
			source: haveEntryTwo,
			want:   wantEntryTwo,
		},
		{
			name:   "preamble declaration",
			source: havePreamble,
			want:   wantPreamble,
		},
		{
			name:   "abbreviation declaration",
			source: haveAbbrev,
			want:   wantAbbrev,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := scan.NewReader(strings.NewReader(c.source))
			s := scan.NewScanner(r)
			p := NewParser(s)
			have, ok := p.Next()
			if !ok {
				t.Errorf("failed to parse the %v", c.name)
			}
			if !have.Eq(c.want) {
				t.Errorf("have %v; want %v", have, c.want)
			}
		})
	}
}
