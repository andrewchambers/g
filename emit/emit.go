package emit

import (
	"bufio"
	"fmt"
	"strings"
	"github.com/andrewchambers/g/parse"
)


type emitter struct {
    curscope *scope
    out *bufio.Writer
    
    curFuncType *GFunc
}



type exprValue struct {
    llvmVal string
    lval bool
    gType GType
}

func newEmitter(out *bufio.Writer) *emitter {
    ret := &emitter{}
    ret.curscope = newScope(nil)
    ret.out = out
    return ret
}

func (e *emitter) pushScope() {
    e.curscope = newScope(e.curscope)
}

func (e *emitter) popScope() {
    e.curscope = e.curscope.parent
}

func (e *emitter) emit(s string,args ...interface{}) {
    fmt.Fprintf(e.out,s,args...)
}

func (e *emitter) emiti(s string,args ...interface{}) {
    e.emit("    "+s,args...)
}


func EmitModule(out *bufio.Writer, file *parse.File) error {
    e := newEmitter(out)
    err := e.collectGlobalSymbols(file)
    if err != nil {
        return err
    }
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
        impdef := imp.Val[1:len(imp.Val) - 1]
        impdef = strings.Split(impdef,"/")[0]
        sym := newSymbol(imp.Span.Start)
        err := e.curscope.declareSym(impdef,sym)
        if err != nil {
           return fmt.Errorf("bad import: %s %s:%v",err , imp.Span.Path, imp.Span.Start)
        }
    }
    for _, fd := range file.FuncDecls {
        sym := newSymbol(fd.Span.Start)
        err := e.curscope.declareSym(fd.Name,sym)
        if err != nil {
            return fmt.Errorf("bad func decl: %s %s:%v",err , fd.Span.Path, fd.Span.Start)
        }
    }
    for _,vd := range file.VarDecls {
        sym := newSymbol(vd.Span.Start)
        err := e.curscope.declareSym(vd.Name,sym)
        if err != nil {
            return fmt.Errorf("bad var decl: %s %s:%v",err , vd.Span.Path, vd.Span.Start)
        }
    }
    return nil
}

func (e *emitter) emitFuncDecl(f *parse.FuncDecl) error {
	//Emit function start
	ft,err := e.funcDeclToGType(f)
	if err != nil {
	    return err
	}
	e.curFuncType = ft 
	e.emit("define %s @%s() {\n","i32", f.Name)
	for _,stmt := range f.Body {
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
    //XXX type check
    e.emiti("ret %s\n",v.llvmVal)
}

func (e *emitter) emitExpression(expr parse.Node) *exprValue {
    switch expr := expr.(type) {
        case *parse.Constant:
            v := &exprValue{}
            v.llvmVal = fmt.Sprintf("i32 %s",expr.Val)
            v.gType = NewGInt(32,true)
            return v
        default:
            panic("unhandled...")
    }
}

func (e *emitter) funcDeclToGType(f *parse.FuncDecl) (*GFunc,error) {
    ret := &GFunc{}
    for _,t := range f.ArgTypes {
        gty,err := e.parseNodeToGType(t)
        if err != nil {
            return nil,err
        }
        ret.ArgTypes = append(ret.ArgTypes,gty)
    }
    if f.RetType != nil {
        t,err := e.parseNodeToGType(f.RetType)
        if err != nil {
            return nil,err
        }
        ret.RetType = t
    }
    
    return ret,nil
}

func (e *emitter) parseNodeToGType(n parse.Node) (GType,error) {
    span := n.GetSpan()
    var err error
    var ret GType
    switch n := n.(type) {
        case *parse.TypeAlias:
            ret,err = e.curscope.lookupType(n.Name)
    }
    if err != nil {
        return nil,fmt.Errorf("expected type (%s) at %s:%v.",err ,span.Path,span.Start)
    }
    return ret,nil
}

