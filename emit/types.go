package emit

type GType interface {
	Equals(GType) bool
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

func NewGInt(bits uint, signed bool) *GInt {
	return &GInt{bits, signed}
}

func (i *GInt) Equals(other GType) bool {
	oint, ok := other.(*GInt)
	if !ok {
		return false
	}
	return i.Bits == oint.Bits && i.Signed == oint.Signed
}

func (i *GFunc) Equals(other GType) bool {
	_, ok := other.(*GFunc)
	if !ok {
		return false
	}
	panic("unimplemented")
}
