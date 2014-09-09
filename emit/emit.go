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
	out      *bufio.Writer
    
    llvmNameCounter uint
	curFuncType *GFunc
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
	return NewGInt(1,false)
}


func newEmitter(out *bufio.Writer) *emitter {
	ret := &emitter{}
	ret.curscope = newScope(nil)
	ret.gscope = ret.curscope
	ret.out = out
	ret.addBuiltinTypes()
	return ret
}


func (e *emitter) newLLVMName() string {
    e.llvmNameCounter++
	return fmt.Sprintf("%%%d",e.llvmNameCounter)
}

func (e *emitter) addBuiltinTypes() {
	e.curscope.declareType("bool",NewGInt(1, false))
	e.curscope.declareType("int", NewGInt(e.getIntWidth(), true))
	e.curscope.declareType("uint",NewGInt(e.getIntWidth(), false))
	e.curscope.declareType("int32", NewGInt(32, true))
	e.curscope.declareType("uint32",NewGInt(32, false))
	e.curscope.declareType("int64", NewGInt(64, true))
	e.curscope.declareType("uint64",NewGInt(64, false))
}

//XXX shift to arch type
func (e *emitter) getIntWidth() uint {
	return 64
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
	for _, stmt := range f.Body {
		e.emitStatement(stmt)
	}
	e.emit("}\n")
	return nil
}

func (e *emitter) emitStatement(stmt parse.Node) {
	switch stmt := stmt.(type) {
	case *parse.VarDecl:
	    e.emitLocalVarDecl(stmt)
	case *parse.Return:
		e.emitReturn(stmt)
	default:
		panic("unhandled Statement type...")
	}
}

func (e *emitter) emitLocalVarDecl(vd *parse.VarDecl) {
    t,err := e.parseNodeToGType(vd.Type)
    if err != nil {
        panic("unhandled err " + err.Error())
    }
    name := e.newLLVMName()
    e.emiti("%s = alloca %s\n",name,gTypeToLLVM(t))
    e.emitZeroMem(name,t)
    s := &localSymbol{
         alloca: name,
         gType:  t,
         defPos: vd.Span.Start,
    }
    e.curscope.declareSym(vd.Name,s)
} 

func (e *emitter) emitReturn(r *parse.Return) {
	var err error
	v := e.emitExpression(r.Expr)
	if v.isLVal() {
	    v,err = e.emitRemoveLValness(v)
	    if err != nil {
	        panic("error emitting return " + err.Error())
	    }
	}
	var llvmType string
	if isConstantVal(v) {
		v,err = e.emitRemoveConstant(v,e.curFuncType.RetType)
		if err != nil {
		    panic("error converting constant to return type")
		}
		llvmType = gTypeToLLVM(e.curFuncType.RetType)
	} else {
		if !v.getGType().Equals(e.curFuncType.RetType) {
			panic("failed type check")
		}
		llvmType = gTypeToLLVM(v.getGType())

	}
	e.emiti("ret %s %s\n", llvmType, v.getLLVMRepr())
}

func (e *emitter) emitZeroMem(name string,t GType) error {
    switch t := t.(type) {
        case *GInt:
            switch t.Bits {
                case 64:
                    e.emiti("store i64 0, i64* %s\n",name)
                case 32:
                    e.emiti("store i32 0, i32* %s\n",name)
                case 16:
                    e.emiti("store i16 0, i16* %s\n",name)
                case 8:
                    e.emiti("store i8 0, i8* %s\n",name)
                case 1:
                    e.emiti("store i1 0, i1* %s\n",name)
                default:
                    panic("internal error")
            }
            return nil
        default:
            return fmt.Errorf("unable to zero memory for type %s",t)
    }
}

func (e *emitter) emitRemoveLValness(v Value) (Value,error) {
    if !v.isLVal() {
        panic("internal error")
    }
    switch v := v.(type) {
        case *exprValue:
            name := e.newLLVMName()
            e.emiti("%s = load %s* %s\n",name,gTypeToLLVM(v.getGType()),v.getLLVMRepr())
            ret := &exprValue{
                llvmName: name,
	            lval:     false,
	            gType:    v.getGType(),
            }
            return ret,nil
        default:
            panic("internal error")
    }
}

func (e *emitter) emitRemoveConstant(v Value,hint GType) (Value,error) {
    if !isConstantVal(v) {
        panic("internal error")
    }
    switch v := v.(type) {
        case *intConstant:
            switch hint := hint.(type) {
                case *GInt:
                    ret := &exprValue{
                        llvmName: fmt.Sprintf("%d",v.val),
	                    lval:     false,
	                    gType:    hint,
                    }
                    return ret,nil
                default:
                    return nil,fmt.Errorf("unable to convert constant to XXX")
            }
        default:
            return nil,fmt.Errorf("internal error emitRemoveConstantNess")
    }
}

func (e *emitter) emitExpression(expr parse.Node) Value {
	switch expr := expr.(type) {
	case *parse.Constant:
		v := &intConstant{expr.Val}
		return v
	case *parse.Binop:
		return e.emitBinop(expr)
	case *parse.Ident:
	    return e.emitIdent(expr)
	default:
		panic("unhandled...")
	}
}

func (e *emitter) emitIdent(i *parse.Ident) Value {
    s,err := e.curscope.lookupSym(i.Val)
    if err != nil {
        panic(err)
    }
    switch s := s.(type) {
        case *localSymbol:
            return &exprValue{
                llvmName: s.alloca,
                lval: true,
                gType: s.getGType(),
            }
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

func (e *emitter) emitBinop(b *parse.Binop) Value {
    var err error
	l := e.emitExpression(b.L)
	r := e.emitExpression(b.R)

	if isConstantVal(l) && isConstantVal(r) {
		c, err := foldConstantBinop(b.Op, l, r)
		if err != nil {
			panic(err)
		}
		return c
	}
	
	if isConstantVal(l) {
	    l,err = e.emitRemoveConstant(l,r.getGType())
	    if err != nil {
	        panic(err.Error())
	    }
	}
	
	if isConstantVal(r) {
	    r,err = e.emitRemoveConstant(r,l.getGType())
	    if err != nil {
	        panic(err.Error())
	    }
	}
	
	if l.isLVal() {
	    l,err = e.emitRemoveLValness(l)
	    if err != nil {
	        panic(err.Error())
	    }
	}
	
	if r.isLVal() {
	    r,err = e.emitRemoveLValness(r)
	    if err != nil {
	        panic(err.Error())
	    }
	}

	lstr := fmt.Sprintf("%s %s", gTypeToLLVM(l.getGType()), l.getLLVMRepr())
	rstr := fmt.Sprintf("%s %s", gTypeToLLVM(r.getGType()), r.getLLVMRepr())

	if !l.getGType().Equals(r.getGType()) {
		panic("arithmetic on incompatible types")
	}

	if !isIntType(l.getGType()) {
		panic("arithmetic on non int type")
	}
    
    ret := &exprValue{
        llvmName: e.newLLVMName(),
        gType: l.getGType(),
        lval: false,
    }
    
	switch b.Op {
	case '+':
		e.emiti("%s = add %s %s\n",ret.llvmName, lstr, rstr)
		return ret
	default:
		panic("unreachable")
	}

	panic("unreachable")
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
		return ret,err
	default:
	    return nil,fmt.Errorf("invalid type")
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
		panic("unreachable: bad gtype " + fmt.Sprintf("%s",t))
	}
}
