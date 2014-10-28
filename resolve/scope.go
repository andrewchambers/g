package resolve

import (
	"fmt"
	//"github.com/andrewchambers/g/parse"
)

type scope interface {
	declareSym(k string, s symbol) error
}

type packageScope struct {
	unresolved map[string][]lazySymbol
	symkv      map[string]symbol
}

type localScope struct {
	parent scope
	symkv  map[string]symbol
}

func newLocalScope(parent scope) scope {
	s := &localScope{}
	s.parent = parent
	s.symkv = make(map[string]symbol)
	return s
}

func (s *localScope) declareSym(k string, sym symbol) error {
	_, ok := s.symkv[k]
	if ok {
		return fmt.Errorf("ident %s already defined", k)
	}
	s.symkv[k] = sym
	return nil
}

/*
func (s *localScope) lookupSym(k string) (symbol, error) {
	v, ok := s.symkv[k]
	if ok {
		return v, nil
	}
	if s.parent != nil {
		return s.parent.lookupSym(k)
	}
	return nil, fmt.Errorf("ident %s is not declared", k)
}
*/
