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

type Symbol struct {
    DefPos parse.FilePos
    GType GType
    SType stype
}

func newSymbol(p parse.FilePos) *Symbol {
    ret := &Symbol{}
    ret.DefPos = p
    return ret
}

type scope struct {
    parent *scope
    symkv map[string]  *Symbol
    typekv map[string] *GType
}

func newScope(parent *scope) *scope {
    s := &scope{}
    s.parent = parent
    s.symkv = make(map[string]*Symbol)
    s.typekv = make(map[string]*GType)
    return s
}

func (s *scope) declareType(k string,t *GType) (error) {
    _,ok := s.typekv[k]
    if ok {
        return fmt.Errorf("type %s already defined",k)
    }
    s.typekv[k] = t
    return nil
}


func (s *scope) declareSym(k string,sym *Symbol) error {
    v,ok := s.symkv[k]
    if ok {
        return fmt.Errorf("symbol %s already defined at %s",k,v.DefPos)
    }
    s.symkv[k] = sym
    return nil
}

func (s *scope) lookupType(k string) (*GType,error) {
    v,ok := s.typekv[k]
    if ok {
        return v,nil
    }
    if s.parent != nil {
        return s.parent.lookupType(k)
    }
    return nil,fmt.Errorf("type %s is not a valid type alias", k)
}

func (s *scope) lookupSym(k string) (*Symbol,error) {
    v,ok := s.symkv[k]
    if ok {
        return v,nil
    }
    if s.parent != nil {
        return s.parent.lookupSym(k)
    }
    return nil,fmt.Errorf("symbol %s is not declared", k)
}

