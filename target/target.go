package target

import (
	"fmt"
	"runtime"
)

type TargetMachine interface {
	// The LLVM triple the target uses
	LLVMTargetTriple() string
	// The native width of machine registers
	// This is used for default int size, and default array index type.
	DefaultIntBitWidth() uint
}

func GetTarget() TargetMachine {
	switch runtime.GOOS {
	case "linux":
		return &X86_64_Linux_Target{}
	default:
		panic(fmt.Sprintf("unknown platform %s", runtime.GOOS))
	}
}
