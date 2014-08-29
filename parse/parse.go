package parse

import (
    "fmt"
)

type yaccAdapter struct {
    lastTok *Token
    c chan *Token
}

func (ya *yaccAdapter) Error(e string) {
    fmt.Printf(e)
}

func (ya *yaccAdapter) Lex(lval *yySymType) int {
    t := <- ya.c
    ya.lastTok = t
    return int(t.Kind)
}


func Parse(c chan *Token) {
    //Read channel until empty incase of errors
    defer func() {
        for {
            x := <- c
            if x == nil {
                break
            }
        }
    }()
    l := &yaccAdapter{nil,c}
    yyParse(l)
}
