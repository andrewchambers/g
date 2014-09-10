package emit

import (
	"bufio"
	"fmt"
	"github.com/andrewchambers/g/parse"
	"strings"
)

type emitter struct {
	gscope   *scope
	curscope *scope
	symbolMap map[parse.Node] symbol
	out      *bufio.Writer

	llvmNameCounter uint
	llvmLabelCounter uint
	curFuncType     *GFunc
}

type Value interface {
	getLLVMRepr() string
	isLVal() bool
	getGType() GType
}

type exprValue struct {
	llvmName string
	lval     bool
	gType    GType
}

type intConstant struct {
	val int64
}

type boolConstant struct {
	val bool
}

func (v *exprValue) getLLVMRepr() string {
	return v.llvmName
}

func (v *exprValue) isLVal() bool {
	return v.lval
}

func (v *exprValue) getGType() GType {
	return v.gType
}

func (c *intConstant) getLLVMRepr() string {
	return fmt.Sprintf("%v", c.val)
}

func (c *intConstant) isLVal() bool {
	return false
}

func (c *intConstant) getGType() GType {
	return &GConstant{}
}

func (c *boolConstant) getLLVMRepr() string {
	if c.val {
		return "1"
	}
	return "0"
}

func (c *boolConstant) isLVal() bool {
	return false
}

func (c *boolConstant) getGType() GType {
	return builtinBoolGType
}

func newEmitter(out *bufio.Writer) *emitter {
	ret := &emitter{}
	ret.curscope = newScope(nil)
	ret.gscope = ret.curscope
	ret.out = out
	ret.symbolMap = make(map[parse.Node] symbol)
	ret.addBuiltinTypes()
	return ret
}

func (e *emitter) newLLVMLabel() string {
	e.llvmLabelCounter++
	return fmt.Sprintf(".L%d", e.llvmLabelCounter)
}

func (e *emitter) newLLVMName() string {
	e.llvmNameCounter++
	return fmt.Sprintf("%%%d", e.llvmNameCounter)
}

func (e *emitter) addBuiltinTypes() {
	e.curscope.declareType("bool",builtinBoolGType)
	e.curscope.declareType("int", builtinInt64GType)
	e.curscope.declareType("int8", builtinInt8GType)
	e.curscope.declareType("int16",builtinInt16GType)
	e.curscope.declareType("int32",builtinInt32GType)
	e.curscope.declareType("int64",builtinInt64GType)
	e.curscope.declareType("uint", builtinUInt64GType)
	e.curscope.declareType("uint8", builtinUInt8GType)
	e.curscope.declareType("uint16",builtinUInt16GType)
	e.curscope.declareType("uint32",builtinUInt32GType)
	e.curscope.declareType("uint64",builtinUInt64GType)
	
	e.curscope.declareSym("true", &constSymbol{&boolConstant{true},nil})
	e.curscope.declareSym("false", &constSymbol{&boolConstant{true},nil})
}

func (e *emitter) pushScope() {
	e.curscope = newScope(e.curscope)
}

func (e *emitter) popScope() {
	e.curscope = e.curscope.parent
}

func (e *emitter) emit(s string, args ...interface{}) {
	fmt.Fprintf(e.out, s, args...)
}

func (e *emitter) emiti(s string, args ...interface{}) {
	e.emit("    "+s, args...)
}

func (e *emitter) emitl(l string) {
	e.emit("  %s:\n",l)
}

func EmitModule(out *bufio.Writer, file *parse.File) error {
	e := newEmitter(out)
	err := e.collectGlobalSymbols(file)
    if err != nil {
	    return err
    }
	e.emit("; compiled from file %s\n", file.Span.Path)
	e.emit("target triple = \"x86_64-pc-linux-gnu\"\n\n")
	for _, fd := range file.FuncDecls {
		err = e.emitFuncDecl(fd)
		if err != nil {
			return err
		}
	}
	out.Flush()
	return nil
}

func (e *emitter) resolveSymbols(n parse.Node) error {
    
        switch n := n.(type) {
            case *parse.FuncDecl:
                e.pushScope()
                for _,subn := range n.Body {
                    err := e.resolveSymbols(subn)
                    if err != nil {
                        return err
                    }
                }
                e.popScope()
            case *parse.VarDecl:
                err := e.resolveLocalVarDecl(n)
                if err != nil {
                    return err
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
                for _,subn := range n.Body {
                    err := e.resolveSymbols(subn)
                    if err != nil {
                        return err
                    }
                }
                for _,subn := range n.Els {
                    err := e.resolveSymbols(subn)
                    if err != nil {
                        return err
                    }
                }
            case *parse.Return:
                err := e.resolveSymbols(n.Expr)
                if err != nil {
                    return err
                }
            case *parse.Constant:
                //nothing
            case *parse.Ident:
                sym,err := e.curscope.lookupSym(n.Val)
                if err != nil {
                    return fmt.Errorf("%s at %s:%s",err,n.Span.Path,n.Span.Start)
                }
                e.symbolMap[n] = sym 
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
	return err
}

func (e *emitter) collectGlobalSymbols(file *parse.File) error {
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
		sym := newGlobalSymbol(fd.Span.Start)
		err := e.curscope.declareSym(fd.Name, sym)
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

func (e *emitter) emitFuncDecl(f *parse.FuncDecl) error {
	//Emit function start
	ft, err := e.funcDeclToGType(f)
	if err != nil {
		return err
	}
	e.curFuncType = ft
	e.emit("define %s @%s() {\n", gTypeToLLVM(ft.RetType), f.Name)
	e.emitl(".entry")
	err = e.resolveSymbols(f)
	if err != nil {
	    return err
	}
	for _, stmt := range f.Body {
		err = e.emitStatement(stmt)
		if err != nil {
			return err
		}
	}
	e.emit("}\n")
	return nil
}

func (e *emitter) emitStatement(stmt parse.Node) error {
	var err error
	switch stmt := stmt.(type) {
	case *parse.VarDecl:
		err = e.emitLocalVarDecl(stmt)
	case *parse.Assign:
		err = e.emitAssign(stmt)
	case *parse.Return:
		err = e.emitReturn(stmt)
	case *parse.If:
		err = e.emitIf(stmt)
	default:
		panic("unhandled Statement type...")
	}
	return err
}

func (e *emitter) emitLocalVarDecl(vd *parse.VarDecl) error {
    return nil
}

func (e *emitter) emitIf(i *parse.If) error {
    v,err := e.emitExpression(i.Cond)
    if err != nil {
        return err
    }
    if !isBool(v.getGType()) {
        return fmt.Errorf("If statements require a bool expression %s:%s",i.Span.Path,i.Span.Start)
    }
    if v.isLVal() {
        v,err = e.emitRemoveLValness(v)
        if err != nil {
            return err
        }
    }
    if isConstantVal(v) {
        v,err = e.emitRemoveConstant(v,builtinBoolGType)
        if err != nil {
            return err
        }
    }
    iftrue := e.newLLVMLabel()
    iffalse := e.newLLVMLabel()
    after := e.newLLVMLabel()
    
    e.emiti("br i1 %s, label %s, label %s\n",v.getLLVMRepr(),iftrue,iffalse)
    e.emitl(iftrue)
    for _, stmt := range i.Body {
        e.emitStatement(stmt)
    }
    e.emiti("br label %s\n",after)
    e.emitl(iffalse)
    for _, stmt := range i.Els {
        e.emitStatement(stmt)
    }
    e.emiti("br label %s\n",after)
    e.emitl(after)
    return nil
}

func (e *emitter) emitAssign(ass *parse.Assign) error {
	
	l,err := e.emitExpression(ass.L)
	if err != nil {
		return err
	}
	r,err := e.emitExpression(ass.R)
	if err != nil {
		return err
	}

	if !l.isLVal() {
		return fmt.Errorf("assigning to a non lvalue")
	}

	if isConstantVal(r) {
		r, err = e.emitRemoveConstant(r, l.getGType())
		if err != nil {
			return err
		}
	}

	if r.isLVal() {
		r, err = e.emitRemoveLValness(r)
		if err != nil {
			return err
		}
	}

	if !l.getGType().Equals(r.getGType()) {
		panic("assignment of incompatible types")
	}
	lstr := fmt.Sprintf("%s* %s", gTypeToLLVM(l.getGType()), l.getLLVMRepr())
	err = e.emitStore(lstr, r)
	if err != nil {
		return err
	}
	return nil
}

func (e *emitter) emitStore(llvmptr string, v Value) error {
	t := v.getGType()
	switch t.(type) {
	case *GInt:
		e.emiti("store %s %s, %s\n", gTypeToLLVM(t), v.getLLVMRepr(), llvmptr)
		return nil
	default:
		return fmt.Errorf("dont know how to store type")
	}
}

func (e *emitter) emitReturn(r *parse.Return) error {
	v,err := e.emitExpression(r.Expr)
	if err != nil {
		return err
	}
	if v.isLVal() {
		v, err = e.emitRemoveLValness(v)
		if err != nil {
			return err
		}
	}
	var llvmType string
	if isConstantVal(v) {
		v, err = e.emitRemoveConstant(v, e.curFuncType.RetType)
		if err != nil {
			return fmt.Errorf("unable to convert constant to return type at %s", r.Span.Start)
		}
		llvmType = gTypeToLLVM(e.curFuncType.RetType)
	} else {
		if !v.getGType().Equals(e.curFuncType.RetType) {
			return fmt.Errorf("type does not match function return type %s", r.Span.Start)
		}
		llvmType = gTypeToLLVM(v.getGType())

	}
	e.emiti("ret %s %s\n", llvmType, v.getLLVMRepr())
	return nil
}

func (e *emitter) emitZeroMem(name string, t GType) error {
	switch t := t.(type) {
	case *GInt:
		switch t.Bits {
		case 64:
			e.emiti("store i64 0, i64* %s\n", name)
		case 32:
			e.emiti("store i32 0, i32* %s\n", name)
		case 16:
			e.emiti("store i16 0, i16* %s\n", name)
		case 8:
			e.emiti("store i8 0, i8* %s\n", name)
		case 1:
			e.emiti("store i1 0, i1* %s\n", name)
		default:
			panic("internal error")
		}
		return nil
	default:
		return fmt.Errorf("unable to zero memory for type %s", t)
	}
}

func (e *emitter) emitRemoveLValness(v Value) (Value, error) {
	if !v.isLVal() {
		panic("internal error")
	}
	switch v := v.(type) {
	case *exprValue:
		name := e.newLLVMName()
		e.emiti("%s = load %s* %s\n", name, gTypeToLLVM(v.getGType()), v.getLLVMRepr())
		ret := &exprValue{
			llvmName: name,
			lval:     false,
			gType:    v.getGType(),
		}
		return ret, nil
	default:
		panic("internal error")
	}
}

func (e *emitter) emitRemoveConstant(v Value, hint GType) (Value, error) {
	if !isConstantVal(v) {
		panic("internal error")
	}
	switch v := v.(type) {
	case *intConstant:
		switch hint := hint.(type) {
		case *GInt:
			if isBool(hint) {
				return nil, fmt.Errorf("cannot convert numeric constant to bool")
			}
			ret := &exprValue{
				llvmName: fmt.Sprintf("%d", v.val),
				lval:     false,
				gType:    hint,
			}
			return ret, nil
		default:
			return nil, fmt.Errorf("unable to convert numeric constant to XXX")
		}
    case *boolConstant:
		switch hint := hint.(type) {
		case *GInt:
			if !isBool(hint) {
				return nil, fmt.Errorf("cannot convert bool constant to number")
			}
			val := 0
			if v.val {
			    val = 1
			}
			ret := &exprValue{
				llvmName: fmt.Sprintf("%d", val),
				lval:     false,
				gType:    hint,
			}
			return ret, nil
		default:
			return nil, fmt.Errorf("unable to convert bool constant to XXX")
		}
	default:
		return nil, fmt.Errorf("internal error emitRemoveConstant %v", v)
	}
}

func (e *emitter) emitExpression(expr parse.Node) (Value,error) {
	switch expr := expr.(type) {
	case *parse.Constant:
		v := &intConstant{expr.Val}
		return v,nil
	case *parse.Binop:
		return e.emitBinop(expr)
	case *parse.Ident:
		return e.emitIdent(expr)
	default:
		panic("unhandled...")
	}
}

func (e *emitter) emitIdent(i *parse.Ident) (Value,error) {
	s, ok := e.symbolMap[i]
	if !ok {
		panic("internal error")
	}
	switch s := s.(type) {
	case *localSymbol:
	    ret :=  &exprValue{
			        llvmName: s.alloca,
			        lval:     true,
			        gType:    s.getGType(),
		        }   
		return ret,nil
    case *constSymbol:
        return s.v,nil
	default:
		panic("unhandled...")
	}
}

func isConstantVal(v Value) bool {
	_, ok := v.(*intConstant)
	if ok {
		return true
	}
	_, ok = v.(*boolConstant)
	if ok {
		return true
	}
	return false
}

func isIntType(t GType) bool {
	_, ok := t.(*GInt)
	return ok
}

func (e *emitter) emitBinop(b *parse.Binop) (Value,error) {
	l,err := e.emitExpression(b.L)
	if err != nil {
	    return nil,err
	}
	r,err := e.emitExpression(b.R)
	if err != nil {
	    return nil,err
	}

	if isConstantVal(l) && isConstantVal(r) {
		c, err := foldConstantBinop(b.Op, l, r)
		if err != nil {
			return nil,err
		}
		return c,nil
	}

	if isConstantVal(l) {
		l, err = e.emitRemoveConstant(l, r.getGType())
		if err != nil {
			return nil,err
		}
	}

	if isConstantVal(r) {
		r, err = e.emitRemoveConstant(r, l.getGType())
		if err != nil {
			return nil,err
		}
	}

	if l.isLVal() {
		l, err = e.emitRemoveLValness(l)
		if err != nil {
			return nil,err
		}
	}

	if r.isLVal() {
		r, err = e.emitRemoveLValness(r)
		if err != nil {
			return nil,err
		}
	}

	if !l.getGType().Equals(r.getGType()) {
		panic("arithmetic on incompatible types")
	}

	if !isIntType(l.getGType()) {
		panic("arithmetic on non int type")
	}


	if isBool(l.getGType()) {
		switch b.Op {
		case parse.EQ:
		default:
			return nil,fmt.Errorf("binop %s cannot be performed on type bool", b.Op)
		}
	}
    
    llty := gTypeToLLVM(l.getGType())
    
	switch b.Op {
	case parse.EQ:
		ret := &exprValue{
		    llvmName: e.newLLVMName(),
		    gType:    builtinBoolGType,
		    lval:     false,
	    }
		e.emiti("%s = icmp eq %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
		return ret,nil
	default:
		panic("unreachable")
	}    

	ret := &exprValue{
		llvmName: e.newLLVMName(),
		gType:    l.getGType(),
		lval:     false,
	}


	switch b.Op {
	case '+':
		e.emiti("%s = add %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	case '-':
		e.emiti("%s = sub %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	case '*':
		e.emiti("%s = mul %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	case '/':
		e.emiti("%s = sdiv %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	case '%':
		e.emiti("%s = srem %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	case '^':
		e.emiti("%s = xor %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	default:
		panic("unreachable")
	}
	return ret,nil
}

func (e *emitter) funcDeclToGType(f *parse.FuncDecl) (*GFunc, error) {
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
	}

	return ret, nil
}

func (e *emitter) parseNodeToGType(n parse.Node) (GType, error) {
	//span := n.GetSpan()
	switch n := n.(type) {
	case *parse.TypeAlias:
		ret, err := e.curscope.lookupType(n.Name)
		return ret, err
	default:
		return nil, fmt.Errorf("invalid type")
	}

	panic("unreachable")
}

func gTypeToLLVM(t GType) string {
	switch t := t.(type) {
	case *GInt:
		switch t.Bits {
		case 64:
			return "i64"
		case 32:
			return "i32"
		case 16:
			return "i16"
		case 8:
			return "i8"
		case 1:
			return "i1"
		default:
			panic("unreachable.")
		}
	default:
		panic("unreachable: bad gtype " + fmt.Sprintf("%s", t))
	}
}
