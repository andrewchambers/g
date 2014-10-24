package parse

import (
	"fmt"
	"strconv"
)

type parser struct {
	curTok  *Token
	nextTok *Token
	c       chan *Token
	ast     *File
	err     error
}

func ParsePackage() (*Package, error) {
	return nil, nil
}

func Parse(c chan *Token) (*File, error) {
	//Read channel until empty incase of errors
	defer func() {
		for {
			x := <-c
			if x == nil {
				break
			}
		}
	}()
	p := &parser{c: c}
	p.next()
	p.next()
	p.parseFile()
	return p.ast, p.err
}

// Panics with aborting error type, does not return
func (p *parser) syntaxError(message string, span FileSpan) {
	p.err = fmt.Errorf("%s at %s:%d:%d", message, span.Path, span.Start.Line, span.Start.Col)
	panic(&breakout{})
}

func (p *parser) next() {
	// Lex error occured
	if p.nextTok != nil && p.nextTok.Kind == ERROR {
		p.syntaxError(p.nextTok.Val, p.nextTok.Span)
	}
	p.curTok = p.nextTok
	p.nextTok = <-p.c
	//On eof insert an EOF token
	if p.nextTok == nil {
		p.nextTok = &Token{EOF, "", p.curTok.Span}
		p.nextTok.Span.Start = p.nextTok.Span.End
	}
}

func (p *parser) expect(k TokenKind) {
	if p.curTok.Kind != k {
		p.syntaxError(fmt.Sprintf("unexpected token '%s', expected '%s'", p.curTok.Val, k), p.curTok.Span)
	}
	p.next()
}

func (p *parser) parseFile() {
	defer func() {
		if e := recover(); e != nil {
			_ = e.(*breakout) // Will re-panic if not a breakout.
		}

	}()
	p.ast = &File{}
	p.expect(PACKAGE)
	//This span is bogus, but a File is just the whole file.
	p.ast.Span = p.curTok.Span
	p.ast.Pkg = p.curTok.Val
	p.expect(IDENTIFIER)
	p.expect(';')
	p.parseImportList()
	p.parseTopLevelDeclarations()
}

func (p *parser) parseImportList() {
	for p.curTok.Kind == IMPORT {
		p.next()
		switch p.curTok.Kind {
		case '(':
			p.next()
			for p.curTok.Kind == STRING {
				p.ast.addImport(p.parseString())
			}
			p.expect(')')
		case STRING:
			p.ast.addImport(p.parseString())
		default:
			p.syntaxError("expected string literal or '('", p.curTok.Span)
		}
	}
}

func (p *parser) parseString() *String {
	ret := &String{}
	ret.Span = p.curTok.Span
	ret.Val = p.curTok.Val
	p.expect(STRING)
	return ret
}

func (p *parser) parseTopLevelDeclarations() {
	for p.curTok.Kind != EOF {
		switch p.curTok.Kind {
		case TYPE:
			t := p.parseTypeDecl()
			p.ast.addTypeDecl(t)
		case FUNC:
			f := p.parseFuncDecl()
			p.ast.addFuncDecl(f)
		case VAR:
			v := p.parseVarDecl()
			p.ast.addVarDecl(v)
		case CONST:
			p.parseConst()
		default:
			p.syntaxError(fmt.Sprintf("expected var, type, const or func got %s", p.curTok.Kind), p.curTok.Span)
		}
		p.expect(';')
	}
}

func (p *parser) parseVarDecl() *VarDecl {
	ret := &VarDecl{}
	ret.Span = p.curTok.Span
	p.expect(VAR)
	ret.Name = p.curTok.Val
	ident := &Ident{}
	ident.Span = p.curTok.Span
	ident.Val = p.curTok.Val
	p.expect(IDENTIFIER)
	ret.Type = p.parseType(false)
	if p.curTok.Kind == '=' {
		p.next()
		r := p.parseExpression()
		ret.Init = &Assign{}
		ret.Init.Op = '='
		ret.Init.R = r
		ret.Init.L = ident
		ret.Init.Span = ident.Span
		ret.Init.Span.End = r.GetSpan().End
	}
	return ret
}

func (p *parser) parseTypeDecl() *TypeDecl {
	ret := &TypeDecl{}
	ret.Span = p.curTok.Span
	p.expect(TYPE)
	ret.Name = p.curTok.Val
	p.expect(IDENTIFIER)
	ret.Type = p.parseType(false)
	return ret
}

func (p *parser) parseFuncDecl() *FuncDecl {
	ret := &FuncDecl{}
	p.expect(FUNC)
	ret.Name = p.curTok.Val
	p.expect(IDENTIFIER)
	p.expect('(')
	p.parseArgList(ret)
	p.expect(')')
	ret.RetType = p.parseType(true)
	p.expect('{')
	p.parseStatementList(&ret.Body)
	p.expect('}')
	return ret
}

func (p *parser) parseType(allowEmpty bool) Node {

	switch p.curTok.Kind {
	case '[':
		ret := &ArrayOf{}
		p.expect('[')
		if p.curTok.Kind != CONSTANT {
			//Trigger syntax error
			p.expect(CONSTANT)
		}
		c, err := tokToInt64(p.curTok)
		if err != nil {
			p.syntaxError(err.Error(), p.curTok.Span)
		}
		if c < 0 {
			p.syntaxError("negative array dimension", p.curTok.Span)
		}
		ret.Dim = uint(c)
		p.expect(CONSTANT)
		p.expect(']')
		t := p.parseType(false)
		ret.SubType = t
		return ret
	case STRUCT:
		return p.parseStruct()
	case IDENTIFIER:
		ret := &TypeAlias{}
		ret.Span = p.curTok.Span
		ret.Name = p.curTok.Val
		p.next()
		return ret
	case '*':
		ret := &PointerTo{}
		ret.Span = p.curTok.Span
		p.next()
		pointsTo := p.parseType(false)
		ret.PointsTo = pointsTo
		ret.Span.End = pointsTo.GetSpan().End
		return ret
	default:
		if allowEmpty {
			return nil
		}
		p.syntaxError("expected a type", p.curTok.Span)
	}
	return nil
}

func (p *parser) parseArgList(f *FuncDecl) {
	for p.curTok.Kind == IDENTIFIER {
		name := p.curTok.Val
		p.next()
		t := p.parseType(false)
		f.addArgument(name, t)
		if p.curTok.Kind == ',' {
			p.next()
		}
	}

	if p.curTok.Kind == ELLIPSIS {
		p.next()
		f.isVarArg = true
	}
}

func (p *parser) parseConst() {
	p.expect(CONST)
}

func (p *parser) parseStatementList(sl *[]Node) {
	for p.curTok.Kind != '}' && p.curTok.Kind != EOF {
		s := p.parseStatement()
		*sl = append(*sl, s)
	}
}

func (p *parser) parseStatement() Node {
	switch p.curTok.Kind {
	case RETURN:
		r := &Return{}
		r.Span = p.curTok.Span
		p.next()
		if p.curTok.Kind == ';' {
			p.next()
			r.Expr = nil
			return r
		}
		r.Expr = p.parseExpression()
		r.Span.End = p.curTok.Span.End
		p.expect(';')
		return r
	case VAR:
		ret := p.parseVarDecl()
		p.expect(';')
		return ret
	case FOR:
		ret := p.parseFor()
		return ret
	case IF:
		ret := p.parseIf()
		return ret
	default:
		ret := p.parseSimpleStatement()
		p.expect(';')
		return ret
	}
}

func (p *parser) parseStruct() Node {
	ret := &Struct{}
	ret.Span = p.curTok.Span
	p.expect(STRUCT)
	p.expect('{')
	for p.curTok.Kind == IDENTIFIER {
		ret.Names = append(ret.Names, p.curTok.Val)
		p.next()
		ret.Types = append(ret.Types, p.parseType(false))
		p.expect(';')
	}
	ret.Span.End = p.curTok.Span.End
	p.expect('}')
	return ret
}

func (p *parser) parseSimpleStatement() Node {
	if p.curTok.Kind == ';' {
		ret := &EmptyStatement{}
		ret.Span = p.curTok.Span
		return ret
	}
	ret := p.parseExpression()
	switch p.curTok.Kind {
	case '=', ADDASSIGN, MULASSIGN:
		ass := &Assign{}
		ass.Op = p.curTok.Kind
		ass.L = ret
		p.next()
		r := p.parseExpression()
		ass.R = r
		ass.Span = ass.L.GetSpan()
		ass.Span.End = ass.R.GetSpan().End
		ret = ass
	case INC, DEC:
		p.next()
	default:
		es := &ExpressionStatement{}
		es.Expr = ret
		es.Span = ret.GetSpan()
		ret = es
	}
	return ret

}

func (p *parser) parseFor() *For {
	ret := &For{}
	ret.Span = p.curTok.Span
	p.expect(FOR)

	if p.curTok.Kind == '{' {
		p.expect('{')
		p.parseStatementList(&ret.Body)
		p.expect('}')
		return ret
	}

	ret.Init = p.parseSimpleStatement()

	if p.curTok.Kind == '{' {
		ret.Cond = ret.Init
		es, ok := ret.Cond.(*ExpressionStatement)
		if !ok {
			p.syntaxError("expected an expression", ret.Cond.GetSpan())
		}
		ret.Cond = es.Expr
		ret.Init = nil
		p.expect('{')
		p.parseStatementList(&ret.Body)
		p.expect('}')
		return ret
	}
	p.expect(';')

	if p.curTok.Kind != ';' {
		ret.Cond = p.parseExpression()
	}
	p.expect(';')
	if p.curTok.Kind != '{' {
		ret.Step = p.parseSimpleStatement()
	}
	p.expect('{')
	p.parseStatementList(&ret.Body)
	p.expect('}')
	return ret
}

func (p *parser) parseIf() *If {
	ret := &If{}
	ret.Span = p.curTok.Span
	p.expect(IF)
	ret.Cond = p.parseExpression()
	p.expect('{')
	p.parseStatementList(&ret.Body)
	p.expect('}')
	if p.curTok.Kind == ELSE {
		p.next()
		switch p.curTok.Kind {
		case IF:
			ret.Els = []Node{p.parseIf()}
		case '{':
			p.expect('{')
			p.parseStatementList(&ret.Els)
			p.expect('}')
		default:
			p.syntaxError("If ", p.curTok.Span)
		}
	}
	return ret
}

func (p *parser) parseExpression() Node {
	return p.parsePrec1()
}

func (p *parser) parsePrec1() Node {
	l := p.parsePrec2()
	for {
		switch p.curTok.Kind {
		case OR:
			n := &Binop{}
			n.Op = p.curTok.Kind
			p.next()
			r := p.parsePrec2()
			n.L = l
			n.R = r
			n.Span = l.GetSpan()
			n.Span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func (p *parser) parsePrec2() Node {
	l := p.parsePrec3()
	for {
		switch p.curTok.Kind {
		case AND:
			n := &Binop{}
			n.Op = p.curTok.Kind
			p.next()
			r := p.parsePrec3()
			n.L = l
			n.R = r
			n.Span = l.GetSpan()
			n.Span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func (p *parser) parsePrec3() Node {
	l := p.parsePrec4()
	for {
		switch p.curTok.Kind {
		case EQ, NEQ, '<', LTEQ, '>', GTEQ:
			n := &Binop{}
			n.Op = p.curTok.Kind
			p.next()
			r := p.parsePrec4()
			n.L = l
			n.R = r
			n.Span = l.GetSpan()
			n.Span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}
func (p *parser) parsePrec4() Node {
	l := p.parsePrec5()
	for {
		switch p.curTok.Kind {
		case '+', '-', '|', '^':
			n := &Binop{}
			n.Op = p.curTok.Kind
			p.next()
			r := p.parsePrec5()
			n.L = l
			n.R = r
			n.Span = l.GetSpan()
			n.Span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func (p *parser) parsePrec5() Node {
	l := p.parsePrimaryExpression()
	for {
		switch p.curTok.Kind {
		case '*', '/', '%', LSHIFT, RSHIFT, '&', ANDNOT:
			n := &Binop{}
			n.Op = p.curTok.Kind
			p.next()
			r := p.parsePrimaryExpression()
			n.L = l
			n.R = r
			n.Span = l.GetSpan()
			n.Span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func tokToInt64(t *Token) (int64, error) {
	v, err := strconv.ParseInt(t.Val, 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (p *parser) parsePrimaryExpression() Node {

	var ret Node = nil

	switch p.curTok.Kind {
	case FUNC, STRUCT, '[':
		// Resolve is going to need to deal with converting Unop * to ptr.
		// Resolve also has to deal with the ambiguity of *foo(bar)
		// If foo is a type, its a cast, else its a call.
		// Confirmed for type cast, parse the type.
		ty := p.parseType(false)
		ret = ty
	case '&', '*', '-':
		newu := &Unop{}
		newu.Op = p.curTok.Kind
		newu.Span = p.curTok.Span
		p.next()
		expr := p.parsePrimaryExpression()
		newu.Expr = expr
		newu.Span.End = expr.GetSpan().End
		ret = newu
	case IDENTIFIER:
		v := &Ident{}
		v.Val = p.curTok.Val
		v.Span = p.curTok.Span
		p.next()
		ret = v
	case CONSTANT:
		v := &Constant{}
		c, err := tokToInt64(p.curTok)
		if err != nil {
			p.syntaxError(err.Error(), p.curTok.Span)
		}
		v.Val = c
		v.Span = p.curTok.Span
		p.next()
		ret = v
	case STRING:
		ret = p.parseString()
	case '(':
		p.expect('(')
		ret = p.parseExpression()
		p.expect(')')
	default:
		p.syntaxError("error parsing expression", p.curTok.Span)
	}

loop:
	for {
		switch p.curTok.Kind {
		case '(':
			ret = p.parseCall(ret)
		case '.':
			ret = p.parseSelector(ret)
		case '[':
			ret = p.parseIndex(ret)
		default:
			break loop
		}
	}

	return ret
}

func (p *parser) parseCall(funcLike Node) *Call {
	call := &Call{}
	call.FuncLike = funcLike
	call.Span = funcLike.GetSpan()
	var args []Node
	p.expect('(')
	for p.curTok.Kind != ')' && p.curTok.Kind != EOF {
		arg := p.parseExpression()
		args = append(args, arg)
		if p.curTok.Kind == ',' {
			p.next()
		}
	}
	call.Span.End = p.curTok.Span.End
	p.expect(')')
	call.Args = args
	return call
}

func (p *parser) parseIndex(l Node) *IndexInto {
	idx := &IndexInto{}
	idx.Span = l.GetSpan()
	idx.Expr = l
	p.expect('[')
	idx.Index = p.parseExpression()
	idx.Span.End = p.curTok.Span.End
	p.expect(']')
	return idx
}

func (p *parser) parseSelector(l Node) *Selector {
	sel := &Selector{}
	p.expect('.')
	sel.Span = l.GetSpan()
	sel.Name = p.curTok.Val
	sel.Expr = l
	sel.Span.End = p.curTok.Span.End
	p.expect(IDENTIFIER)
	return sel
}
