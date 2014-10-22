package parse


type Node interface {
	GetSpan() FileSpan
}

type Package struct {
	Files []*File
}

type File struct {
	SpanProvider
	Pkg string
	//List of imports in the translation unit.
	Imports    []*String
	FuncDecls  []*FuncDecl
	TypeDecls  []*TypeDecl
	ConstDecls []*ConstDecl
	VarDecls   []*VarDecl
}

type For struct {
	SpanProvider
	Init, Cond, Step Node
	Body             []Node
}

type If struct {
	SpanProvider
	Cond Node
	Body []Node
	Els  []Node
}

type Selector struct {
	SpanProvider
	Name string
	Expr Node
}

type Unop struct {
	SpanProvider
	Op   TokenKind
	Expr Node
}

type Call struct {
	SpanProvider
	FuncLike Node
	Args     []Node
}

type TypeAlias struct {
	SpanProvider
	Name string
}

type PointerTo struct {
	SpanProvider
	PointsTo Node
}

type ArrayOf struct {
	SpanProvider
	Dim     uint
	SubType Node
}

type IndexInto struct {
	SpanProvider
	Index Node
	Expr  Node
}

type Struct struct {
	SpanProvider
	Names []string
	Types []Node
}

type Binop struct {
	SpanProvider
	Op   TokenKind
	L, R Node
}

type EmptyStatement struct {
	SpanProvider
}

type ExpressionStatement struct {
	SpanProvider
	Expr Node
}

type Assign struct {
	SpanProvider
	Op   TokenKind
	L, R Node
}

type Constant struct {
	SpanProvider
	Val int64
}

type Ident struct {
	SpanProvider
	Val string
}

type String struct {
	SpanProvider
	Val string
}

type VarDecl struct {
	SpanProvider
	Name string
	Type Node
	Init *Assign
}

type TypeDecl struct {
	SpanProvider
	Name string
	Type Node
}

type ConstDecl struct {
	SpanProvider
	Name string
	Body Node
}

type FuncDecl struct {
	SpanProvider
	Name     string
	RetType  Node
	ArgNames []string
	ArgTypes []Node
	Body     []Node
}

type Return struct {
	SpanProvider
	Expr Node
}

type SpanProvider struct {
	// The file span of the token.
	Span FileSpan
}

func (s *SpanProvider) GetSpan() FileSpan {
	return s.Span
}

func ws(depth uint) string {
	//Nicer way to do this?
	ret := ""
	for depth != 0 {
		ret += " "
		depth -= 1
	}
	return ret
}
func (n *File) addImport(s *String) {
	n.Imports = append(n.Imports, s)
}

func (n *File) addFuncDecl(f *FuncDecl) {
	n.FuncDecls = append(n.FuncDecls, f)
}

func (n *File) addTypeDecl(t *TypeDecl) {
	n.TypeDecls = append(n.TypeDecls, t)
}

func (n *File) addConstDecl(c *ConstDecl) {
	n.ConstDecls = append(n.ConstDecls, c)
}

func (n *File) addVarDecl(v *VarDecl) {
	n.VarDecls = append(n.VarDecls, v)
}

func (n *FuncDecl) addArgument(name string, t Node) {
	n.ArgNames = append(n.ArgNames, name)
	n.ArgTypes = append(n.ArgTypes, t)
}

func (n *FuncDecl) addStatement(s Node) {
	n.Body = append(n.Body, s)
}

