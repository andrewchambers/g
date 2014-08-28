package parse

import (
	"bufio"
	"io"
)

// breakout is a dummy type just used for leaving the lexing loop with panic.
type breakout struct {}

type lexer struct {
	// Marked token start position in current file.
	markedPos FilePos
	// The io source the lexer is reading for.
	brdr *bufio.Reader
	// Current position information
	curLine int
	curCol  int
	// Path of current file (used for errors and other position info).
	path string
	// Output only channel for sending tokens.
	out chan *Token
}

// Lexers the reader in a goroutine.
// The goroutine must be read until completion or error.
// The Error token is returned on lex error.
func Lex(path string, r io.Reader) chan *Token {
	out := make(chan *Token, 1024)
	return out
}

// Saves the current lexer position in markedPos.
func (l *lexer) mark() {
	l.markedPos = l.currentPos()
}

func (l *lexer) currentPos() FilePos {
	return FilePos{l.path, l.curLine, l.curCol}
}

// Saves the current lexer position in markedPos.
func (l *lexer) sendTok(k TokenKind, val string) {
	l.out <- &Token{k, val, FileSpan{l.markedPos, l.currentPos()}}
}

// Panics with aborting error type, does not return
func (l *lexer) lexError(message string) {
    
}

// Returns the next rune, and a bool representing eof
func (l *lexer) readRune() (rune, bool) {
	r, _, err := l.brdr.ReadRune()
	if err != nil {
		if err == io.EOF {
			return 0, true
		}
		l.lexError(err.Error())
	}
	switch r {
	case '\n':
	    l.curCol = 1
		l.curLine += 1
	case '\t':
		l.curCol += l.curCol + (4 - ((l.curCol-1) % 4))
	default:
		l.curCol += 1
	}
	return r, false
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
		switch first {
		
		}
	}
    
}
