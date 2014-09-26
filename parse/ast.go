package parse

import (
	"fmt"
)

type Node interface {
	Dump(depth uint) string
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

type TypeAlias struct {
	SpanProvider
	Name string
}

type PointerTo struct {
	SpanProvider
	PointsTo Node
}

type ArrayOf struct {
	SpanProvider
	Dim uint
	SubType Node
}

type Struct struct {
	SpanProvider
	names []string
	types []Node
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

func (n *File) Dump(d uint) string {
	ret := ws(d) + "File:\n"
	ret += ws(d+2) + "Package:\n"
	ret += ws(d+4) + n.Pkg + "\n"
	ret += ws(d+2) + "Imports:\n"
	for _, v := range n.Imports {
		ret += ws(d+4) + v.Val + "\n"
	}
	ret += ws(d+2) + "TypeDecls:\n"
	for _, v := range n.TypeDecls {
		ret += v.Dump(d + 4)
	}
	ret += ws(d+2) + "ConstDecls:\n"
	ret += ws(d+2) + "VarDecls:\n"
	for _, v := range n.VarDecls {
		ret += v.Dump(d + 4)
	}
	ret += ws(d+2) + "FuncDecls:\n"
	for _, f := range n.FuncDecls {
		ret += f.Dump(d + 4)
	}
	return ret
}

func (n *EmptyStatement) Dump(d uint) string {
	return ws(d) + "empty;\n"
}

func (n *ExpressionStatement) Dump(d uint) string {
	ret := ws(d) + "ExpressionStatement:\n"
	ret += n.Expr.Dump(d + 2)
	return ret
}

func (n *For) Dump(d uint) string {
	return fmt.Sprintf(ws(d) + "(FOR)\n")
}

func (n *If) Dump(d uint) string {
	ret := ws(d) + "If:\n"
	ret += ws(d+2) + "Cond:\n"
	ret += n.Cond.Dump(d + 4)
	ret += ws(d+2) + "Body:\n"
	for _, stmt := range n.Body {
		ret += stmt.Dump(d + 4)
	}
	if n.Els != nil {
		for _, stmt := range n.Els {
			ret += stmt.Dump(d + 4)
		}
	}
	return ret
}

func (n *Binop) Dump(d uint) string {
	ret := ws(d) + fmt.Sprintf("Binop %s:\n", n.Op)
	ret += n.L.Dump(d + 2)
	ret += n.R.Dump(d + 2)
	return ret
}

func (n *Assign) Dump(d uint) string {
	ret := ws(d) + fmt.Sprintf("Assign %s:\n", n.Op)
	ret += n.L.Dump(d + 2)
	ret += n.R.Dump(d + 2)
	return ret
}

func (n *Unop) Dump(d uint) string {
	return "(UNOP)"
}

func (n *Selector) Dump(d uint) string {
	ret := ws(d) + "Selector:\n"
	ret += n.Expr.Dump(d + 2)
	ret += ws(d+2) + n.Name + "\n"
	return ret
}

func (n *Call) Dump(d uint) string {
	ret := ws(d) + "Call:\n"
	ret += ws(d+2) + "FuncLike:\n"
	ret += n.FuncLike.Dump(d + 4)
	ret += ws(d+2) + "Args:\n"
	for _, v := range n.Args {
		ret += v.Dump(d + 4)
	}
	return ret
}

func (n *TypeAlias) Dump(d uint) string {
	ret := ws(d) + "TypeAlias:\n"
	ret += ws(d+2) + n.Name + "\n"
	return ret
}

func (n *Struct) Dump(depth uint) string {
	return "(STRUCT)"
}

func (n *PointerTo) Dump(d uint) string {
	return "(*)"
}

func (n *ArrayOf) Dump(d uint) string {
	return "([])"
}


func (n *Ident) Dump(d uint) string {
	return ws(d) + n.Val + "\n"
}

func (n *Constant) Dump(d uint) string {
	return ws(d) + fmt.Sprintf("%v\n", n.Val)
}

func (n *String) Dump(d uint) string {
	return ws(d) + n.Val + "\n"
}

func (n *VarDecl) Dump(d uint) string {
	ret := ws(d) + "VarDecl:\n"
	ret += ws(d+2) + "Name:\n"
	ret += ws(d+4) + n.Name + "\n"
	ret += ws(d+2) + "Type:\n"
	ret += n.Type.Dump(d + 4)
	if n.Init != nil {
		ret += ws(d+2) + "Init:\n"
		ret += n.Init.Dump(d + 4)
	}
	return ret
}

func (n *TypeDecl) Dump(d uint) string {
	ret := ws(d) + "TypeDecl:\n"
	ret += ws(d+2) + "Name: " + n.Name + "\n"
	ret += n.Type.Dump(d + 4)
	return ret
}

func (n *ConstDecl) Dump(depth uint) string {
	return "(Constdecl)"
}

func (n *Return) Dump(d uint) string {
	ret := ws(d) + "Return:\n"
	if n.Expr != nil {
		ret += n.Expr.Dump(d + 2)
	}
	return ret
}

func (n *FuncDecl) addArgument(name string, t Node) {
	n.ArgNames = append(n.ArgNames, name)
	n.ArgTypes = append(n.ArgTypes, t)
}

func (n *FuncDecl) addStatement(s Node) {
	n.Body = append(n.Body, s)
}

func (n *FuncDecl) Dump(d uint) string {
	ret := ""
	ret = ws(d) + "FuncDecl:\n"
	ret += ws(d+2) + "Name:\n"
	ret += ws(d+4) + n.Name + "\n"
	ret += ws(d+2) + "RetType:\n"
	ret += n.RetType.Dump(d + 4)
	ret += ws(d+2) + "Arguments:\n"
	for idx, name := range n.ArgNames {
		ret += ws(d+4) + name + n.ArgTypes[idx].Dump(0)
	}
	ret += ws(d+2) + "Statements:\n"
	for _, statement := range n.Body {
		ret += statement.Dump(d + 4)
	}
	return ret
}
