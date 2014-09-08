package emit

type GType interface {
}

type GAlias struct {
	Name   string
	Actual GType
}

type GStruct struct {
	Names []string
	Types []GType
}

type GInt struct {
	Bits   uint
	Signed bool
}

func NewGInt(bits uint,signed bool) *GInt {
    return &GInt{bits,signed}
}


type GPointer struct {
	PointsTo GType
}

type GFunc struct {
	RetType  GType
	ArgTypes []GType
}

type GArray struct {
	Dim     int64
	ArrayOf GType
}
