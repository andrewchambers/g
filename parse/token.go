package parse

type TokenKind int

type Token struct {
    Kind TokenKind
    Span FileSpan
}

type FilePos struct {
    Path string
    Pos int
}

type FileSpan struct {
    Start FilePos
    End FilePos
}
