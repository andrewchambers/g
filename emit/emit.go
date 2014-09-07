package emit

import (
	"bufio"
	"fmt"
	"github.com/andrewchambers/g/parse"
)

func EmitModule(out *bufio.Writer,file *parse.File ) {
}

func emitFunction(out *bufio.Writer, f *parse.FuncDecl) {
	//Emit function start
	fmt.Fprintln(out, "define %s @%s {", "","")
	//emit body as basic blocks
	//XXX
	fmt.Fprintln(out, "}")
}
