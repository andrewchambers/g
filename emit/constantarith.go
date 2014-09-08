package emit

import "github.com/andrewchambers/g/parse"

func foldConstantBinop(op parse.TokenKind,l,r *exprConstant) *exprConstant {
    return &exprConstant{"1337"}
}
