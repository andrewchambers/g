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

type exprConstant struct {
	val int64
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

func (c *exprConstant) getLLVMRepr() string {
	return fmt.Sprintf("%v", c.val)
}

func (c *exprConstant) isLVal() bool {
	return false
}

func (c *exprConstant) getGType() GType {
	return &GConstant{}
}

func newEmitter(out *bufio.Writer) *emitter {
	ret := &emitter{}
	ret.curscope = newScope(nil)
	ret.gscope = ret.curscope
	ret.out = out
	ret.addBuiltinTypes()
	return ret
}

func (e *emitter) addBuiltinTypes() {
	e.curscope.declareType("int", &GInt{e.getIntWidth(), true})
	e.curscope.declareType("uint", &GInt{e.getIntWidth(), false})
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
		sym := newSymbol(imp.Span.Start)
		err := e.curscope.declareSym(impdef, sym)
		if err != nil {
			return fmt.Errorf("bad import: %s %s:%v", err, imp.Span.Path, imp.Span.Start)
		}
	}
	for _, fd := range file.FuncDecls {
		sym := newSymbol(fd.Span.Start)
		err := e.curscope.declareSym(fd.Name, sym)
		if err != nil {
			return fmt.Errorf("bad func decl: %s %s:%v", err, fd.Span.Path, fd.Span.Start)
		}
	}
	for _, vd := range file.VarDecls {
		sym := newSymbol(vd.Span.Start)
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
	case *parse.Return:
		e.emitReturn(stmt)
	default:
		panic("unhandled Statement type...")
	}
}

func (e *emitter) emitReturn(r *parse.Return) {
	v := e.emitExpression(r.Expr)
	var llvmType string
	_, ok := v.getGType().(*GConstant)
	if ok {
		_, ok := e.curFuncType.RetType.(*GInt)
		if !ok {
			// XXX
			panic("cannot cast constant to return type")
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

func (e *emitter) emitExpression(expr parse.Node) Value {
	switch expr := expr.(type) {
	case *parse.Constant:
		v := &exprConstant{expr.Val}
		return v
	case *parse.Binop:
		return e.emitBinop(expr)
	default:
		panic("unhandled...")
	}
}

func isConstantVal(v Value) bool {
	_, ok := v.(*exprConstant)
	return ok
}

func isIntType(t GType) bool {
	_, ok := t.(*GInt)
	return ok
}

func (e *emitter) emitBinop(b *parse.Binop) Value {

	l := e.emitExpression(b.L)
	r := e.emitExpression(b.R)

	if isConstantVal(l) && isConstantVal(r) {
		c, err := foldConstantBinop(b.Op, l.(*exprConstant), r.(*exprConstant))
		if err != nil {
			panic(err)
		}
		return c
	}

	lstr := fmt.Sprintf("%s %s", gTypeToLLVM(l.getGType()), l.getLLVMRepr())
	rstr := fmt.Sprintf("%s %s", gTypeToLLVM(r.getGType()), r.getLLVMRepr())

	if !l.getGType().Equals(r.getGType()) {
		panic("arithmetic on incompatible types")
	}

	if !isIntType(l.getGType()) {
		panic("arithmetic on non int type")
	}

	switch b.Op {
	case '+':
		e.emiti("new = add %s %s\n", lstr, rstr)
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
	span := n.GetSpan()
	var err error
	var ret GType
	switch n := n.(type) {
	case *parse.TypeAlias:
		ret, err = e.curscope.lookupType(n.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("expected type (%s) at %s:%v.", err, span.Path, span.Start)
	}
	return ret, nil
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
		panic("unreachable")
	}
}
