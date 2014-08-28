package parse

import (
	"bufio"
	"io"
)

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

func (l *lexer) lex() {
	for {

	}
}
