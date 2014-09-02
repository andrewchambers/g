package parse

import (
	"fmt"
)

type SpanProvider struct {
	// The file span of the token.
	Span FileSpan
}

func (s *SpanProvider) GetSpan() FileSpan {
	return s.Span
}

type Node interface {
	String() string
	GetSpan() FileSpan
}

type File struct {
	SpanProvider
	Pkg string
	//List of imports in the translation unit.
	Imports []*Token
	Body    []Node
}

func (n *File) addImport(t *Token) {
	if t.Kind != STRING_LITERAL {
		panic("internal error!")
	}
	n.Imports = append(n.Imports, t)
}

func (n *File) String() string {
	return fmt.Sprintf("(TUnit)")
}

type For struct {
	span             FileSpan
	init, cond, step Node
	body             []Node
}

func (n *For) GetSpan() FileSpan { return n.span }
func (n *For) String() string {
	return fmt.Sprintf("(For %s %s %s %s)", n.init, n.cond, n.step, n.body)
}

type If struct {
	span FileSpan
	cond Node
	body []Node
	els  []Node
}

func (n *If) GetSpan() FileSpan { return n.span }
func (n *If) String() string {
	return fmt.Sprintf("(if %s %s %s)", n.cond, n.body, n.els)
}

type Binop struct {
	span FileSpan
	op   TokenKind
	l, r Node
}

func (n *Binop) GetSpan() FileSpan { return n.span }
func (n *Binop) String() string {
	return "(BINOP)"
}

type Call struct {
	span     FileSpan
	funcLike Node
	args     []Node
}

func (n *Call) GetSpan() FileSpan { return n.span }
func (n *Call) String() string {
	return "(CALL)"
}

type Struct struct {
	span  FileSpan
	names []string
	types []Node
}

func (n *Struct) GetSpan() FileSpan { return n.span }
func (n *Struct) String() string {
	return "(STRUCT)"
}

type Ident struct {
	SpanProvider
	Val string
}

func (n *Ident) String() string {
	return n.Val
}

type Constant struct {
	span FileSpan
	val  string
}

func (n *Constant) GetSpan() FileSpan { return n.span }
func (n *Constant) String() string {
	return n.val
}

type String struct {
	SpanProvider
	val string
}

func (n *String) String() string {
	return n.val
}

type VarDecl struct {
	SpanProvider
	val  string
	init Node
}

func (n *VarDecl) String() string {
	return "(VARDECL)"
}

type FuncDecl struct {
	SpanProvider
	Name    string
	RetType Node
	Args    []Node
	Body    []Node
}

func (n *FuncDecl) String() string {
	return "(FUNCDECL)"
}
