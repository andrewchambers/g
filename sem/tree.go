package sem

//A function AST is converted into a type checked tree.

type TypeProvider struct {
	Type GType
}

func (tp *TypeProvider) GetType() GType {
	return tp.Type
}

type SemNode interface {
	GetType() GType
}

type SemFunction struct {
	Name       string
	Statements []SemNode
}

type SemReturn struct {
	Expr SemNode
}

type SemGCall struct {
	TypeProvider
	Label string
	args  []SemNode
}

type SemCall struct {
	TypeProvider
}

type SemString struct {
}
