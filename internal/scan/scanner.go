package scan

import (
	"strings"
	"unicode"
)

const (
	ItemErr ItemType = iota
	ItemEOF
	ItemEntryDelim // @
	ItemLeftBrace  // {
	ItemRightBrace // }
	ItemLeftDelim  // {, (
	ItemRightDelim // }, )
	ItemLeftParen  // (
	ItemRightParen // )
	ItemEqSgn      // =
	ItemComma      // ,
	ItemCiteKey
	ItemEntry
	ItemComment
	ItemAbbrev
	ItemPreamble
	ItemFieldType
	ItemFieldText
	ItemTexCode
)

const (
	null state = iota
	entryDelim
	topLvlComment
	entryComment
	entryType
	entryLeftBodyDelim
	entryCiteKey
	entryComma
	entryFieldType
	entryRightBodyDelim
	entryEqSgn
	entryFieldText
	entryTypeOrBrace
	eof
	err
)

const (
	entry entryT = iota
	preamble
	abbrev
)

type Scannable interface {
	Next() Item
}

type (
	// BibTeX entry syntactic element type.
	ItemType uint8

	// The state of the scanner.
	state uint8

	// BibTeX entry type.
	entryT uint8
)

// Item is a single lexical syntactic element emitted by the scanner.
type Item struct {
	T   ItemType
	Val string
}

// Scanner parses BibTeX entries.
type Scanner struct {
	reader  readable
	items   chan Item
	states  map[state]func(*Scanner) state
	state   state
	bracers int
	entryT  entryT
	delim   rune
}

var delims = map[rune]rune{
	'{': '}',
	'}': '{',
	'(': ')',
	')': '(',
}

// NewScanner creates a new Scanner instance.
func NewScanner(r readable) *Scanner {
	return &Scanner{
		reader: r,
		items:  make(chan Item, 2), // buffered channel of size 2 is necessary and sufficent
		states: map[state]func(*Scanner) state{
			null:                (*Scanner).null,
			topLvlComment:       (*Scanner).topLvlComment,
			entryComment:        (*Scanner).entryComment,
			entryDelim:          (*Scanner).entryDelim,
			entryType:           (*Scanner).entryType,
			entryLeftBodyDelim:  (*Scanner).leftBodyDelim,
			entryRightBodyDelim: (*Scanner).rightBodyDelim,
			entryCiteKey:        (*Scanner).citeKey,
			entryComma:          (*Scanner).entryComma,
			entryFieldType:      (*Scanner).entryFieldType,
			entryEqSgn:          (*Scanner).entryEqSgn,
			entryFieldText:      (*Scanner).entryFieldText,
			entryTypeOrBrace:    (*Scanner).entryTypeOrBrace,
			eof:                 (*Scanner).eof,
			err:                 (*Scanner).err,
		},
		state: null,
	}
}

// Item returns the next valid Item parsed by the scanner.
func (s *Scanner) Next() Item {
	for {
		select {
		case i := <-s.items:
			return i
		default:
			s.state = s.states[s.state](s)
		}
	}
}

// Null is the default startup scanner state.
func (s *Scanner) null() state {
	return topLvlComment
}

func (s *Scanner) topLvlComment() state {
	buf := ``
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case '@':
			defer s.reader.Revert()
			buf = strings.TrimSpace(buf)
			if buf != "" {
				s.items <- Item{T: ItemComment, Val: buf}
			}
			return entryDelim
		default:
			buf += string(char.val)
		}
	}
}

// EntryDelim seeks a new BibTeX entry delimiter.
func (s *Scanner) entryDelim() state {
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case '@':
			s.items <- Item{T: ItemEntryDelim, Val: string(char.val)}
			return entryType
		}
	}
}

// EntryType parses the specified BibTeX entry type.
func (s *Scanner) entryType() state {
	buf := ``
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		var t ItemType
		switch char.val {
		case '{', '(':
			buf = strings.TrimSpace(buf)
			lower := strings.ToLower(buf)
			if lower == "preamble" {
				s.entryT = preamble
				t = ItemPreamble
			} else if lower == "string" {
				s.entryT = abbrev
				t = ItemAbbrev
			} else {
				s.entryT = entry
				t = ItemEntry
			}
			if !IsValidName(buf) {
				return err
			}
			s.items <- Item{T: t, Val: buf}
			defer s.reader.Revert()
			return entryLeftBodyDelim
		default:
			buf += string(char.val)
		}
	}
}

// EntryLeftBrace looks for the left brace character.
func (s *Scanner) leftBodyDelim() state {
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case '{', '(':
			s.items <- Item{T: ItemLeftDelim, Val: string(char.val)}
			s.delim = char.val
			s.bracers++
			switch s.entryT {
			case entry:
				return entryCiteKey
			case preamble:
				return entryFieldText
			case abbrev:
				return entryFieldType
			}
		}
	}
}

// EntryRightBrace looks for the right brace character.
func (s *Scanner) rightBodyDelim() state {
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case '}', ')':
			if !delimsMatch(s.delim, char.val) {
				return err
			}
			s.items <- Item{T: ItemRightDelim, Val: string(char.val)}
			s.bracers--
			return null
		}
	}
}

// CiteKey parses the provided BibTeX cite key.
func (s *Scanner) citeKey() state {
	buf := ``
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch c := char.val; {
		case c == ',':
			buf = strings.TrimSpace(buf)
			if !IsValidName(buf) {
				return err
			}
			s.items <- Item{T: ItemCiteKey, Val: buf}
			defer s.reader.Revert()
			return entryComma
		default:
			buf += string(c)
		}
	}
}

// EntryComma looks for the next comma character.
func (s *Scanner) entryComma() state {
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case ',':
			s.items <- Item{T: ItemComma, Val: string(char.val)}
			return entryTypeOrBrace
		}
	}
}

func (s *Scanner) entryComment() state {
	buf := ``
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case '\n':
			// emit the item and traverse to the next state
			buf = strings.TrimSpace(buf)
			if buf != "" {
				s.items <- Item{T: ItemComment, Val: buf}
			}
			goto cont
		default:
			buf += string(char.val)
		}
	}

cont:
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch c := char.val; {
		case c == '%':
			return entryComment
		case isDelim(c):
			s.reader.Revert()
			return entryRightBodyDelim
		case IsValidNameRune(c):
			s.reader.Revert()
			return entryFieldType
		}
	}
}

// EntryTypeOrBrace checks if the next token is the field type or the end right
// brace.
func (s *Scanner) entryTypeOrBrace() state {
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch c := char.val; {
		case c == '}' || c == ')':
			defer s.reader.Revert()
			return entryRightBodyDelim
		case c == '%':
			return entryComment
		case IsValidNameRune(c):
			defer s.reader.Revert()
			return entryFieldType
		}
	}
}

// EntryFieldType parses the field type identifier.
func (s *Scanner) entryFieldType() state {
	buf := ``
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case '=':
			buf = strings.TrimSpace(buf)
			if !IsValidName(buf) {
				return err
			}
			s.items <- Item{T: ItemFieldType, Val: buf}
			defer s.reader.Revert()
			return entryEqSgn
		default:
			buf += string(char.val)
		}
	}
}

// EntryEqSgn scans the reader for the equal sign character.
func (s *Scanner) entryEqSgn() state {
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch char.val {
		case '=':
			s.items <- Item{T: ItemEqSgn, Val: string(char.val)}
			return entryFieldText
		}
	}
}

// EntryFieldText reads character from the reader looking for the text
// delimiter.
func (s *Scanner) entryFieldText() state {
	buf := ``
	quotes := 0
	var prev rune
	for {
		char := s.reader.Next()
		if state := checkErr(char); state != null {
			return state
		}
		switch c := char.val; {
		case c == '{':
			s.bracers++
			buf += string(char.val)
		case c == '"':
			if prev != '\\' {
				quotes++
			}
			buf += string(char.val)
		case (c == '}' || c == ')') && s.bracers == 1:
			buf = strings.TrimSpace(buf)
			if !isValidInt(buf) {
				if !isProperQuoted(buf) {
					return err
				}
			}
			s.items <- Item{T: ItemFieldText, Val: buf}
			defer s.reader.Revert()
			return entryRightBodyDelim
		case c == '%' && s.bracers == 1:
			buf = strings.TrimSpace(buf)
			if !isValidInt(buf) {
				if !isProperQuoted(buf) {
					return err
				}
			}
			s.items <- Item{T: ItemFieldText, Val: buf}
			return entryComment
		case c == '}' && s.bracers > 0:
			s.bracers--
			buf += string(char.val)
		case c == ',' && quotes%2 == 0 && s.bracers == 1:
			buf = strings.TrimSpace(buf)
			if !isValidInt(buf) {
				if !isProperQuoted(buf) {
					return err
				}
			}
			s.items <- Item{T: ItemFieldText, Val: buf}
			defer s.reader.Revert()
			return entryComma
		default:
			buf += string(char.val)
		}
		prev = char.val
	}
}

// Eof puts the scanner in the continuous end-of-file state.
func (s *Scanner) eof() state {
	s.items <- Item{T: ItemEOF, Val: ``}
	return eof
}

// Err puts the scanner in the continuous error state.
func (s *Scanner) err() state {
	s.items <- Item{T: ItemErr, Val: ``}
	return err
}

// IsContinuous checks if a string contains white space characters.
func isContinuous(s string) bool {
	if s == `` {
		return false
	}
	for _, r := range s {
		if unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// IsValidName verifies if the BibTeX NAME has only valid characters.
func IsValidName(s string) bool {
	if s == `` {
		return false
	}
	for _, r := range s {
		if !IsValidNameRune(r) {
			return false
		}
	}
	return true
}

// IsValidNameRune checks if the rune is a valid BibTeX NAME character.
func IsValidNameRune(r rune) bool {
	if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !IsSpecial(r) {
		return false
	}
	return true
}

// IsSpecial checks if the the rune is an allowed BibTeX NAME character.
func IsSpecial(r rune) bool {
	for _, c := range "_-/!?$&*+.:;<>[]^`|" {
		if r == c {
			return true
		}
	}
	return false
}

// IsDelim tells whether a character is an entry delimiter.
func isDelim(r rune) bool {
	for _, c := range "{}()" {
		if r == c {
			return true
		}
	}
	return false
}

// isInteger checks if the string is composed of digits only.
func isValidInt(s string) bool {
	if s == `` {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsLetter tests if the string comprises of letters only.
func isLetter(s string) bool {
	if s == `` {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsProperQuoted checks if the string is enclosed in quotation marks or curly
// brackets.
func isProperQuoted(s string) bool {
	if s == `` {
		return false
	}

	braces, quotes := 0, 0

	chars := []rune(s)
	for i := 0; i < len(chars); i++ {
		switch c := chars[i]; {
		case c == '\\':
			// Skip over the next escaped character, e.g. ", {, }
			i++
		case c == '{':
			braces++
		case c == '}' && braces > 0:
			braces--
		case c == '"':
			quotes++
		}
	}
	if braces != 0 || quotes%2 != 0 {
		return false

	}
	return true
}

// DelimsMatch checks if two entry delimiters form a match.
func delimsMatch(i, j rune) bool {
	other, ok := delims[i]
	if !ok {
		return false
	}
	if j != other {
		return false
	}
	return true
}

func checkErr(c char) state {
	if c.t == charErr {
		return err
	}
	if c.t == charEOF {
		return eof
	}
	return null
}
