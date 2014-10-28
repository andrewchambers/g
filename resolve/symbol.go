package resolve

const (
	LOCAL = iota
	GLOBAL
	ARG
	PACKAGE
)

type symbol interface{}

type lazySymbol struct {
	s *symbol
}

type localSymbol struct {
}

type constSymbol struct {
}

type globalSymbol struct {
}
