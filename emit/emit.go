package emit

import (
	"bufio"
	"fmt"
	"github.com/andrewchambers/g/parse"
	"github.com/andrewchambers/g/target"
	"strings"
)

type emitter struct {

    machine target.TargetMachine

	gscope    *scope
	curscope  *scope
	symbolMap map[parse.Node]symbol
	out       *bufio.Writer

	llvmNameCounter  uint
	llvmLabelCounter uint
	curFuncType      *GFunc

	// Have we emitted any instructions into
	// The current basic block?
	isCurBlockEmpty bool
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

func newEmitter(m target.TargetMachine,out *bufio.Writer) *emitter {
	ret := &emitter{}
	ret.machine = m
	ret.curscope = newScope(nil)
	ret.gscope = ret.curscope
	ret.out = out
	ret.symbolMap = make(map[parse.Node]symbol)
	ret.addBuiltinTypes()
	return ret
}

func (e *emitter) newLLVMLabel() string {
	ret := fmt.Sprintf(".L%d", e.llvmLabelCounter)
	e.llvmLabelCounter++
	return ret
}

func (e *emitter) newLLVMName() string {
	ret := fmt.Sprintf("%%%d", e.llvmNameCounter)
	e.llvmNameCounter++
	return ret
}

func (e *emitter) addBuiltinTypes() {
	// Built in types
	e.curscope.declareType("bool", builtinBoolGType)
	e.curscope.declareType("int", builtinInt64GType)
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
	e.isCurBlockEmpty = false
	e.emit("    "+s, args...)
}

func (e *emitter) emitrawi(s string) {
	e.isCurBlockEmpty = false
	fmt.Fprint(e.out, "    "+s)
}

func (e *emitter) emitl(l string) {
	if e.isCurBlockEmpty {
		e.emiti("br label %%%s\n", l)
	}
	e.isCurBlockEmpty = true
	e.emit("  %s:\n", l)
}

func EmitModule(machine target.TargetMachine,out *bufio.Writer, file *parse.File) error {
	e := newEmitter(machine,out)
	err := e.collectGlobalSymbols(file)
	if err != nil {
		return err
	}
	
	e.emitPrelude()

	for _, fd := range file.FuncDecls {
		e.llvmLabelCounter = 0
		e.llvmNameCounter = 0
		err = e.emitFuncDecl(fd)
		if err != nil {
			return err
		}
	}
	out.Flush()
	return nil
}

func (e *emitter) emitPrelude() {
    e.emit("target = \"%s\"\n\n",e.machine.LLVMTargetTriple());
}

func (e *emitter) handleFuncPrologue(fd *parse.FuncDecl) error {
	for idx, arg := range fd.ArgNames {
		ty := fd.ArgTypes[idx]
		gty, err := e.parseNodeToGType(ty)
		if err != nil {
			return err
		}
		llname := e.newLLVMName()
		v := &exprValue{}
		v.gType = gty
		v.llvmName = "%" + arg
		e.emiti("%s = alloca %s\n", llname, gTypeToLLVM(gty))
		e.emitStore(llname, v)
		s := &localSymbol{
			alloca: llname,
			gType:  gty,
			defPos: fd.Span.Start,
		}
		e.curscope.declareSym(arg, s)
	}
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

func (e *emitter) emitFuncDecl(f *parse.FuncDecl) error {
	//Emit function start
	ft, err := e.funcDeclToGType(f)
	if err != nil {
		return err
	}
	e.curFuncType = ft
	args := ""
	for idx, argname := range f.ArgNames {
		ty := f.ArgTypes[idx]
		gty, err := e.parseNodeToGType(ty)
		if err != nil {
			return err
		}
		args += fmt.Sprintf("%s %%%s", gTypeToLLVM(gty), argname)
		if idx != len(f.ArgNames)-1 {
			args += ","
		}
	}
	llvmRetTy := ""
	rty := ft.RetType
	llvmRetTy = gTypeToLLVM(rty)
	e.emit("define %s @%s( %s ) {\n", llvmRetTy, f.Name, args)
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
	e.emit("}\n\n")
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
	case *parse.For:
		err = e.emitFor(stmt)
	case *parse.EmptyStatement:
		err = nil
	case *parse.ExpressionStatement:
		_, err = e.emitExpression(stmt.Expr)
	default:
		panic(stmt)
	}
	return err
}

func (e *emitter) emitLocalVarDecl(vd *parse.VarDecl) error {
	if vd.Init != nil {
		err := e.emitAssign(vd.Init)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *emitter) emitIf(i *parse.If) error {
	v, err := e.emitExpression(i.Cond)
	if err != nil {
		return err
	}
	if !isBool(v.getGType()) {
		return fmt.Errorf("If statements require a bool expression %s:%s", i.Span.Path, i.Span.Start)
	}
	if v.isLVal() {
		v, err = e.emitRemoveLValness(v)
		if err != nil {
			return err
		}
	}
	if isConstantVal(v) {
		v, err = e.emitRemoveConstant(v, builtinBoolGType)
		if err != nil {
			return err
		}
	}
	iftrue := e.newLLVMLabel()
	iffalse := e.newLLVMLabel()
	after := e.newLLVMLabel()

	e.emiti("br i1 %s, label %%%s, label %%%s\n", v.getLLVMRepr(), iftrue, iffalse)
	e.emitl(iftrue)
	for _, stmt := range i.Body {
		e.emitStatement(stmt)
	}
	e.emiti("br label %%%s\n", after)
	e.emitl(iffalse)
	for _, stmt := range i.Els {
		e.emitStatement(stmt)
	}
	e.emiti("br label %%%s\n", after)
	e.emitl(after)
	return nil
}

func (e *emitter) emitFor(f *parse.For) error {
	if f.Init != nil {
		err := e.emitStatement(f.Init)
		if err != nil {
			return err
		}
	}

	loopbegin := e.newLLVMLabel()
	loopbody := e.newLLVMLabel()
	loopexit := e.newLLVMLabel()

	e.emiti("br label %%%s\n", loopbegin)
	e.emitl(loopbegin)

	if f.Cond != nil {
		v, err := e.emitExpression(f.Cond)
		if err != nil {
			return err
		}
		if !isBool(v.getGType()) {
			return fmt.Errorf("For loop condition requires a bool expression %s:%s", f.Cond.GetSpan().Path, f.Cond.GetSpan().Start)
		}
		if v.isLVal() {
			v, err = e.emitRemoveLValness(v)
			if err != nil {
				return err
			}
		}
		if isConstantVal(v) {
			v, err = e.emitRemoveConstant(v, builtinBoolGType)
			if err != nil {
				return err
			}
		}
		e.emiti("br i1 %s, label %%%s, label %%%s\n", v.getLLVMRepr(), loopbody, loopexit)
	}
	e.emitl(loopbody)
	for _, stmt := range f.Body {
		err := e.emitStatement(stmt)
		if err != nil {
			return err
		}
	}
	if f.Step != nil {
		err := e.emitStatement(f.Step)
		if err != nil {
			return err
		}
	}
	e.emiti("br label %%%s\n", loopbegin)
	e.emitl(loopexit)
	return nil
}

func (e *emitter) emitAssign(ass *parse.Assign) error {
    
    
	l, err := e.emitExpression(ass.L)
	if err != nil {
		return err
	}
	r, err := e.emitExpression(ass.R)
	if err != nil {
		return err
	}

	if !l.isLVal() {
		return fmt.Errorf("assigning to a non lvalue")
	}
	
    switch ass.Op {
        case parse.ADDASSIGN:
            r,err = e.emitBinop2('+',l,r)
            if err != nil {
                return err
            }
        case parse.SUBASSIGN:
            r,err = e.emitBinop2('-',l,r)
            if err != nil {
                return err
            }
        case parse.MULASSIGN:
            r,err = e.emitBinop2('*',l,r)
            if err != nil {
                return err
            }
        case '=':
            //pass
        default:
            panic(ass.Op)
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
		return fmt.Errorf("assignment of incompatible types %s and %s",l.getGType(),r.getGType())
	}
	err = e.emitStore(l.getLLVMRepr(), r)
	if err != nil {
		return err
	}
	return nil
}

func (e *emitter) emitStore(llvmptr string, v Value) error {
	t := v.getGType()
	switch t.(type) {
	case *GPointer:
		e.emiti("store %s %s, %s* %s\n", gTypeToLLVM(t), v.getLLVMRepr(), gTypeToLLVM(t), llvmptr)
		return nil
	case *GInt:
		e.emiti("store %s %s, %s* %s\n", gTypeToLLVM(t), v.getLLVMRepr(), gTypeToLLVM(t), llvmptr)
		return nil
	default:
		return fmt.Errorf("dont know how to store type")
	}
}

func (e *emitter) emitReturn(r *parse.Return) error {

	if r.Expr == nil {
		if !e.curFuncType.RetType.Equals(builtinVoidGType) {
			return fmt.Errorf("function expects empty return")
		}
		e.emiti("ret void\n")
		return nil
	}

	v, err := e.emitExpression(r.Expr)
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
	case *GPointer:
		e.emiti("store %s null, %s* %s\n", gTypeToLLVM(t), gTypeToLLVM(t), name)
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
	default:
	    szx := e.newLLVMName()
	    sz := e.newLLVMName()
	    llty := gTypeToLLVM(t)
	    e.emiti("%s = getelementptr %s* null, i32 1\n",szx,llty)
	    e.emiti("%s = cast %s* %s to i64\n",sz,llty,szx)
	    e.emiti("call void @memset(%s, i8 0, i64 %s)\n",name,sz)
	}
	return nil
}

func (e *emitter) emitRemoveLValness(v Value) (Value, error) {
	if !v.isLVal() {
		panic("internal error")
	}
	switch v := v.(type) {
	case *exprValue:
	    if isArrayVal(v) {
	        //Arrays are already stored as pointers, so there is nothing to do.
	        //XXX this may be wrong.
	        ret := &exprValue{
			    llvmName: v.llvmName,
			    lval:     false,
			    gType:    v.getGType(),
		    }
		    return ret, nil
	    }
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

func (e *emitter) emitExpression(expr parse.Node) (Value, error) {
	switch expr := expr.(type) {
	case *parse.Constant:
		v := &intConstant{expr.Val}
		return v, nil
	case *parse.Call:
		return e.emitCall(expr)
	case *parse.Binop:
		return e.emitBinop(expr)
	case *parse.Unop:
		return e.emitUnop(expr)
	case *parse.IndexInto:
		return e.emitIndex(expr)
	case *parse.Ident:
		return e.emitIdent(expr)
	default:
		panic(expr)
	}
}

func (e *emitter) emitIndex(i *parse.IndexInto) (Value,error) {
    v,err := e.emitExpression(i.Expr)
    if err != nil {
        return nil,err
    }
    
    idx,err := e.emitExpression(i.Index)
    if err != nil {
        return nil,err
    }
	if isConstantVal(idx) {
	    // Depends on what builtin int type is maybe?
		idx, err = e.emitRemoveConstant(idx, builtinUInt64GType)
		if err != nil {
			return nil, err
		}
	}
	if idx.isLVal() {
		idx, err = e.emitRemoveLValness(idx)
		if err != nil {
			return nil, err
		}
	}
    
    if !isPointerVal(v) && !isArrayVal(v) {
        return nil,fmt.Errorf("%s is a non indexable type",v.getGType())
    }
    
    var ret Value = nil
    
    if v.isLVal() {
        v,err = e.emitRemoveLValness(v)
        if err != nil {
            return nil,err
        }
    }
    retname := e.newLLVMName()
    e.emiti("%s = getelementptr %s %s, %s %s",retname,gTypeToLLVM(v.getGType()),v.getLLVMRepr(),gTypeToLLVM(idx.getGType()),idx.getLLVMRepr())
    retv := &exprValue{}
    retv.lval = true
    retv.gType = getSubType(v.getGType())
    retv.llvmName = retname
    ret = retv

    return ret,nil
} 

func (e *emitter) emitCall(c *parse.Call) (Value, error) {

	var funcType *GFunc
	isGlobalCall := false
	funcName := ""

	sym, ok := e.symbolMap[c.FuncLike]
	if ok {
		fsym, ok := sym.(*globalFuncSymbol)
		if ok {
			funcType = fsym.gType
			ident, ok := c.FuncLike.(*parse.Ident)
			if !ok {
				panic("internal error - non ident in symbol map!")
			}
			funcName = ident.Val
			isGlobalCall = true
		}
	} else {
		panic("unimplemented...")
	}
	if len(c.Args) != len(funcType.ArgTypes) {
		return nil, fmt.Errorf("expected %d argument(s), got %d", len(funcType.ArgTypes), len(c.Args))
	}

	argvalues := make([]Value, len(c.Args))
	for idx, argNode := range c.Args {
		arg, err := e.emitExpression(argNode)
		if err != nil {
			return nil, err
		}
		if isConstantVal(arg) {
			arg, err = e.emitRemoveConstant(arg, funcType.ArgTypes[idx])
			if err != nil {
				return nil, err
			}
		}
		if arg.isLVal() {
			arg, err = e.emitRemoveLValness(arg)
			if err != nil {
				return nil, err
			}
		}
		if !funcType.ArgTypes[idx].Equals(arg.getGType()) {
			return nil, fmt.Errorf("incorrect arg type.")
		}
		argvalues[idx] = arg
	}
	funcret := ""
	if isGlobalCall {
		callinst := ""
		if funcType.RetType.Equals(builtinVoidGType) {
			callinst = fmt.Sprintf("call void @%s (", funcName)
		} else {
			funcret = e.newLLVMName()
			callinst = fmt.Sprintf("%s = call %s @%s (", funcret, gTypeToLLVM(funcType.RetType), funcName)
		}
		for i, v := range argvalues {
			callinst += fmt.Sprintf("%s %s", gTypeToLLVM(v.getGType()), v.getLLVMRepr())
			if i != len(argvalues)-1 {
				callinst += ", "
			}
		}
		callinst += ")\n"
		e.emitrawi(callinst)
		var ret Value
		if funcType.RetType.Equals(builtinVoidGType) {
			ret = &exprValue{
				llvmName: "void",
				gType:    builtinVoidGType,
				lval:     false,
			}
		} else {
			ret = &exprValue{
				llvmName: funcret,
				gType:    funcType.RetType,
				lval:     false,
			}
		}

		return ret, nil
	} else {
		panic("unimplemented")
	}

	panic("unreachable")
}

func (e *emitter) emitIdent(i *parse.Ident) (Value, error) {
	s, ok := e.symbolMap[i]
	if !ok {
		panic("internal error")
	}
	switch s := s.(type) {
	case *localSymbol:
		ret := &exprValue{
			llvmName: s.alloca,
			lval:     true,
			gType:    s.getGType(),
		}
		return ret, nil
	case *constSymbol:
		return s.v, nil
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

func isPointerVal(v Value) bool {
	t := v.getGType()
	_,ok := t.(*GPointer)
	return ok
}

func isArrayVal(v Value) bool {
	t := v.getGType()
	_,ok := t.(*GArray)
	return ok
}

func getSubType(t GType) GType {
    switch t := t.(type) {
        case *GPointer:
            return t.PointsTo
        case *GArray:
            return t.SubType
        default:
        panic("bad type")
    }
}

func isIntType(t GType) bool {
	_, ok := t.(*GInt)
	return ok
}

func (e *emitter) emitBinop2(op parse.TokenKind, l,r Value) (Value, error) {
    
    var err error
    
	if isConstantVal(l) && isConstantVal(r) {
		c, err := foldConstantBinop(op, l, r)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	if isConstantVal(l) {
		l, err = e.emitRemoveConstant(l, r.getGType())
		if err != nil {
			return nil, err
		}
	}

	if isConstantVal(r) {
		r, err = e.emitRemoveConstant(r, l.getGType())
		if err != nil {
			return nil, err
		}
	}

	if l.isLVal() {
		l, err = e.emitRemoveLValness(l)
		if err != nil {
			return nil, err
		}
	}

	if r.isLVal() {
		r, err = e.emitRemoveLValness(r)
		if err != nil {
			return nil, err
		}
	}

	if !l.getGType().Equals(r.getGType()) {
		panic("arithmetic on incompatible types")
	}

	if !isIntType(l.getGType()) {
		panic("arithmetic on non int type")
	}

	if isBool(l.getGType()) {
		switch op {
		case parse.EQ:
		default:
			return nil, fmt.Errorf("binop %s cannot be performed on type bool", op)
		}
	}

	llty := gTypeToLLVM(l.getGType())

	//XXX refactor this into less code
	switch op {
	case parse.EQ:
		ret := &exprValue{
			llvmName: e.newLLVMName(),
			gType:    builtinBoolGType,
			lval:     false,
		}
		e.emiti("%s = icmp eq %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
		return ret, nil
	case '<':
		ret := &exprValue{
			llvmName: e.newLLVMName(),
			gType:    builtinBoolGType,
			lval:     false,
		}
		e.emiti("%s = icmp slt %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
		return ret, nil
	}

	ret := &exprValue{
		llvmName: e.newLLVMName(),
		gType:    l.getGType(),
		lval:     false,
	}

	switch op {
	case parse.RSHIFT:
	    e.emiti("%s = shr %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	case parse.LSHIFT:
	    e.emiti("%s = shl %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
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
	case '|':
		e.emiti("%s = or %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	case '&':
		e.emiti("%s = and %s %s, %s\n", ret.llvmName, llty, l.getLLVMRepr(), r.getLLVMRepr())
	default:
		return nil,fmt.Errorf("unimplemented binary operator %s",op)
	}
	return ret, nil
}

func (e *emitter) emitBinop(b *parse.Binop) (Value, error) {
    l, err := e.emitExpression(b.L)
	if err != nil {
		return nil, err
	}
	r, err := e.emitExpression(b.R)
	if err != nil {
		return nil, err
	}
	v,err := e.emitBinop2(b.Op,l,r)
	return v,err
}

func (e *emitter) emitUnop(u *parse.Unop) (Value, error) {
	v, err := e.emitExpression(u.Expr)
	if err != nil {
		return nil, err
	}

	if isConstantVal(v) {
	    v,err = foldConstantUnop(u.Op,v)
	    return v,err
	}

	switch u.Op {
	case '&':
		if !v.isLVal() {
			return nil, fmt.Errorf("cannot take address of non lvalue")
		}
		return &exprValue{
			lval:     false,
			llvmName: v.getLLVMRepr(),
			gType:    &GPointer{v.getGType()},
		}, nil
	case '*':
		if v.isLVal() {
			v, err = e.emitRemoveLValness(v)
			if err != nil {
				return nil, err
			}
		}
		p, ok := v.getGType().(*GPointer)
		if !ok {
			return nil, fmt.Errorf("cannot dereference non pointer type")
		}
		return &exprValue{
			lval:     true,
			gType:    p.PointsTo,
			llvmName: v.getLLVMRepr(),
		}, nil
	}
	panic("internal error")
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
	} else {
		ret.RetType = builtinVoidGType
	}

	return ret, nil
}

func (e *emitter) parseNodeToGType(n parse.Node) (GType, error) {
	//span := n.GetSpan()
	switch n := n.(type) {
	case *parse.TypeAlias:
		ret, err := e.curscope.lookupType(n.Name)
		return ret, err
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
		panic(n)
		return nil, fmt.Errorf("invalid type %v", n)
	}

	panic("unreachable")
}

func (e *emitter) parseStructToGType(n *parse.Struct) (GType, error) {
    ret := &GStruct{}
    for idx,name := range n.Names {
        t,err := e.parseNodeToGType(n.Types[idx])
        if err != nil {
            return nil,err
        }
        ret.Names = append(ret.Names,name)
        ret.Types = append(ret.Types,t)
    }
    return ret,nil
}

func gTypeToLLVM(t GType) string {
	switch t := t.(type) {
	case *GVoid:
		return "void"
	case *GStruct:
	    ret := "type {"
	    for idx,subt := range t.Types {
	        ret += gTypeToLLVM(subt)
	        if idx != len(t.Types) - 1 {
	            ret += ","
	        }
	    }
	    ret += "}"
	    return ret
	case *GPointer:
		return fmt.Sprintf("%s*", gTypeToLLVM(t.PointsTo))
    case *GArray:
		return fmt.Sprintf("[%d * %s]", t.Dim,gTypeToLLVM(t.SubType))
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
		panic("unreachable: bad gtype " + fmt.Sprintf("%v", t))
	}
}
