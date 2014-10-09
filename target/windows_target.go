package target

type X86_Windows_Target struct{}

func (*X86_Windows_Target) LLVMTargetTriple() string {
	return "i686-pc-mingw32"
}

func (*X86_Windows_Target) DefaultIntBitWidth() uint {
	return 64
}
