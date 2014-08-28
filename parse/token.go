package parse

type TokenKind int

const (
    TOK_ERROR = iota
)

// The type representing a G lexical token.
type Token struct {
    // Tag for the token kind.
    Kind TokenKind
    // The raw contents of the token.
    Val string
    // The file span of the token.
    Span FileSpan
}

// Represents a single point in the file as shown in most text editors.
// Tabs are treated as being aligned to 4 places.
type FilePos struct {
    //Path 
    Path string
    // Line number starting at one.
    Line int
    // Col starting at one.
    Col int
}

// A span across two points.
type FileSpan struct {
    Start FilePos
    End FilePos
}
