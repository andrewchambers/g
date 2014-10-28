package emit

import (
	"fmt"
	"github.com/andrewchambers/g/parse"
)

func foldConstantUnop(op parse.TokenKind, v Value) (Value, error) {
	switch v := v.(type) {
	case *intConstant:
		switch op {
		case '-':
			return &intConstant{-v.val}, nil
		default:
			return nil, fmt.Errorf("unhandled unary operator %s", op)
		}
	default:
		return nil, fmt.Errorf("internal error (unhandled constant type)")
	}
}

func foldConstantBinop(op parse.TokenKind, l, r Value) (Value, error) {

	switch l := l.(type) {
	case *intConstant:
		r, ok := r.(*intConstant)
		if !ok {
			return nil, fmt.Errorf("mismatched types for %s operator", op)
		}
		switch op {
		case '+':
			return &intConstant{l.val + r.val}, nil
		case '&':
			return &intConstant{l.val & r.val}, nil
		case '^':
			return &intConstant{l.val ^ r.val}, nil
		case '-':
			return &intConstant{l.val - r.val}, nil
		case '*':
			return &intConstant{l.val * r.val}, nil
		case '%':
			if r.val == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return &intConstant{l.val % r.val}, nil
		case '/':
			if r.val == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return &intConstant{l.val / r.val}, nil
		case parse.ANDNOT:
			return &intConstant{l.val &^ r.val}, nil
		case parse.EQ:
			return &boolConstant{l.val == r.val}, nil
		default:
			return nil, fmt.Errorf("unhandled binary operator %s", op)
		}
	case *boolConstant:
		_, ok := r.(*boolConstant)
		if !ok {
			return nil, fmt.Errorf("mismatched types for %s operator", op)
		}
	default:
		return nil, fmt.Errorf("internal error (unhandled constant type)")
	}

	panic("unreachable")
}
