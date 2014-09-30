package target


type TargetMachine interface {
    LLVMTargetTriple() string
}
