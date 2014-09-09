package emit

type GType interface {
	String() string
	Equals(GType) bool
}

type GStruct struct {
	Names []string
	Types []GType
}

type GInt struct {
	Alias  string
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

type GConstant struct {
}

type GArray struct {
	Dim     int64
	ArrayOf GType
}

func NewGInt(alias string, bits uint, signed bool) *GInt {
	return &GInt{alias, bits, signed}
}

func isBool(t GType) bool {
	v, ok := t.(*GInt)
	if ok {
		return v.Bits == 1
	}
	return false
}

func (*GConstant) Equals(other GType) bool {
	_, ok := other.(*GConstant)
	return ok
}

func (c *GConstant) String() string {
	return "constant"
}

func (i *GInt) Equals(other GType) bool {
	oint, ok := other.(*GInt)
	if !ok {
		return false
	}
	return i.Bits == oint.Bits && i.Signed == oint.Signed
}

func (i *GInt) String() string {
	if i.Alias != "" {
		return i.Alias
	}

	if i.Signed {
		switch i.Bits {
		case 64:
			return "int64"
		case 32:
			return "int32"
		case 16:
			return "int16"
		case 8:
			return "int8"
		case 1:
			return "bool"
		}
	} else {
		switch i.Bits {
		case 64:
			return "uint64"
		case 32:
			return "uint32"
		case 16:
			return "uint16"
		case 8:
			return "uint8"
		case 1:
			return "bool"
		}
	}
	panic("internal error")
}

func (f *GFunc) Equals(other GType) bool {
	_, ok := other.(*GFunc)
	if !ok {
		return false
	}
	panic("unimplemented")
}

func (f *GFunc) String() string {
	return "function"
}
