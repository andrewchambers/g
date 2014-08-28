package parse

import (
    "io"
)

type lexer struct {
    // Marked token start position in current file.
    markedPos *FilePos
    // The io source the lexer is reading for.
    brdr *bufio.Reader
    // Current position information
    curLine int
    curCol int
    // Path of current file (used for errors and other position info).
    path string
    // Output only channel for sending tokens.
    stream chan *Token
}

// Lexers the reader in a goroutine.
// The goroutine must be read until completion or error.
// Lexing errors will call onError.
func Lex(path string, r io.Reader, onError func (error,FilePos)) chan *Token {
    
}

// Saves the current lexer position in markedPos.
func (l *Lexer) mark() {
    l.markedPos = l.currentPos()
}

func (l *Lexer) currentPos() *FilePos {
    return &FilePos{l.path,l.curLine,l.curCol}
}

// Saves the current lexer position in markedPos.
func (l *Lexer) sendTok(k TokenKind,val string) {
    start := l.markedPos
    end := l.currentPos()
    span := &FileSpan{start,end}
    return &Token{k,val,span}
}



func (l *Lexer) lex() {
    for {
    
    }
}
