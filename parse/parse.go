package parse

type yaccAdapter struct {
	lastTok *Token
	c       chan *Token
	onError func(string, FileSpan)
}

func (ya *yaccAdapter) Error(e string) {
	if ya.lastTok == nil {
		panic("unreachable")
	}
	ya.onError(e, ya.lastTok.Span)
}

func (ya *yaccAdapter) Lex(lval *yySymType) int {
	t := <-ya.c
	if t == nil {
		return 0
	}
	ya.lastTok = t
	return int(t.Kind)
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
	l := &yaccAdapter{nil, c, onError}
	yyParse(l)
}
