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
func (p *parser) parseError(message string, span FileSpan) {
	p.onError(message, span)
	panic(&breakout{})
}

func (p *parser) next() {
	// Lex error occured
	if p.nextTok != nil && p.nextTok.Kind == ERROR {
		p.parseError(p.nextTok.Val, p.nextTok.Span)
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
		p.parseError(fmt.Sprintf("unexpected token %s, expected %s", p.curTok.Val, k), p.curTok.Span)
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
			for p.curTok.Kind == STRING_LITERAL {
				p.next()
			}
			p.expect(')')
		case STRING_LITERAL:
			p.next()
		default:
			p.parseError("expected string literal or '('", p.curTok.Span)
		}
	}
}

func (p *parser) parseDeclarations() {
	for p.curTok.Kind != EOF {
		switch p.curTok.Kind {
		case FUNC:
			p.parseFuncDecl()
		case VAR:
			p.parseVarDecl()
		case CONST:
			p.parseConst()
		default:
			p.parseError("expected var, const or func", p.curTok.Span)
		}
	}
}

func (p *parser) parseVarDecl() {

}

func (p *parser) parseFuncDecl() {
	p.expect(FUNC)
	p.expect(IDENTIFIER)
	p.expect('(')
	p.parseArgList()
	p.expect(')')
	p.expect('{')
	p.parseStatementList()
	p.expect('}')
}

func (p *parser) parseArgList() {

}

func (p *parser) parseConst() {

}

func (p *parser) parseStatementList() {

}
