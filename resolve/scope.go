package resolve

import (
	"fmt"
	//"github.com/andrewchambers/g/parse"
)

type scope interface {
	declareSym(k string, s symbol) error
	lookupSym(k string) (symbol, error)
}

type packageScope struct {
	unresolved map[string]*lazySymbol
	symkv      map[string]symbol
}

type localScope struct {
	parent scope
	symkv  map[string]symbol
}

func newPackageScope() *packageScope {
	return &packageScope{
		unresolved: make(map[string]*lazySymbol),
		symkv:      make(map[string]symbol),
	}
}

func (s *packageScope) declareSym(k string, sym symbol) error {
	_, ok := s.symkv[k]
	if ok {
		_, ok = s.unresolved[k]
		if !ok {
			return fmt.Errorf("redefinition of %s", k)
		}
	}
	s.symkv[k] = sym
	// Patch all the unresolved symbols.
	lazy, ok := s.unresolved[k]
	if ok {
		lazy.s = sym
		delete(s.unresolved, k)
	}

	return nil
}

func (s *packageScope) lookupSym(k string) (symbol, error) {
	sym, ok := s.symkv[k]
	if ok {
		return sym, nil
	}
	ret := &lazySymbol{}
	s.symkv[k] = ret
	s.unresolved[k] = ret
	return ret, nil
}

func newLocalScope(parent scope) *localScope {
	s := &localScope{}
	s.parent = parent
	s.symkv = make(map[string]symbol)
	return s
}

func (s *localScope) declareSym(k string, sym symbol) error {
	_, ok := s.symkv[k]
	if ok {
		return fmt.Errorf("redefinition of %s", k)
	}
	s.symkv[k] = sym
	return nil
}

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
