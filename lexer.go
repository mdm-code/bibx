package parse

import (
	"strings"
	"unicode"
)

const (
	itmErr itmT = iota
	itmEOF
	itmEntryDelim // @
	itmLeftBrace  // {
	itmRightBrace // }
	itmEqSgn      // =
	itmComma      // ,
	itmCiteKey
	itmEntryType
	itmFieldType
	itmFieldText
)

const (
	null state = iota
	entryDelim
	entryType
	entryLeftBrace
	entryCiteKey
	entryComma
	entryFieldType
	entryRightBrace
	entryEqSgn
	entryFieldText
	entryTypeOrBrace
	eof
	err
)

// BibTeX entry syntactic element type.
type itmT uint8

// the state of the lexer.
type state uint8

// item is a single lexical syntactic element emitted by the lexer.
type item struct {
	t   itmT
	val string
}

// lexer parses BibTeX entries.
type lexer struct {
	reader  readable
	items   chan item
	states  map[state]func(*lexer) state
	state   state
	bracers int
	inEntry bool
}

// NewLexer creates a new lexer instance.
func newLexer(r readable) *lexer {
	return &lexer{
		reader: r,
		items:  make(chan item, 2),
		states: map[state]func(*lexer) state{
			null:             (*lexer).null,
			entryDelim:       (*lexer).entryDelim,
			entryType:        (*lexer).entryType,
			entryLeftBrace:   (*lexer).entryLeftBrace,
			entryRightBrace:  (*lexer).entryRightBrace,
			entryCiteKey:     (*lexer).citeKey,
			entryComma:       (*lexer).entryComma,
			entryFieldType:   (*lexer).entryFieldType,
			entryEqSgn:       (*lexer).entryEqSgn,
			entryFieldText:   (*lexer).entryFieldText,
			entryTypeOrBrace: (*lexer).entryTypeOrBrace,
			eof:              (*lexer).eof,
			err:              (*lexer).err,
		},
		state: null,
	}
}

// Item returns the next valid item parsed by the lexer.
func (l *lexer) item() item {
	for {
		select {
		case i := <-l.items:
			return i
		default:
			l.state = l.states[l.state](l)
		}
	}
}

// Null is the default startup lexer state.
func (l *lexer) null() state {
	return entryDelim
}

// EntryDelim seeks a new BibTeX entry delimiter.
func (l *lexer) entryDelim() state {
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch char.val {
		case '@':
			l.items <- item{t: itmEntryDelim, val: string(char.val)}
			return entryType
		}
	}
}

// EntryType parses the specified BibTeX entry type.
func (l *lexer) entryType() state {
	buf := ``
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch c := char.val; {
		case c == '{':
			buf = strings.TrimSpace(buf)
			if !isContinuous(buf) || !isLetter(buf) {
				return err
			}
			l.items <- item{t: itmEntryType, val: buf}
			defer l.reader.revert()
			return entryLeftBrace
		default:
			buf += string(char.val)
		}
	}
}

// EntryLeftBrace looks for the left brace character.
func (l *lexer) entryLeftBrace() state {
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		// NOTE: disallow nested entries
		if l.inEntry {
			return err
		}
		switch char.val {
		case '{':
			l.items <- item{t: itmLeftBrace, val: string(char.val)}
			l.bracers++
			l.inEntry = true
			return entryCiteKey
		}
	}
}

// EntryRightBrace looks for the right brace character.
func (l *lexer) entryRightBrace() state {
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		// NOTE: no entry to close
		if !l.inEntry {
			return err
		}
		switch char.val {
		case '}':
			l.items <- item{t: itmRightBrace, val: string(char.val)}
			l.bracers--
			l.inEntry = false
			return entryDelim
		}
	}
}

// CiteKey parses the provided BibTeX cite key.
func (l *lexer) citeKey() state {
	buf := ``
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch c := char.val; {
		case c == ',':
			buf = strings.TrimSpace(buf)
			if !isValidCiteKey(buf) {
				return err
			}
			l.items <- item{t: itmCiteKey, val: buf}
			defer l.reader.revert()
			return entryComma
		default:
			buf += string(c)
		}
	}
}

// EntryComma looks for the next comma character.
func (l *lexer) entryComma() state {
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch char.val {
		case ',':
			l.items <- item{t: itmComma, val: string(char.val)}
			return entryTypeOrBrace
		}
	}
}

// EntryTypeOrBrace checks if the next token is the field type or the end right brace.
func (l *lexer) entryTypeOrBrace() state {
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch c := char.val; {
		case c == '}':
			defer l.reader.revert()
			return entryRightBrace
		case !unicode.IsSpace(c):
			defer l.reader.revert()
			return entryFieldType
		}
	}
}

// EntryFieldType parses the field type identifier.
func (l *lexer) entryFieldType() state {
	buf := ``
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch char.val {
		case '=':
			buf = strings.TrimSpace(buf)
			if !isContinuous(buf) || !isLetter(buf) {
				return err
			}
			l.items <- item{t: itmFieldType, val: buf}
			defer l.reader.revert()
			return entryEqSgn
		default:
			buf += string(char.val)
		}
	}
}

// EntryEqSgn scans the reader for the equal sign character.
func (l *lexer) entryEqSgn() state {
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch char.val {
		case '=':
			l.items <- item{t: itmEqSgn, val: string(char.val)}
			return entryFieldText
		}
	}
}

// EntryFieldText
func (l *lexer) entryFieldText() state {
	buf := ``
	quotes := 0
	for {
		char := l.reader.next()
		if char.t == charErr {
			return err
		}
		if char.t == charEOF {
			return eof
		}
		switch c := char.val; {
		case c == '{':
			l.bracers++
			buf += string(char.val)
		case c == '"':
			quotes++
			buf += string(char.val)
		case c == '}' && l.bracers == 1:
			buf = strings.TrimSpace(buf)
			if !isValidInt(buf) {
				if !isProperQuoted(buf) {
					return err
				}
			}
			l.items <- item{t: itmFieldText, val: buf}
			defer l.reader.revert()
			return entryRightBrace
		case c == '}' && l.bracers > 0:
			l.bracers--
			buf += string(char.val)
		case c == ',' && quotes%2 == 0 && l.bracers == 1:
			buf = strings.TrimSpace(buf)
			if !isValidInt(buf) {
				if !isProperQuoted(buf) {
					return err
				}
			}
			l.items <- item{t: itmFieldText, val: buf}
			defer l.reader.revert()
			return entryComma
		default:
			buf += string(char.val)
		}
	}
}

// Eof puts the lexer in the continuous end-of-file state.
func (l *lexer) eof() state {
	l.items <- item{t: itmEOF, val: ``}
	return eof
}

// Err puts the lexer in the continuous error state.
func (l *lexer) err() state {
	l.items <- item{t: itmErr, val: ``}
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

// IsValidCiteKey verifies if the BibTeX cite key has valid characters.
func isValidCiteKey(s string) bool {
	if s == `` {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) &&
			!unicode.IsDigit(r) &&
			r != ':' &&
			r != '-' &&
			r != '_' {
			return false
		}
	}
	return true
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

	if !strings.HasPrefix(s, "\"") && !strings.HasPrefix(s, "{") {
		return false
	}

	if !strings.HasSuffix(s, "\"") && !strings.HasSuffix(s, "}") {
		return false
	}

	braces, quotes := 0, 0
	for _, c := range s {
		switch {
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
