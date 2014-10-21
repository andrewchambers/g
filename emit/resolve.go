package emit

import (
	"fmt"
	"github.com/andrewchambers/g/parse"
)

type symbolResolver struct {
	gscope    *scope
	curscope  *scope
	symbolMap map[parse.Node]symbol
}

//XXX rename all variables from e

func resolveASTSymbols(n parse.Node) (map[parse.Node]symbol, error) {
	sr := &symbolResolver{}
	sr.gscope = newScope(nil)
	sr.curscope = sr.gscope

	sr.addBuiltinTypes()

	err := sr.resolveSymbols(n)
	if err != nil {
		return make(map[parse.Node]symbol), err
	}
	return sr.symbolMap, nil
}

func (e *symbolResolver) addBuiltinTypes() {
	// Built in types
	e.curscope.declareType("bool", builtinBoolGType)
	e.curscope.declareType("int", getDefaultIntType(e.machine))
	e.curscope.declareType("int8", builtinInt8GType)
	e.curscope.declareType("int16", builtinInt16GType)
	e.curscope.declareType("int32", builtinInt32GType)
	e.curscope.declareType("int64", builtinInt64GType)
	e.curscope.declareType("uint", builtinUInt64GType)
	e.curscope.declareType("uint8", builtinUInt8GType)
	e.curscope.declareType("uint16", builtinUInt16GType)
	e.curscope.declareType("uint32", builtinUInt32GType)
	e.curscope.declareType("uint64", builtinUInt64GType)
	// Built in symbols
	e.curscope.declareSym("true", &constSymbol{&boolConstant{true}, nil})
	e.curscope.declareSym("false", &constSymbol{&boolConstant{true}, nil})
}

func (e *symbolResolver) pushScope() {
	e.curscope = newScope(e.curscope)
}

func (e *symbolResolver) popScope() {
	e.curscope = e.curscope.parent
}

func (e *symbolResolver) collectGlobalSymbols(file *parse.File) error {
	for _, imp := range file.Imports {
		impdef := imp.Val[1 : len(imp.Val)-1]
		impdef = strings.Split(impdef, "/")[0]
		sym := newGlobalSymbol(imp.Span.Start)
		err := e.curscope.declareSym(impdef, sym)
		if err != nil {
			return fmt.Errorf("bad import: %s %s:%v", err, imp.Span.Path, imp.Span.Start)
		}
	}
	for _, fd := range file.FuncDecls {
		ty, err := e.funcDeclToGType(fd)
		if err != nil {
			return err
		}
		sym := newGlobalFuncSymbol(ty, fd.Span.Start)
		err = e.curscope.declareSym(fd.Name, sym)
		if err != nil {
			return fmt.Errorf("bad func decl: %s %s:%v", err, fd.Span.Path, fd.Span.Start)
		}
	}
	for _, vd := range file.VarDecls {
		sym := newGlobalSymbol(vd.Span.Start)
		err := e.curscope.declareSym(vd.Name, sym)
		if err != nil {
			return fmt.Errorf("bad var decl: %s %s:%v", err, vd.Span.Path, vd.Span.Start)
		}
	}
	return nil
}

func (e *symbolResolver) resolveSymbols(n parse.Node) error {
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

func (e *symbolResolver) resolveLocalVarDecl(vd *parse.VarDecl) error {
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
