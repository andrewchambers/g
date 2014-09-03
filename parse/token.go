package parse

import "fmt"

type TokenKind int

const (
	ERROR = 0xffff + iota
	EOF
	FOR
	PACKAGE
	IMPORT
	FUNC
	RETURN
	STRUCT
	CONSTANT
	STRING
	IDENTIFIER
	VAR
	CONST
	TYPE
	IF
	ELSE
	NEQ
	EQ
	LTEQ
	GTEQ
	INC
	DEC
	ADDASSIGN
	SUBASSIGN
	MULASSIGN
	AND
	OR
	LSHIFT
	RSHIFT
)

func (k TokenKind) String() string {
	if k < ERROR {
		return fmt.Sprintf("%c", k)
	}

	var lut = map[TokenKind]string{
		FOR:     "for",
		PACKAGE: "package",
		IMPORT:  "import",
		RETURN:  "return",
		TYPE:    "type",
		IF:      "IF",
	}
	s, ok := lut[k]
	if ok {
		return s
	}
	return fmt.Sprintf("%d", k)
}

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
	// Line number starting at one.
	Line int
	// Col starting at one.
	Col int
}

// A span across two points.
type FileSpan struct {
	//Path
	Path  string
	Start FilePos
	End   FilePos
}
