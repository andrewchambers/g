package emit

import (
	"fmt"
	"github.com/andrewchambers/g/parse"
)

func foldConstantBinop(op parse.TokenKind, l, r *exprConstant) (*exprConstant, error) {
	ret := &exprConstant{}
	switch op {
	case '+':
		ret.val = l.val + r.val
	case '-':
		ret.val = l.val - r.val
	default:
		return ret, fmt.Errorf("unhandled binary operator %s", op)
	}
	return ret, nil
}
