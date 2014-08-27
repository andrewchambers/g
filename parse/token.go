package parse

type Token struct {
    Kind TokenKind
    Span FileSpan
}

type FilePos struct {
    Path String
    Pos int
}

type FileSpan struct {
    Start FilePos
    End FilePos
}
