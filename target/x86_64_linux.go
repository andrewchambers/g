package target

type X86_64_Linux_Target struct {

}


func (*X86_64_Linux_) LLVMTargetTriple() string {
    return "x86_64-pc-linux-gnu"
}
