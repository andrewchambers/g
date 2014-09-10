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
			p.syntaxError("expected var, type, const or func", p.curTok.Span)
		}
	}
}

func (p *parser) parseVarDecl() *VarDecl {
	ret := &VarDecl{}
	ret.Span = p.curTok.Span
	p.expect(VAR)
	ret.Name = p.curTok.Val
	p.expect(IDENTIFIER)
	ret.Type = p.parseType(false)
	if p.curTok.Kind == '=' {
		p.next()
		ret.Init = p.parseExpression()
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
	case STRUCT:
		return p.parseStruct()
	case IDENTIFIER:
		ret := &TypeAlias{}
		ret.Span = p.curTok.Span
		ret.Name = p.curTok.Val
		p.next()
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
}

func (p *parser) parseConst() {
	p.expect(CONST)
}

func (p *parser) parseStatementList(sl *[]Node) {
	for p.curTok.Kind != '}' && p.curTok.Kind != EOF {
		s := p.parseStatement()
		*sl = append(*sl,s)
	}
}

func (p *parser) parseStatement() Node {
	switch p.curTok.Kind {
	case RETURN:
		r := &Return{}
		r.Span = p.curTok.Span
		p.next()
		r.Expr = p.parseExpression()
		r.Span.End = p.curTok.Span.End
		p.expect(';')
		return r
	case VAR:
		ret := p.parseVarDecl()
		p.expect(';')
		return ret
	case IDENTIFIER, CONSTANT, STRING:
		ret := p.parseSimpleStatement()
		p.expect(';')
		return ret
	case FOR:
		p.parseFor()
	case IF:
		ret := p.parseIf()
		return ret
	default:
		p.syntaxError("error parsing statement", p.curTok.Span)
	}
	panic("unreachable")
}

func (p *parser) parseStruct() Node {
	ret := &Struct{}
	ret.span = p.curTok.Span
	p.expect(STRUCT)
	p.expect('{')
	for p.curTok.Kind == IDENTIFIER {
		ret.names = append(ret.names, p.curTok.Val)
		p.next()
		ret.types = append(ret.types, p.parseType(false))
	}
	ret.span.End = p.curTok.Span.End
	p.expect('}')
	return ret
}

func (p *parser) parseSimpleStatement() Node {
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
	}
	return ret

}

func (p *parser) parseFor() {
	p.expect(FOR)
	if p.curTok.Kind != '{' {
		p.parseSimpleStatement()
	}
	if p.curTok.Kind == ';' {
		p.next()
		p.parseExpression()
		p.expect(';')
		p.parseSimpleStatement()
	}
	p.expect('{')
	p.parseStatementList(nil)
	p.expect('}')
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
	return strconv.ParseInt(t.Val, 10, 64)
}

func (p *parser) parsePrimaryExpression() Node {
	var ret Node = nil
	switch p.curTok.Kind {
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
		args = append(args, p.parseExpression())
		if p.curTok.Kind == ',' {
			p.next()
		}
	}
	call.Span.End = p.curTok.Span.End
	p.expect(')')
	call.Args = args
	return call
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
