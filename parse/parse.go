package parse

import (
	"fmt"
)

type parser struct {
	curTok  *Token
	nextTok *Token
	c       chan *Token
	onError func(string, FileSpan)
}

func Parse(c chan *Token, onError func(string, FileSpan)) {
	//Read channel until empty incase of errors
	defer func() {
		for {
			x := <-c
			if x == nil {
				break
			}
		}
	}()
	defer func() {
		if e := recover(); e != nil {
			_ = e.(*breakout) // Will re-panic if not a breakout.
		}

	}()
	p := &parser{c: c, onError: onError}
	p.next()
	p.next()
	p.parseTranslationUnit()
}

// Panics with aborting error type, does not return
func (p *parser) syntaxError(message string, span FileSpan) {
	p.onError(message, span)
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

func (p *parser) parseTranslationUnit() {
	p.expect(PACKAGE)
	p.expect(IDENTIFIER)
	p.parseImportList()
	p.parseDeclarations()
}

func (p *parser) parseImportList() {
	for p.curTok.Kind == IMPORT {
		p.next()
		switch p.curTok.Kind {
		case '(':
			p.next()
			for p.curTok.Kind == STRING_LITERAL {
				p.next()
			}
			p.expect(')')
		case STRING_LITERAL:
			p.next()
		default:
			p.syntaxError("expected string literal or '('", p.curTok.Span)
		}
	}
}

func (p *parser) parseDeclarations() {
	for p.curTok.Kind != EOF {
		switch p.curTok.Kind {
		case TYPE:
			p.parseTypeDecl()
		case FUNC:
			p.parseFuncDecl()
		case VAR:
			p.parseVarDecl()
		case CONST:
			p.parseConst()
		default:
			p.syntaxError("expected var, const or func", p.curTok.Span)
		}
	}
}

func (p *parser) parseVarDecl() {
	p.expect(VAR)
	p.expect(IDENTIFIER)
	p.parseType(false)
	if p.curTok.Kind == '=' {
		p.next()
		p.parseExpression()
	}
}

func (p *parser) parseTypeDecl() {
	p.expect(TYPE)
	p.expect(IDENTIFIER)
	p.parseType(false)
}

func (p *parser) parseFuncDecl() {
	p.expect(FUNC)
	p.expect(IDENTIFIER)
	p.expect('(')
	p.parseArgList()
	p.expect(')')
	p.parseFuncReturnType()
	p.expect('{')
	p.parseStatementList()
	p.expect('}')
}

func (p *parser) parseType(allowEmpty bool) ASTNode {
	switch p.curTok.Kind {
	case STRUCT:
		return p.parseStruct()
	case IDENTIFIER:
		ret := &ASTIdent{}
		ret.span = p.curTok.Span
		ret.val = p.curTok.Val
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

func (p *parser) parseFuncReturnType() {
	p.parseType(true)
}

func (p *parser) parseArgList() {
	for p.curTok.Kind == IDENTIFIER {
		p.next()
		p.parseType(false)
		if p.curTok.Kind == ',' {
			p.next()
		}
	}
}

func (p *parser) parseConst() {
	p.expect(CONST)
}

func (p *parser) parseStatementList() {
	for p.curTok.Kind != '}' && p.curTok.Kind != EOF {
		p.parseStatement()
	}
}

func (p *parser) parseStatement() {
	switch p.curTok.Kind {
	case RETURN:
		p.next()
		p.parseExpression()
		p.expect(';')
	case VAR:
		p.parseVarDecl()
		p.expect(';')
	case IDENTIFIER, CONSTANT, STRING_LITERAL:
		p.parseSimpleStatement()
		p.expect(';')
	case FOR:
		p.parseFor()
	case IF:
		p.parseIf()
	default:
		p.syntaxError("error parsing statement", p.curTok.Span)
	}
}

func (p *parser) parseSimpleStatement() {
	p.parseExpression()
	switch p.curTok.Kind {
	case '=', ADDASSIGN, MULASSIGN:
		p.next()
		p.parseExpression()
	case INC, DEC:
		p.next()
	default:
	}

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
	p.parseStatementList()
	p.expect('}')
}

func (p *parser) parseIf() {
	p.expect(IF)
	p.parseExpression()
	p.expect('{')
	p.parseStatementList()
	p.expect('}')
	if p.curTok.Kind == ELSE {
		p.next()
		switch p.curTok.Kind {
		case IF:
			p.parseIf()
		case '{':
			p.expect('{')
			p.parseStatementList()
			p.expect('}')
		default:
			p.syntaxError("If ", p.curTok.Span)
		}
	}
}

func (p *parser) parseStruct() ASTNode {
	ret := &ASTStruct{}
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

func (p *parser) parseExpression() ASTNode {
	return p.parsePrec1()
}

func (p *parser) parsePrec1() ASTNode {
	l := p.parsePrec2()
	for {
		switch p.curTok.Kind {
		case OR:
			p.next()
			r := p.parsePrec2()
			n := &ASTBinop{}
			n.op = p.curTok.Kind
			n.l = l
			n.r = r
			n.span = l.GetSpan()
			n.span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func (p *parser) parsePrec2() ASTNode {
	l := p.parsePrec3()
	for {
		switch p.curTok.Kind {
		case AND:
			p.next()
			r := p.parsePrec3()
			n := &ASTBinop{}
			n.op = p.curTok.Kind
			n.l = l
			n.r = r
			n.span = l.GetSpan()
			n.span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func (p *parser) parsePrec3() ASTNode {
	l := p.parsePrec4()
	for {
		switch p.curTok.Kind {
		case EQ, NEQ, '<', LTEQ, '>', GTEQ:
			p.next()
			r := p.parsePrec4()
			n := &ASTBinop{}
			n.op = p.curTok.Kind
			n.l = l
			n.r = r
			n.span = l.GetSpan()
			n.span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}
func (p *parser) parsePrec4() ASTNode {
	l := p.parsePrec5()
	for {
		switch p.curTok.Kind {
		case '+', '-', '|', '^':
			p.next()
			r := p.parsePrec5()
			n := &ASTBinop{}
			n.op = p.curTok.Kind
			n.l = l
			n.r = r
			n.span = l.GetSpan()
			n.span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func (p *parser) parsePrec5() ASTNode {
	l := p.parsePrimaryExpression()
	for {
		switch p.curTok.Kind {
		case '*', '/', '%', LSHIFT, RSHIFT, '&':
			p.next()
			r := p.parsePrimaryExpression()
			n := &ASTBinop{}
			n.op = p.curTok.Kind
			n.l = l
			n.r = r
			n.span = l.GetSpan()
			n.span.End = r.GetSpan().End
			l = n
		default:
			return l
		}
	}
}

func (p *parser) parsePrimaryExpression() ASTNode {
	switch p.curTok.Kind {
	case IDENTIFIER:
		ret := &ASTIdent{}
		ret.val = p.curTok.Val
		ret.span = p.curTok.Span
		p.next()
		return ret
	case CONSTANT:
		ret := &ASTConstant{}
		ret.val = p.curTok.Val
		ret.span = p.curTok.Span
		p.next()
		return ret
	case STRING_LITERAL:
		ret := &ASTString{}
		ret.val = p.curTok.Val
		ret.span = p.curTok.Span
		p.next()
		return ret
	default:
		p.syntaxError("error parsing expression", p.curTok.Span)
	}
	return nil
}
