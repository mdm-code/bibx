package parse

import (
	"reflect"
	"strings"

	"github.com/mdm-code/bibx/internal/scan"
)

const (
	NodeBadDecl NodeT = iota
	NodeEntry
	NodeAbbrev
	NodePreamble
	NodeBadStmt
	NodeFieldStmt
	NodeBadExpr
	NodeCommentExpr
	NodeCommentGroupExpr
)

const (
	null state = iota
	comms
	decl
	entry
	preamble
	abbrev
	err
	eof
)

var nodeNames = [...]string{
	NodeBadDecl:          "NodeBadDecl",
	NodeEntry:            "NodeEntry",
	NodeAbbrev:           "NodeAbbrev",
	NodePreamble:         "NodePreamble",
	NodeBadStmt:          "NodeBadStmt",
	NodeFieldStmt:        "NodeFieldStmt",
	NodeBadExpr:          "NodeBadExpr",
	NodeCommentExpr:      "NodeCommentExpr",
	NodeCommentGroupExpr: "NodeCommentGroupExpr",
}

type Node interface {
	Type() NodeT
	Eq(Node) bool
}

type NodeT uint8

type state uint8

type (
	EntryDecl struct {
		Name     string
		CiteKey  string
		Comments *CommentGroupExpr
		Fields   []*FieldStmt
	}

	AbbrevDecl struct {
		Comments *CommentGroupExpr
		Field    *FieldStmt
	}

	PreambleDecl struct {
		Comments *CommentGroupExpr
		Value    string
	}

	BadDecl struct{}

	FieldStmt struct {
		Key, Value string
	}

	BadStmt struct{}

	CommentGroupExpr struct {
		Values []*CommentExpr
	}

	CommentExpr struct {
		Value string
	}

	BadExpr struct{}
)

type Parser struct {
	scanner  scan.Scannable
	nodes    chan Node
	comments *CommentGroupExpr
	currDecl Node
	states   map[state]func(*Parser) state
	state    state
}

func NewParser(s scan.Scannable) *Parser {
	return &Parser{
		scanner: s,
		nodes:   make(chan Node, 2),
		states: map[state]func(*Parser) state{
			null:     (*Parser).null,
			comms:    (*Parser).comms,
			decl:     (*Parser).decl,
			entry:    (*Parser).entry,
			preamble: (*Parser).preamble,
			abbrev:   (*Parser).abbrev,
			err:      (*Parser).err,
			eof:      (*Parser).eof,
		},
		comments: new(CommentGroupExpr),
		state:    null,
	}
}

func (*EntryDecl) Type() NodeT      { return NodeEntry }
func (e *EntryDecl) String() string { return nodeNames[e.Type()] }

func (e *EntryDecl) Eq(n Node) bool {
	d, ok := n.(*EntryDecl)
	if !ok {
		return false
	}
	if e.Name != d.Name {
		return false
	}
	if e.CiteKey != d.CiteKey {
		return false
	}
	if !e.Comments.Eq(d.Comments) {
		return false
	}
	if !reflect.DeepEqual(e.Fields, d.Fields) {
		return false
	}
	return true
}

func (*AbbrevDecl) Type() NodeT      { return NodeAbbrev }
func (a *AbbrevDecl) String() string { return nodeNames[a.Type()] }

func (a *AbbrevDecl) Eq(n Node) bool {
	d, ok := n.(*AbbrevDecl)
	if !ok {
		return false
	}
	if !a.Field.Eq(d.Field) {
		return false
	}
	return true
}

func (*PreambleDecl) Type() NodeT      { return NodePreamble }
func (p *PreambleDecl) String() string { return nodeNames[p.Type()] }

func (p *PreambleDecl) Eq(n Node) bool {
	d, ok := n.(*PreambleDecl)
	if !ok {
		return false
	}
	if p.Value != d.Value {
		return false
	}
	if !p.Comments.Eq(d.Comments) {
		return false
	}
	return true
}

func (*BadDecl) Type() NodeT      { return NodeBadDecl }
func (b *BadDecl) String() string { return nodeNames[b.Type()] }

func (b *BadDecl) Eq(n Node) bool {
	if _, ok := n.(*BadDecl); !ok {
		return false
	}
	return true
}

func (*FieldStmt) Type() NodeT      { return NodeFieldStmt }
func (f *FieldStmt) String() string { return nodeNames[f.Type()] }

func (f *FieldStmt) Eq(n Node) bool {
	d, ok := n.(*FieldStmt)
	if !ok {
		return false
	}
	if f.Key != d.Key {
		return false
	}
	if f.Value != d.Value {
		return false
	}
	return true
}

// Ok checks whether a statement has both a key and value set.
func (f *FieldStmt) ok() bool {
	if f.Key == `` || f.Value == `` {
		return false
	}
	return true
}

func (*BadStmt) Type() NodeT      { return NodeBadStmt }
func (b *BadStmt) String() string { return nodeNames[b.Type()] }

func (b *BadStmt) Eq(n Node) bool {
	if _, ok := n.(*BadStmt); !ok {
		return false
	}
	return true
}

func (*CommentGroupExpr) Type() NodeT      { return NodeCommentGroupExpr }
func (c *CommentGroupExpr) String() string { return nodeNames[c.Type()] }

func (c *CommentGroupExpr) Eq(n Node) bool {
	d, ok := n.(*CommentGroupExpr)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(c.Values, d.Values) {
		return false
	}
	return true
}

func (*CommentExpr) Type() NodeT      { return NodeCommentExpr }
func (c *CommentExpr) String() string { return nodeNames[c.Type()] }

func (c *CommentExpr) Eq(n Node) bool {
	d, ok := n.(*CommentExpr)
	if !ok {
		return false
	}
	if c.Value != d.Value {
		return false
	}
	return true
}

func (*BadExpr) Type() NodeT      { return NodeBadExpr }
func (b *BadExpr) String() string { return nodeNames[b.Type()] }

func (b *BadExpr) Eq(n Node) bool {
	if _, ok := n.(*BadExpr); !ok {
		return false
	}
	return true
}

func (p *Parser) Next() (Node, bool) {
	for {
		select {
		case n, ok := <-p.nodes:
			return n, ok
		default:
			p.state = p.states[p.state](p)
		}
	}
}

func (p *Parser) resetComms() { p.comments = new(CommentGroupExpr) }

func (p *Parser) resetDecl() { p.currDecl = nil }

func (p *Parser) null() state {
	return comms
}

func (p *Parser) err() state {
	defer close(p.nodes)
	return err
}

func (p *Parser) eof() state {
	defer close(p.nodes)
	return eof
}

func (p *Parser) comms() state {
	for {
		i := p.scanner.Next()
		if state := checkErr(i.T); state != null {
			return state
		}
		switch i.T {
		case scan.ItemComment:
			v := CommentExpr{i.Val}
			p.comments.Values = append(p.comments.Values, &v)
		case scan.ItemEntryDelim:
			return decl
		default:
			p.resetComms()
			return err
		}
	}
}

func (p *Parser) decl() state {
	i := p.scanner.Next()
	if state := checkErr(i.T); state != null {
		return state
	}
	switch i.T {
	case scan.ItemEntry:
		lower := strings.ToLower(i.Val)
		decl := EntryDecl{Name: lower}
		p.currDecl = &decl
		return entry
	case scan.ItemAbbrev:
		decl := AbbrevDecl{}
		p.currDecl = &decl
		return abbrev
	case scan.ItemPreamble:
		decl := PreambleDecl{}
		p.currDecl = &decl
		return preamble
	}
	return err
}

func (p *Parser) entry() state {
	decl, ok := p.currDecl.(*EntryDecl)
	if !ok {
		return err
	}

	stmt := &FieldStmt{}
	var i scan.Item

	// Consume body delimiter
	i = p.scanner.Next()
	if state := checkErr(i.T); state != null {
		return state
	}

	// Attempt to assign cite key to the declaration
	i = p.scanner.Next()
	if state := checkErr(i.T); state != null {
		return state
	}
	if i.T != scan.ItemCiteKey {
		return err
	}
	decl.CiteKey = i.Val

	for {
		i = p.scanner.Next()
		if state := checkErr(i.T); state != null {
			return state
		}
		switch i.T {
		case scan.ItemComment:
			v := CommentExpr{Value: i.Val}
			p.comments.Values = append(p.comments.Values, &v)
		case scan.ItemFieldType:
			stmt.Key = i.Val
		case scan.ItemFieldText:
			stmt.Value = i.Val
			if !stmt.ok() {
				return err
			}
			decl.Fields = append(decl.Fields, stmt)
			stmt = &FieldStmt{}
		case scan.ItemRightDelim:
			decl.Comments = p.comments
			p.resetComms()
			p.nodes <- decl
			return null
		case scan.ItemComma, scan.ItemEqSgn: // consume
		default:
			return err
		}
	}
}

func (p *Parser) preamble() state {
	decl, ok := p.currDecl.(*PreambleDecl)
	if !ok {
		return err
	}
	var i scan.Item

	// Consume body delimiter
	i = p.scanner.Next()
	if state := checkErr(i.T); state != null {
		return state
	}

	for {
		i = p.scanner.Next()
		if state := checkErr(i.T); state != null {
			return state
		}
		switch i.T {
		case scan.ItemComment:
			v := CommentExpr{Value: i.Val}
			p.comments.Values = append(p.comments.Values, &v)
		case scan.ItemFieldText:
			decl.Value = i.Val
		case scan.ItemRightDelim:
			decl.Comments = p.comments
			p.resetComms()
			p.nodes <- decl
			return null
		default:
			return err
		}
	}
}

func (p *Parser) abbrev() state {
	decl, ok := p.currDecl.(*AbbrevDecl)
	stmt := FieldStmt{}
	if !ok {
		return err
	}

	var i scan.Item

	// Consume body delimiter
	i = p.scanner.Next()
	if state := checkErr(i.T); state != null {
		return state
	}

	for {
		i = p.scanner.Next()
		if state := checkErr(i.T); state != null {
			return state
		}
		switch i.T {
		case scan.ItemComment:
			v := CommentExpr{Value: i.Val}
			p.comments.Values = append(p.comments.Values, &v)
		case scan.ItemFieldType:
			stmt.Key = i.Val
		case scan.ItemFieldText:
			stmt.Value = i.Val
			if !stmt.ok() {
				return err
			}
			decl.Field = &stmt
		case scan.ItemRightDelim:
			decl.Comments = p.comments
			p.resetComms()
			p.nodes <- decl
			return null
		case scan.ItemEqSgn: // consume
		default:
			return err
		}
	}
}

func checkErr(t scan.ItemType) state {
	if t == scan.ItemErr {
		return err
	}
	if t == scan.ItemEOF {
		return eof
	}
	return null
}
