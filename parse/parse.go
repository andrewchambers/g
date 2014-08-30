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

func (p *parser) parseType(allowEmpty bool) {
	switch p.curTok.Kind {
	case STRUCT:
		p.parseStruct()
	case IDENTIFIER:
		p.next()
	default:
		if allowEmpty {
			return
		}
		p.syntaxError("expected a type", p.curTok.Span)
	}
}

func (p *parser) parseFuncReturnType() {
	p.parseType(true)
}

func (p *parser) parseArgList() {

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
	case VAR:
		p.parseVarDecl()
	case IDENTIFIER, CONSTANT, STRING_LITERAL:
		p.parseExpression()
	case FOR:
		p.parseFor()
	case IF:
		p.parseIf()
	default:
		p.syntaxError("error parsing statement", p.curTok.Span)
	}
	p.expect(';')
}

func (p *parser) parseFor() {
	p.expect(FOR)
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

func (p *parser) parseStruct() {
	p.expect(STRUCT)
	p.expect('{')
	p.expect('}')
}

func (p *parser) parseExpression() {
	p.parsePrimaryExpression()
}

func (p *parser) parsePrimaryExpression() {
	switch p.curTok.Kind {
	case IDENTIFIER:
		p.next()
	case CONSTANT:
		p.next()
	case STRING_LITERAL:
		p.next()
	default:
		p.syntaxError("error parsing primary expression", p.curTok.Span)
	}
}
