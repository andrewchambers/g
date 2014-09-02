package emit

//This package takes the output from the sem package, and emits text llvm assembly.
//One llvm module corresponds to one g package. Should aim to preserve debug information.
//We emit naive llvm with loads and stores to trivially preserve ssa without phi nodes.
//This module panics on error because there should be no errors at this stage given a well formed module.

import (
	"bufio"
	"fmt"
	"github.com/andrewchambers/g/sem"
)

//A type to store alloca info
type alloca struct {
	//An llvm register name
	regName string
	sz      int
}

type llvmFunctionEmitter struct {
	allocaMap map[*sem.Symbol]*alloca
}

func EmitModule(out *bufio.Writer, module []sem.SemNode) {
	for _, v := range module {
		switch v := v.(type) {
		case *sem.SemFunction:
			emitFunction(out, v)
		default:
			panic("unimplemented case.")
		}
	}
}

func emitFunction(out *bufio.Writer, f *sem.semFunction) {
	//Emit function start
	fmt.Fprintln(out, "define %s @%s {", gtypeToLLVMType(nil), f.Name)
	//emit body as basic blocks
	//XXX
	fmt.Fprintln(out, "}")
}
