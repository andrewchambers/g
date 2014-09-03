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

func ws(depth uint) string {
    //Nicer way to do this?
    ret := ""
    for depth != 0 {
        ret += " "
        depth -= 1
    }
    return ret
}

type StatementList interface {
    addStatement(n Node)
}

type Node interface {
	Dump(depth uint) string
	GetSpan() FileSpan
}

type File struct {
	SpanProvider
	Pkg string
	//List of imports in the translation unit.
	Imports      []*String
	FuncDecls    []*FuncDecl
	TypeDecls    []*TypeDecl
	ConstDecls   []*ConstDecl
	VarDecls     []*VarDecl
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
	ret += ws(d + 2) + "Package:\n" 
	ret += ws(d + 4) + n.Pkg + "\n"
	ret += ws(d + 2) + "Imports:\n"
	for _,v := range n.Imports {
	    ret += ws(d + 4) + v.Val + "\n"
	} 
	ret += ws(d + 2) + "TypeDecls:\n"
	for _,v := range n.TypeDecls {
	    ret += v.Dump(d + 4)
	} 
	ret += ws(d + 2) + "ConstDecls:\n"
	ret += ws(d + 2) + "VarDecls:\n"
	ret += ws(d + 2) + "FuncDecls:\n"
	for _,f := range n.FuncDecls {
	    ret += f.Dump(d + 4)
	}
	return ret
}

type For struct {
	span             FileSpan
	init, cond, step Node
	body             []Node
}

func (n *For) GetSpan() FileSpan { return n.span }
func (n *For) Dump(depth uint) string {
	return fmt.Sprintf("(For %s %s %s %s)", n.init, n.cond, n.step, n.body)
}

type If struct {
	span FileSpan
	cond Node
	body []Node
	els  []Node
}

func (n *If) GetSpan() FileSpan { return n.span }
func (n *If) Dump(depth uint) string {
	return fmt.Sprintf("(if %s %s %s)", n.cond, n.body, n.els)
}

type Binop struct {
	span FileSpan
	op   TokenKind
	l, r Node
}

func (n *Binop) GetSpan() FileSpan { return n.span }
func (n *Binop) Dump(depth uint) string {
	return "(BINOP)"
}

type Call struct {
	span     FileSpan
	funcLike Node
	args     []Node
}

func (n *Call) GetSpan() FileSpan { return n.span }
func (n *Call) Dump(depth uint) string {
	return "(CALL)"
}

type TypeAlias struct {
	SpanProvider
	Name string
}

func (n *TypeAlias) Dump(d uint) string {
	ret := ws(d) + "TypeAlias:\n"
	ret += ws(d + 2) + n.Name + "\n"
	return ret
}


type Struct struct {
	span  FileSpan
	names []string
	types []Node
}

func (n *Struct) GetSpan() FileSpan { return n.span }
func (n *Struct) Dump(depth uint) string {
	return "(STRUCT)"
}

type Ident struct {
	SpanProvider
	Val string
}

func (n *Ident) Dump(depth uint) string {
	return n.Val
}

type Constant struct {
	SpanProvider
	Val  string
}

func (n *Constant) Dump(d uint) string {
	return ws(d) + n.Val
}

type String struct {
	SpanProvider
	Val string
}

func (n *String) Dump(d uint) string {
	return ws(d) + n.Val
}

type VarDecl struct {
	SpanProvider
	val  string
	init Node
}

func (n *VarDecl) Dump(depth uint) string {
	return "(VARDECL)"
}

type TypeDecl struct {
	SpanProvider
    Name string
	Type Node
}

func (n *TypeDecl) Dump(d uint) string {
	ret := ws(d) + "TypeDecl:\n"
	ret += ws(d + 2) + "Name: " + n.Name + "\n"
	ret += n.Type.Dump(d + 4)
	return ret
}

type ConstDecl struct {
	SpanProvider
    Name string
	Body Node
}

func (n *ConstDecl) Dump(depth uint) string {
	return "(Constdecl)"
}

type Return struct {
	SpanProvider
	Expr Node
}

func (n *Return) Dump(d uint) string {
	ret := ws(d) + "Return:\n"
	if n.Expr != nil {
	    ret += n.Expr.Dump(d + 2)
	}
	return ret
}

type FuncDecl struct {
	SpanProvider
	Name    string
	RetType Node
	ArgNames    []string
	ArgTypes    []Node
	Body    []Node
}

func (n *FuncDecl) addArgument(name string, t Node) {
    n.ArgNames = append(n.ArgNames,name)
    n.ArgTypes = append(n.ArgTypes,t)
}

func (n *FuncDecl) addStatement(s Node) {
    n.Body = append(n.Body,s)
}

func (n *FuncDecl) Dump(d uint) string {
	ret := ""
	ret = ws(d) + "FuncDecl:\n"
	ret += ws(d + 2) + "Name: " + n.Name + "\n"
	ret += ws(d + 2) + "RetType:\n"
	ret += n.RetType.Dump(d + 4)
	ret += ws(d + 2) + "Arguments:\n"
	for idx,name := range n.ArgNames {
	    ret += ws(d + 4) + name + n.ArgTypes[idx].Dump(0)
	}
	ret += ws(d + 2) + "Statements:\n"
	for _,statement := range n.Body {
	    ret += statement.Dump(d + 4)
	}
	return ret
}
