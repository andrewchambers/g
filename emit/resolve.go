package emit

import (
	"fmt"
	"strings"
	"github.com/andrewchambers/g/parse"
	"github.com/andrewchambers/g/target"
)

type symbolTable map[parse.Node]symbol

type symbolResolver struct {
	gscope    *scope
	curscope  *scope
	symbols map[parse.Node]symbol
}

//XXX rename all variables from e

func resolveASTSymbols(machine target.TargetMachine,n parse.Node) (symbolTable, error) {
	sr := &symbolResolver{}
	sr.gscope = newScope(nil)
	sr.curscope = sr.gscope
    sr.symbols = make(symbolTable)
	addBuiltinTypes(machine,sr)
	err := sr.resolveSymbols(n)
	if err != nil {
		return make(symbolTable), err
	}
	return sr.symbols, nil
}

func addBuiltinTypes(machine target.TargetMachine,e *symbolResolver) {
	// Built in types
	e.curscope.declareType("bool", builtinBoolGType)
	e.curscope.declareType("int", getDefaultIntType(machine))
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
		//XXX declare function name and other symbols...
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
		e.symbols[n] = sym
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
	s := &localSymbol{
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

func (e *symbolResolver) funcDeclToGType(f *parse.FuncDecl) (*GFunc, error) {
	ret := &GFunc{}
	for _, t := range f.ArgTypes {
		gty, err := e.parseNodeToGType(t)
		if err != nil {
			return nil, err
		}
		ret.ArgTypes = append(ret.ArgTypes, gty)
	}
	if f.RetType != nil {
		t, err := e.parseNodeToGType(f.RetType)
		if err != nil {
			return nil, err
		}
		ret.RetType = t
	} else {
		ret.RetType = builtinVoidGType
	}

	return ret, nil
}

func (e *symbolResolver) parseNodeToGType(n parse.Node) (GType, error) {
	//span := n.GetSpan()
	switch n := n.(type) {
	case *parse.TypeAlias:
		/*
			ret, err := e.curscope.lookupType(n.Name)
			return ret, err
		*/
		return nil, fmt.Errorf("foo\n")
	case *parse.PointerTo:
		t, err := e.parseNodeToGType(n.PointsTo)
		if err != nil {
			return nil, err
		}
		ret := &GPointer{PointsTo: t}
		return ret, nil
	case *parse.Struct:
		return e.parseStructToGType(n)
	case *parse.ArrayOf:
		t, err := e.parseNodeToGType(n.SubType)
		if err != nil {
			return nil, err
		}
		ret := &GArray{}
		ret.Dim = 12
		ret.SubType = t
		return ret, nil
	default:
		return nil, fmt.Errorf("invalid type %v", n)
	}

}

func (e *symbolResolver) parseStructToGType(n *parse.Struct) (GType, error) {
	ret := &GStruct{}
	for idx, name := range n.Names {
		t, err := e.parseNodeToGType(n.Types[idx])
		if err != nil {
			return nil, err
		}
		ret.Names = append(ret.Names, name)
		ret.Types = append(ret.Types, t)
	}
	return ret, nil
}

