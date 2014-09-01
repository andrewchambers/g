package sem

import (
	"github.com/andrewchambers/g/parse"
)

func Process(tUnit *parse.ASTTUnit) {
	funcDecls := gatherFuncDecls(tUnit)
	for _, v := range funcDecls {
		_ = funcDeclToGType(v)
	}
}

func funcDeclToGType(decl *parse.ASTFuncDecl) GType {
	ret := &GFunc{}
	ret.RetType = astNodeToGType(decl)
	return ret
}

func astNodeToGType(n parse.ASTNode) GType {
	switch v := n.(type) {
	case *parse.ASTIdent:
		switch {
		case v.Val == "int":
			ret := &GInt{}
			ret.Bits = 64
			ret.Signed = true
			return ret
		default:
			panic("unimplemented error")
		}
	default:
		panic("unimplemented error")
	}
	panic("unreachable")
}

func gatherFuncDecls(tunit *parse.ASTTUnit) []*parse.ASTFuncDecl {
	var ret []*parse.ASTFuncDecl
	for _, v := range tunit.Body {
		switch f := v.(type) {
		case *parse.ASTFuncDecl:
			ret = append(ret, f)
		default:
		}
	}
	return ret
}
