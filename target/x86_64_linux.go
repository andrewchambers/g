package target

type X86_64_Linux_Target struct {}

func (*X86_64_Linux_Target) LLVMTargetTriple() string {
    return "x86_64-pc-linux-gnu"
}

func (*X86_64_Linux_Target) DefaultIntBitWidth() uint {
    return 64
}
