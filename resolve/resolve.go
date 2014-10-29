package resolve

import (
	"github.com/andrewchambers/g/parse"
)

// Resolver walks the package AST and resolves all symbols to either
// types, global variables, local variables, or constants.

type Resolver struct {
	ps *packageScope
	ls *localScope
	kv map[*parse.Ident]symbol
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
		r.resolveFuncBodyNode(n)
	}

	if r.ls != funcScope {
		panic("internal error")
	}
}

// Walk the function tree handling scopes and definitions while mapping ident
// nodes to symbol objects.

func (r *Resolver) resolveFuncBodyNode(n parse.Node) {

	if n == nil {
		return
	}

	switch n := n.(type) {
	case *parse.VarDecl:
		err := r.ls.declareSym(n.Name, &localSymbol{n})
		if err != nil {
			panic(err)
		}
	case *parse.Ident:
		_, err := r.ls.lookupSym(n.Val)
		if err != nil {
			panic(err)
		}
	case *parse.Assign:
		r.resolveFuncBodyNode(n.L)
		r.resolveFuncBodyNode(n.R)
	case *parse.Binop:
		r.resolveFuncBodyNode(n.L)
		r.resolveFuncBodyNode(n.R)
	case *parse.Unop:
		r.resolveFuncBodyNode(n.Expr)
	case *parse.For:
		r.pushScope()
		r.resolveFuncBodyNode(n.Init)
		r.resolveFuncBodyNode(n.Cond)
		r.resolveFuncBodyNode(n.Step)
		for _, sub := range n.Body {
			r.resolveFuncBodyNode(sub)
		}
		r.popScope()
	case *parse.If:
		r.resolveFuncBodyNode(n.Cond)
		r.pushScope()
		for _, sub := range n.Body {
			r.resolveFuncBodyNode(sub)
		}
		r.popScope()
		r.pushScope()
		for _, sub := range n.Els {
			r.resolveFuncBodyNode(sub)
		}
		r.popScope()
	case *parse.ExpressionStatement:
		r.resolveFuncBodyNode(n.Expr)
	case *parse.Call:
		r.resolveFuncBodyNode(n.FuncLike)
		for _, arg := range n.Args {
			r.resolveFuncBodyNode(arg)
		}
	default:
		panic(n)
	}
}
