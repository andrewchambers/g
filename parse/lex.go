package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// breakout is a dummy type just used for leaving the lexing loop with panic.
type breakout struct{}

type lexer struct {
	// Marked token start position in current file.
	markedPos FilePos
	// The io source the lexer is reading for.
	brdr *bufio.Reader
	// Current position information
	curLine int
	curCol  int
	//Position information saved for unreads sake
	prevLine int
	prevCol  int
	// Path of current file (used for errors and other position info).
	path string
	// Output only channel for sending tokens.
	out chan *Token
	//Set to true if we have hit eof
	eof bool
	// Set to true if the last token emitted was one of the semicolon hack tokens.
	// This is what allows g to omit semicolons when they would otherwise be needed.
	semiHack bool
}

// Lexers the reader in a goroutine.
// The goroutine must be read until completion or error.
// The Error token is returned on lex error.
func Lex(path string, r io.Reader) chan *Token {
	out := make(chan *Token, 1024)
	l := new(lexer)
	l.brdr = bufio.NewReader(r)
	l.path = path
	l.curLine = 1
	l.curCol = 1
	l.prevLine = 1
	l.prevCol = 1
	l.out = out
	go l.lex()
	return out
}

// Saves the current lexer position in markedPos.
func (l *lexer) mark() {
	l.markedPos = l.currentPos()
}

func (l *lexer) currentPos() FilePos {
	return FilePos{l.curLine, l.curCol}
}

func isSemiColonInjectToken(k TokenKind) bool {
    switch k {
        case IDENTIFIER,CONSTANT,STRING,BREAK,CONTINUE,RETURN,')','}':
            return true
    }
    return false
}

// Saves the current lexer position in markedPos.
func (l *lexer) sendTok(k TokenKind, val string) {
    if isSemiColonInjectToken(k) {
        l.semiHack = true
    } else {
        l.semiHack = false
    }
	l.out <- &Token{k, val, FileSpan{l.path, l.markedPos, l.currentPos()}}
}

// Panics with aborting error type, does not return
func (l *lexer) lexError(message string) {
	l.sendTok(ERROR, fmt.Sprintf("Error while lexing: "+message))
	panic(&breakout{})
}

// Returns the next rune, and a bool representing eof
func (l *lexer) readRune() (rune, bool) {

	l.prevLine = l.curLine
	l.prevCol = l.curCol
	r, _, err := l.brdr.ReadRune()
	if err != nil {
		if err == io.EOF {
			l.eof = true
			return 0, true
		}
		l.lexError(err.Error())
	}
	switch r {
	case '\n':
		l.curCol = 1
		l.curLine += 1
	case '\t':
		l.curCol = (l.curCol - 1) + 4
		l.curCol -= l.curCol % 4
		l.curCol += 1
	default:
		l.curCol += 1
	}
	return r, false
}

//Cannot go back more than one rune ever.
func (l *lexer) unreadRune() {
	if l.eof {
		return
	}
	//I dont think we will ever unget a \t or \n.
	l.curCol = l.prevCol
	l.curLine = l.prevLine
	l.brdr.UnreadRune()
}

func (l *lexer) lex() {
	//This recovery happens if lexError is called.
	defer func() {
		if e := recover(); e != nil {
			_ = e.(*breakout) // Will re-panic if not a breakout.
		}

	}()
	defer close(l.out)

	for {
		l.mark()
		first, eof := l.readRune()
		if eof {
			break
		}
		switch {
		case isAlpha(first) || first == '_':
			l.unreadRune()
			l.readIdentOrKeyword()
		case isNumeric(first):
			l.unreadRune()
			l.readConstantIntOrFloat()
		case isWhiteSpace(first):
			l.unreadRune()
			l.skipWhiteSpace()
		default:
			switch first {
			case '(':
				l.sendTok('(', "(")
			case ')':
				l.sendTok(')', ")")
			case '[':
				l.sendTok('[', "[")
			case ']':
				l.sendTok(']', "]")
			case ';':
				l.sendTok(';', ";")
			case ',':
				l.sendTok(',', ",")
			case '{':
				l.sendTok('{', "{")
			case '}':
				l.sendTok('}', "}")
			case '.':
				l.sendTok('.', ".")
			case '"':
				l.unreadRune()
				l.readStringLiteral()
			case '=':
				next, _ := l.readRune()
				switch next {
				case '=':
					l.sendTok(EQ, "==")
				default:
					l.unreadRune()
					l.sendTok('=', "=")
				}
			case '|':
				next, _ := l.readRune()
				switch next {
				case '|':
					l.sendTok(OR, "||")
				case '=':
					l.sendTok(ORASSIGN, "|=")
				default:
					l.unreadRune()
					l.sendTok('|', "|")
				}
			case '&':
				next, _ := l.readRune()
				switch next {
				case '&':
					l.sendTok(AND, "&&")
				case '=':
					l.sendTok(ANDASSIGN, "&=")
				case '^':
					l.sendTok(ANDNOT, "&^")
				default:
					l.unreadRune()
					l.sendTok('&', "&")
				}
			case '/':
				next, _ := l.readRune()
				switch next {
				case '/':
					//Line comment
					for {
						next, eof := l.readRune()
						if next == '\n' || eof {
							l.maybeDoSemiHack(next)
							break
						}
					}
				case '*':
					l.skipUntilBlockCommentTerminator()
				default:
					l.unreadRune()
					l.sendTok('/', "/")
				}
			case '^':
				next, _ := l.readRune()
				switch next {
				case '=':
					l.sendTok(XORASSIGN, "^=")
				default:
					l.unreadRune()
					l.sendTok('^', "^")
				}
			case '%':
				next, _ := l.readRune()
				switch next {
				default:
					l.unreadRune()
					l.sendTok('%', "%")
				}
			case '+':
				next, _ := l.readRune()
				switch next {
				case '+':
					l.sendTok(INC, "++")
				case '=':
					l.sendTok(ADDASSIGN, "+=")
				default:
					l.unreadRune()
					l.sendTok('+', "+")
				}
			case '-':
				next, _ := l.readRune()
				switch next {
				case '-':
					l.sendTok(DEC, "--")
				case '=':
					l.sendTok(SUBASSIGN, "-=")
				default:
					l.unreadRune()
					l.sendTok('-', "-")
				}
			case '*':
				next, _ := l.readRune()
				switch next {
				case '=':
					l.sendTok(MULASSIGN, "+=")
				default:
					l.unreadRune()
					l.sendTok('*', "*")
				}
			case '<':
				next, _ := l.readRune()
				switch next {
				case '<':
					l.sendTok(LSHIFT, "<<")
				case '=':
					l.sendTok(LTEQ, "<=")
				default:
					l.unreadRune()
					l.sendTok('<', "<")
				}
			case '>':
				next, _ := l.readRune()
				switch next {
				case '>':
					l.sendTok(RSHIFT, ">>")
				case '=':
					l.sendTok(GTEQ, ">=")
				default:
					l.unreadRune()
					l.sendTok('>', ">")
				}
			default:
				l.lexError(fmt.Sprintf("unknown character %d", first))
			}
		}
	}
}

var keywordLUT = map[string]TokenKind{
	"func":    FUNC,
	"return":  RETURN,
	"package": PACKAGE,
	"struct":  STRUCT,
	"import":  IMPORT,
	"for":     FOR,
	"if":      IF,
	"break":   BREAK,
	"continue":CONTINUE,
	"else":    ELSE,
	"type":    TYPE,
	"var":     VAR,
	"const":   CONST,
}

func (l *lexer) skipUntilBlockCommentTerminator() {
	for {
		c, eof := l.readRune()
		if eof {
			l.lexError("unclosed block comment.")
		}
		l.maybeDoSemiHack(c)
		if c == '*' {
			closeBar, eof := l.readRune()
			if eof {
				l.lexError("unclosed block comment.")
			}
			if closeBar == '/' {
				break
			}
			l.unreadRune()
		}
	}
}

func (l *lexer) readIdentOrKeyword() {
	var buff bytes.Buffer
	l.mark()
	first, _ := l.readRune()
	if !isValidIdentStart(first) {
		panic("internal error")
	}
	buff.WriteRune(first)
	for {
		b, _ := l.readRune()
		if isValidIdentTail(b) {
			buff.WriteRune(b)
		} else {
			l.unreadRune()
			str := buff.String()
			tokType, ok := keywordLUT[str]
			if !ok {
				tokType = IDENTIFIER
			}
			l.sendTok(tokType, str)
			break
		}
	}
}

// Need to validate these with a regex or another mechanism
func (l *lexer) readConstantIntOrFloat() {
	var buff bytes.Buffer
	l.mark()
	first, _ := l.readRune()
	if !isNumeric(first) {
		panic("internal error")
	}
	buff.WriteRune(first)
	for {
		b, _ := l.readRune()
		if isHexDigit(b) || b == '.' {
			buff.WriteRune(b)
		} else {
			l.unreadRune()
			str := buff.String()
			l.sendTok(CONSTANT, str)
			break
		}
	}
}

// Need to validate these with a regex or another mechanism
func (l *lexer) readStringLiteral() {
	var buff bytes.Buffer
	l.mark()
	first, _ := l.readRune()
	if first != '"' {
		panic("internal error")
	}
	buff.WriteRune(first)
	for {
		b, eof := l.readRune()
		if eof {
			l.lexError("eof while reading string literal")
		}
		buff.WriteRune(b)
		if b == '"' {
			break
		}
	}
	l.sendTok(STRING, buff.String())
}

func (l *lexer) maybeDoSemiHack(r rune) {
    if r == '\n' && l.semiHack {
	    l.sendTok(';',";")
	    l.semiHack = false
	}
}
		

func (l *lexer) skipWhiteSpace() {
	for {
		r, _ := l.readRune()
		l.maybeDoSemiHack(r)
		if !isWhiteSpace(r) {
			l.unreadRune()
			break
		}
	}
}

func isValidIdentTail(b rune) bool {
	return isValidIdentStart(b) || isNumeric(b)
}

func isValidIdentStart(b rune) bool {
	return b == '_' || isAlpha(b)
}

func isAlpha(b rune) bool {
	if b >= 'a' && b <= 'z' {
		return true
	}
	if b >= 'A' && b <= 'Z' {
		return true
	}
	return false
}

func isWhiteSpace(b rune) bool {
	return b == ' ' || b == '\r' || b == '\n' || b == '\t' || b == '\f'
}

func isNumeric(b rune) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	return false
}

func isHexDigit(b rune) bool {
	return isNumeric(b) || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
