package emit

import (
	"fmt"
	"github.com/andrewchambers/g/parse"
)

type stype int

const (
	LOCAL = iota
	GLOBAL
	ARG
	PACKAGE
)

type symbol interface {
	getDefPos() parse.FilePos
	getGType() GType
}

type localSymbol struct {
	gType  GType
	defPos parse.FilePos
}

type constSymbol struct {
	v      Value
	defPos *parse.FilePos
}

type globalSymbol struct {
	gType  GType
	defPos parse.FilePos
}

type scope struct {
	parent *scope
	symkv  map[string]symbol
	typekv map[string]GType
}

func newGlobalSymbol(p parse.FilePos) symbol {
	ret := &globalSymbol{}
	ret.defPos = p
	return ret
}

func newGlobalFuncSymbol(ty *GFunc, p parse.FilePos) symbol {
	ret := &globalSymbol{}
	ret.defPos = p
	ret.gType = ty
	return ret
}

func (g *globalSymbol) getGType() GType {
	return g.gType
}

func (g *globalSymbol) getDefPos() parse.FilePos {
	return g.defPos
}

func (g *constSymbol) getGType() GType {
	return &GConstant{}
}

func (g *constSymbol) getDefPos() parse.FilePos {
	return *g.defPos
}

func (l *localSymbol) getGType() GType {
	return l.gType
}

func (l *localSymbol) getDefPos() parse.FilePos {
	return l.defPos
}

func newScope(parent *scope) *scope {
	s := &scope{}
	s.parent = parent
	s.symkv = make(map[string]symbol)
	s.typekv = make(map[string]GType)
	return s
}

func (s *scope) declareType(k string, t GType) error {
	_, ok := s.typekv[k]
	if ok {
		return fmt.Errorf("type %s already defined", k)
	}
	s.typekv[k] = t
	return nil
}

func (s *scope) declareSym(k string, sym symbol) error {
	v, ok := s.symkv[k]
	if ok {
		return fmt.Errorf("ident %s already defined at %s", k, v.getDefPos())
	}
	s.symkv[k] = sym
	return nil
}

func (s *scope) lookupType(k string) (GType, error) {
	v, ok := s.typekv[k]
	if ok {
		return v, nil
	}
	if s.parent != nil {
		return s.parent.lookupType(k)
	}
	return nil, fmt.Errorf("%s is not a valid type alias", k)
}

func (s *scope) lookupSym(k string) (symbol, error) {
	v, ok := s.symkv[k]
	if ok {
		return v, nil
	}
	if s.parent != nil {
		return s.parent.lookupSym(k)
	}
	return nil, fmt.Errorf("ident %s is not declared", k)
}
