package resolve

import (
	"github.com/andrewchambers/g/parse"
)

// Resolver walks the package AST and resolves all symbols to either
// types, global variables, local variables, or constants.

type Resolver struct {
	ps *packageScope
	ls *localScope
}

func New() *Resolver {
	ret := &Resolver{}
	ret.ps = newPackageScope()
	return ret
}

func (r *Resolver) ResolvePackage(files []*parse.File) {

	r.resolvePackageScope(files)

	for _, f := range files {
		for _, fd := range f.FuncDecls {
			r.resolveFuncDecl(fd)
		}
	}
}

func (r *Resolver) resolvePackageScope(files []*parse.File) {
	for _, f := range files {
		// Push the file scope
		r.pushScope()
		r.resolveImports(f)
		r.resolvePackageLevel(f)
		r.popScope()
	}

	if len(r.ps.unresolved) != 0 {
		panic("unresolved symbols...")
	}
}

func (r *Resolver) resolveImports(f *parse.File) {

}

func (r *Resolver) resolvePackageLevel(f *parse.File) {
	for _, td := range f.TypeDecls {
		r.resolveType(r.ps, td.Type)
		ts := &typeSymbol{td}
		err := r.ps.declareSym(td.Name, ts)
		if err != nil {
			panic(err)
		}
	}

	for _, fd := range f.FuncDecls {
		fs := &funcSymbol{fd}
		err := r.ps.declareSym(fd.Name, fs)
		if err != nil {
			panic(err)
		}
	}
}

func (r *Resolver) resolveType(sc scope, n parse.Node) {
	switch n := n.(type) {
	case *parse.Ident:
		_, err := sc.lookupSym(n.Val)
		if err != nil {
			panic(err)
		}
	case *parse.Struct:
		for idx := range n.Types {
			r.resolveType(sc, n.Types[idx])
		}
	case *parse.PointerTo:
		r.resolveType(sc, n.PointsTo)
	default:
		panic(n)
	}
}

func (r *Resolver) pushScope() {
	r.ls = newLocalScope(r.ls)
}

func (r *Resolver) popScope() {
	r.ls = r.ls.parent.(*localScope)
}

func (r *Resolver) resolveFuncDecl(fd *parse.FuncDecl) {
	funcScope := newLocalScope(r.ps)
	r.ls = funcScope

	for _, n := range fd.Body {
		r.resolveFuncNode(n)
	}

	if r.ls != funcScope {
		panic("internal error")
	}
}

func (r *Resolver) resolveFuncNode(n parse.Node) {
	switch n := n.(type) {
	case *parse.VarDecl:
	case *parse.Assign:
	case *parse.Binop:
	case *parse.Unop:
	case *parse.For:
	case *parse.If:
	case *parse.ExpressionStatement:
	case *parse.Call:
	default:
		panic(n)
	}
}
