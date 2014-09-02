package sem

import (
	"fmt"
	"github.com/andrewchambers/g/parse"
)

func Process(tunit *parse.ASTTUnit, onError func (string,parse.FileSpan)) {
	state := newSemState(onError)
	state.process(tunit)
}

type semState struct {
	scope *scope
	onError func (string,parse.FileSpan)
}

func newSemState(onError func (string,parse.FileSpan)) *semState {
	ret := &semState{}
	ret.scope = newScope(nil)
	ret.onError = onError
	return ret
}

type scope struct {
	parent *scope
	kv     map[string]*Symbol
}

func newScope(parent *scope) *scope {
	ret := &scope{}
	ret.parent = parent
	ret.kv = make(map[string]*Symbol)
	return ret
}

//Returns the symbol, and true if symbol is global
func (s *scope) lookup(name string) (*Symbol, bool) {
	sym, ok := s.kv[name]
	if ok {
		return sym, s.parent == nil
	}
	if s.parent == nil {
		return nil, false
	}
	return s.parent.lookup(name)
}

func (s *scope) define(name string, sym *Symbol) error {
	_, ok := s.kv[name]
	if ok {
		return fmt.Errorf("Redefining symbol %s", name)
	}
	s.kv[name] = sym
	return nil
}

type Symbol struct {
	Name      string
	DefinedAt parse.FileSpan
	Type      GType
}

func (s *semState) pushScope() {
	s.scope = newScope(s.scope)
}

func (s *semState) popScope() {
	s.scope = s.scope.parent
	if s.scope == nil {
		panic("internal error unbalanced scopes!")
	}
}

func (state *semState) process(tunit *parse.ASTTUnit) {
	state.defineGlobalSymbols(tunit)
	for _, v := range tunit.Body {
		switch v := v.(type) {
		case *parse.ASTFuncDecl:
			_ = state.emitSemFunctionTree(v)
		}
	}
}

func (state *semState) defineGlobalSymbols(tunit *parse.ASTTUnit) {
	for _, v := range tunit.Body {
		switch v := v.(type) {
		case *parse.ASTFuncDecl:
			sym := &Symbol{}
			sym.DefinedAt = v.GetSpan()
			sym.Name = v.Name
			t,err := astNodeToGType(v)
			sym.Type = t
			if err != nil {
			    state.onError(err.Error(),v.Span)
			}
			err = state.scope.define(sym.Name, sym)
			if err != nil {
				state.onError(err.Error(),v.Span)
			}
		case *parse.ASTVarDecl:
			panic("unimplemented var global")
		default:
		}
	}

}
func astNodeToGType(n parse.ASTNode) (GType,error) {
	switch v := n.(type) {
	case *parse.ASTIdent:
		switch {
		case v.Val == "int":
			ret := &GInt{}
			ret.Bits = 64
			ret.Signed = true
			return ret,nil
		default:
			return nil,fmt.Errorf("%s is not a valid type",v.Val)
		}
	case *parse.ASTFuncDecl:
		ret := &GFunc{}
		t,err := astNodeToGType(n)
		ret.RetType = t
		if err != nil {
		    return nil,err
		}
		return ret,nil
	default:
		panic("unimplemented error")
	}
	panic("unreachable")
}

func (state *semState) emitSemFunctionTree(ast *parse.ASTFuncDecl) *SemFunction {
	f := &SemFunction{}
	for _, statement := range ast.Body {
		semStatement := state.emitSemStatement(statement)
		f.Statements = append(f.Statements, semStatement)
	}
	return f
}

func (state *semState) emitSemStatement(statement parse.ASTNode) SemNode {
	var ret SemNode = nil
	switch n := statement.(type) {
	case *parse.ASTCall:
		ret = state.emitSemCall(n)
	default:
		panic("unimplemented case or error...")
	}
	return ret
}

func (state *semState) emitSemCall(ast parse.ASTNode) SemNode {
	scall := &SemCall{}
	return scall
}
