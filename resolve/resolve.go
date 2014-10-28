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
	return ret
}

func (r *Resolver) pushScope() {

}

func (r *Resolver) popScope() {

}

func (r *Resolver) ResolvePackage(files []*parse.File) {

	r.resolvePackageScope(files)

	for _, f := range files {
		for _, fd := range f.FuncDecls {
			r.pushScope()
			// XXX cache file scope from before?
			r.resolveImports(f)
			r.resolveFuncDecl(fd)
			r.popScope()
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

	/*
	   if len(r.ps.unresolvedSymbols) != 0 {
	       panic("unresolved symbols...")
	   }
	*/

}

func (r *Resolver) resolveImports(f *parse.File) {

}

func (r *Resolver) resolvePackageLevel(f *parse.File) {
	for _, _ = range f.TypeDecls {
		//ty := r.astNodeToGType(td.Type)
		//r.ps.DeclareSym(td.Name,nil)
	}

	/*
	   for _, fd := range f.FuncDecls {
	       ty := r.astNodeToGType(fd)
	       r.ps.DeclareSym(fd.Name,nil)
	   }
	*/
}

func (r *Resolver) resolveFuncDecl(fd *parse.FuncDecl) {

}

func (r *Resolver) astNodeToGType(n parse.Node) GType {
	return nil
}
