package parse

import (
	"fmt"
)

type ASTNode interface {
	String() string
	GetSpan() FileSpan
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
	span FileSpan
	val  string
}

func (n *ASTIdent) GetSpan() FileSpan { return n.span }
func (n *ASTIdent) String() string {
	return n.val
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
	span FileSpan
	val  string
}

func (n *ASTString) GetSpan() FileSpan { return n.span }
func (n *ASTString) String() string {
	return n.val
}

type ASTVarDecl struct {
	span FileSpan
	val  string
	init ASTNode
}

func (n *ASTVarDecl) GetSpan() FileSpan { return n.span }
func (n *ASTVarDecl) String() string {
	return "(VARDECL)"
}

type ASTFuncDecl struct {
	span    FileSpan
	name    string
	retType ASTNode
	args    []ASTNode
	body    []ASTNode
}

func (n *ASTFuncDecl) GetSpan() FileSpan { return n.span }
func (n *ASTFuncDecl) String() string {
	return "(FUNCDECL)"
}
