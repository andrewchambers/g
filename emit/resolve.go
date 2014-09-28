package emit

import (
    "fmt"
	"github.com/andrewchambers/g/parse"
)


func (e *emitter) resolveSymbols(n parse.Node) error {
	switch n := n.(type) {
	case *parse.FuncDecl:
		e.pushScope()
		err := e.handleFuncPrologue(n)
		if err != nil {
			return err
		}
		e.pushScope()
		for _, subn := range n.Body {
			err := e.resolveSymbols(subn)
			if err != nil {
				return err
			}
		}
		e.popScope()
		e.popScope()
	case *parse.VarDecl:
		err := e.resolveLocalVarDecl(n)
		if err != nil {
			return err
		}
	case *parse.IndexInto:
	    err := e.resolveSymbols(n.Expr)
		if err != nil {
			return err
		}
	    err = e.resolveSymbols(n.Index)
		if err != nil {
			return err
		}
	case *parse.Call:
		err := e.resolveSymbols(n.FuncLike)
		if err != nil {
			return err
		}
		for _, arg := range n.Args {
			err = e.resolveSymbols(arg)
			if err != nil {
				return err
			}
		}
	case *parse.Binop:
		err := e.resolveSymbols(n.L)
		if err != nil {
			return err
		}
		err = e.resolveSymbols(n.R)
		if err != nil {
			return err
		}
	case *parse.Unop:
		err := e.resolveSymbols(n.Expr)
		if err != nil {
			return err
		}
	case *parse.Assign:
		err := e.resolveSymbols(n.L)
		if err != nil {
			return err
		}
		err = e.resolveSymbols(n.R)
		if err != nil {
			return err
		}
	case *parse.If:
		err := e.resolveSymbols(n.Cond)
		if err != nil {
			return err
		}
		for _, subn := range n.Body {
			err := e.resolveSymbols(subn)
			if err != nil {
				return err
			}
		}
		for _, subn := range n.Els {
			err := e.resolveSymbols(subn)
			if err != nil {
				return err
			}
		}
	case *parse.PointerTo:
		err := e.resolveSymbols(n.PointsTo)
		if err != nil {
			return err
		}
	case *parse.For:
		if n.Init != nil {
			err := e.resolveSymbols(n.Init)
			if err != nil {
				return err
			}
		}
		if n.Cond != nil {
			err := e.resolveSymbols(n.Cond)
			if err != nil {
				return err
			}
		}
		if n.Step != nil {
			err := e.resolveSymbols(n.Step)
			if err != nil {
				return err
			}
		}
		for _, subn := range n.Body {
			err := e.resolveSymbols(subn)
			if err != nil {
				return err
			}
		}
	case *parse.Return:
		if n.Expr != nil {
			err := e.resolveSymbols(n.Expr)
			if err != nil {
				return err
			}
		}
	case *parse.Ident:
		sym, err := e.curscope.lookupSym(n.Val)
		if err != nil {
			return fmt.Errorf("%s at %s:%s", err, n.Span.Path, n.Span.Start)
		}
		e.symbolMap[n] = sym
	case *parse.ExpressionStatement:
		err := e.resolveSymbols(n.Expr)
		if err != nil {
			return err
		}
	case *parse.EmptyStatement, *parse.Constant:
		//nothing
	default:
		panic(n)
	}
	return nil
}


func (e *emitter) resolveLocalVarDecl(vd *parse.VarDecl) error {
	t, err := e.parseNodeToGType(vd.Type)
	if err != nil {
		return err
	}
	name := e.newLLVMName()
	e.emiti("%s = alloca %s\n", name, gTypeToLLVM(t))
	err = e.emitZeroMem(name, t)
	if err != nil {
		return err
	}
	s := &localSymbol{
		alloca: name,
		gType:  t,
		defPos: vd.Span.Start,
	}
	err = e.curscope.declareSym(vd.Name, s)
	if err != nil {
		return err
	}
	if vd.Init != nil {
		err = e.resolveSymbols(vd.Init)
	}
	return err
}


