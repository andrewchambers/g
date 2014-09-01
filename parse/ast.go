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

type ASTNode interface {
	String() string
	GetSpan() FileSpan
}

type ASTTUnit struct {
	SpanProvider
	Pkg string
	//List of imports in the translation unit.
	Imports []*Token
	Body    []ASTNode
}

func (n *ASTTUnit) addImport(t *Token) {
	if t.Kind != STRING_LITERAL {
		panic("internal error!")
	}
	n.Imports = append(n.Imports, t)
}

func (n *ASTTUnit) String() string {
	return fmt.Sprintf("(TUnit)")
}

type ASTFor struct {
	span             FileSpan
	init, cond, step ASTNode
	body             []ASTNode
}

func (n *ASTFor) GetSpan() FileSpan { return n.span }
func (n *ASTFor) String() string {
	return fmt.Sprintf("(for %s %s %s %s)", n.init, n.cond, n.step, n.body)
}

type ASTIf struct {
	span FileSpan
	cond ASTNode
	body []ASTNode
	els  []ASTNode
}

func (n *ASTIf) GetSpan() FileSpan { return n.span }
func (n *ASTIf) String() string {
	return fmt.Sprintf("(if %s %s %s)", n.cond, n.body, n.els)
}

type ASTBinop struct {
	span FileSpan
	op   TokenKind
	l, r ASTNode
}

func (n *ASTBinop) GetSpan() FileSpan { return n.span }
func (n *ASTBinop) String() string {
	return "(BINOP)"
}

type ASTCall struct {
	span     FileSpan
	funcLike ASTNode
	args     []ASTNode
}

func (n *ASTCall) GetSpan() FileSpan { return n.span }
func (n *ASTCall) String() string {
	return "(CALL)"
}

type ASTStruct struct {
	span  FileSpan
	names []string
	types []ASTNode
}

func (n *ASTStruct) GetSpan() FileSpan { return n.span }
func (n *ASTStruct) String() string {
	return "(STRUCT)"
}

type ASTIdent struct {
	SpanProvider
	Val string
}

func (n *ASTIdent) String() string {
	return n.Val
}

type ASTConstant struct {
	span FileSpan
	val  string
}

func (n *ASTConstant) GetSpan() FileSpan { return n.span }
func (n *ASTConstant) String() string {
	return n.val
}

type ASTString struct {
	SpanProvider
	val string
}

func (n *ASTString) String() string {
	return n.val
}

type ASTVarDecl struct {
	SpanProvider
	val  string
	init ASTNode
}

func (n *ASTVarDecl) String() string {
	return "(VARDECL)"
}

type ASTFuncDecl struct {
	SpanProvider
	Name    string
	RetType ASTNode
	Args    []ASTNode
	Body    []ASTNode
}

func (n *ASTFuncDecl) String() string {
	return "(FUNCDECL)"
}
