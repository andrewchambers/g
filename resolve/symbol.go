package resolve

import (
	"github.com/andrewchambers/g/parse"
)

const (
	LOCAL = iota
	GLOBAL
	ARG
	PACKAGE
)

type symbol interface{}

type lazySymbol struct {
	s symbol
}

type typeSymbol struct {
	Decl *parse.TypeDecl
}

type localSymbol struct {
	Decl *parse.VarDecl
}

type constSymbol struct {
}

type globalSymbol struct {
}

type funcSymbol struct {
	Decl *parse.FuncDecl
}
