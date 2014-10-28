package parse

import (
	"fmt"
	"io"
)

type Node interface {
	GetSpan() FileSpan
}

type File struct {
	SpanProvider
	Pkg string
	//List of imports in the translation unit.
	Imports    []*String
	FuncDecls  []*FuncDecl
	TypeDecls  []*TypeDecl
	ConstDecls []*ConstDecl
	VarDecls   []*VarDecl
}

type For struct {
	SpanProvider
	Init, Cond, Step Node
	Body             []Node
}

type If struct {
	SpanProvider
	Cond Node
	Body []Node
	Els  []Node
}

type Selector struct {
	SpanProvider
	Name string
	Expr Node
}

type Unop struct {
	SpanProvider
	Op   TokenKind
	Expr Node
}

type Call struct {
	SpanProvider
	FuncLike Node
	Args     []Node
}

type PointerTo struct {
	SpanProvider
	PointsTo Node
}

type ArrayOf struct {
	SpanProvider
	Dim     uint
	SubType Node
}

type IndexInto struct {
	SpanProvider
	Index Node
	Expr  Node
}

type Struct struct {
	SpanProvider
	Names []string
	Types []Node
}

type Binop struct {
	SpanProvider
	Op   TokenKind
	L, R Node
}

type EmptyStatement struct {
	SpanProvider
}

type ExpressionStatement struct {
	SpanProvider
	Expr Node
}

type Assign struct {
	SpanProvider
	Op   TokenKind
	L, R Node
}

type Constant struct {
	SpanProvider
	Val int64
}

type Initializer struct {
	SpanProvider
	Sub []Node
}

type Ident struct {
	SpanProvider
	Val string
}

type String struct {
	SpanProvider
	Val string
}

type VarDecl struct {
	SpanProvider
	Name string
	Type Node
	Init *Assign
}

type TypeDecl struct {
	SpanProvider
	Name string
	Type Node
}

type ConstDecl struct {
	SpanProvider
	Name string
	Body Node
}

type FuncDecl struct {
	SpanProvider
	Name     string
	RetType  Node
	isVarArg bool
	ArgNames []string
	ArgTypes []Node
	Body     []Node
}

type Return struct {
	SpanProvider
	Expr Node
}

type SpanProvider struct {
	// The file span of the token.
	Span FileSpan
}

func (s *SpanProvider) GetSpan() FileSpan {
	return s.Span
}

func ws(depth uint) string {
	//Nicer way to do this?
	ret := ""
	for depth != 0 {
		ret += " "
		depth -= 1
	}
	return ret
}
func (n *File) addImport(s *String) {
	n.Imports = append(n.Imports, s)
}

func (n *File) addFuncDecl(f *FuncDecl) {
	n.FuncDecls = append(n.FuncDecls, f)
}

func (n *File) addTypeDecl(t *TypeDecl) {
	n.TypeDecls = append(n.TypeDecls, t)
}

func (n *File) addConstDecl(c *ConstDecl) {
	n.ConstDecls = append(n.ConstDecls, c)
}

func (n *File) addVarDecl(v *VarDecl) {
	n.VarDecls = append(n.VarDecls, v)
}

func (n *FuncDecl) addArgument(name string, t Node) {
	n.ArgNames = append(n.ArgNames, name)
	n.ArgTypes = append(n.ArgTypes, t)
}

func (n *FuncDecl) addStatement(s Node) {
	n.Body = append(n.Body, s)
}

func DebugDump(w io.Writer, n Node) {
	debugDump(0, w, n)
}

func debugDump(d int, w io.Writer, n Node) {

	// XXX
	ws := func(c int) string {
		ret := ""
		for i := 0; i < c; i++ {
			ret += " "
		}
		return ret
	}

	p := func(d int, format string, args ...interface{}) {
		fmt.Fprintf(w, ws(d)+format, args...)
	}

	switch n := n.(type) {
	case *File:
		p(d+0, "File:\n")
		p(d+2, "TypeDecls:\n")
		for _, td := range n.TypeDecls {
			debugDump(d+4, w, td)
		}
		p(d+2, "FuncDecls:\n")
		for _, fd := range n.FuncDecls {
			debugDump(d+4, w, fd)
		}
	case *FuncDecl:
		p(d+0, "FuncDecl:\n")
		p(d+2, "Name: %s\n", n.Name)
		p(d+2, "Body:\n")
		for _, n := range n.Body {
			debugDump(d+4, w, n)
		}
	case *TypeDecl:
		p(d+0, "TypeDecl:\n")
		p(d+2, "Name:\n")
		p(d+4, "%s\n", n.Name)
		p(d+2, "Type:\n")
		debugDump(d+4, w, n.Type)
	case *Struct:
		p(d+0, "Struct:\n")
		for idx := range n.Names {
			p(d+2, "Member: %s\n", n.Names[idx])
			debugDump(d+4, w, n.Types[idx])
		}
	case *PointerTo:
		p(d+0, "PointerTo:\n")
		debugDump(d+2, w, n.PointsTo)
	case *Ident:
		p(d+0, "Ident: %s\n", n.Val)
	case *ExpressionStatement:
		p(d+0, "ExpressionStatement:\n")
		debugDump(d+2, w, n.Expr)
	case *Unop:
		p(d+0, "Unop: %s\n", n.Op)
		debugDump(d+2, w, n.Expr)
	case *Call:
		p(d+0, "Call:\n")
	default:
		p(d+0, "unhandled: %T\n", n)
	}
}
