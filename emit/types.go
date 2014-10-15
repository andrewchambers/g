package emit

import (
	"fmt"
	"github.com/andrewchambers/g/target"
)

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

type GVoid struct {
}

type GPointer struct {
	PointsTo GType
}

type GArray struct {
	Dim     uint
	SubType GType
}

type GFunc struct {
	RetType  GType
	ArgTypes []GType
}

type GConstant struct {
}

var builtinVoidGType GType = &GVoid{}
var builtinBoolGType GType = &GInt{"bool", 1, false}
var builtinInt8GType GType = &GInt{"int8", 8, true}
var builtinInt16GType GType = &GInt{"int16", 16, true}
var builtinInt32GType GType = &GInt{"int32", 32, true}
var builtinInt64GType GType = &GInt{"int64", 64, true}
var builtinUInt8GType GType = &GInt{"int8", 8, false}
var builtinUInt16GType GType = &GInt{"uint16", 16, false}
var builtinUInt32GType GType = &GInt{"uint32", 32, false}
var builtinUInt64GType GType = &GInt{"uint64", 64, false}

func getDefaultIntType(tm target.TargetMachine) GType {
	switch tm.DefaultIntBitWidth() {
	case 32:
		return builtinInt32GType
	case 64:
		return builtinInt64GType
	}
	panic("internal error")
}

func isBool(t GType) bool {
	v, ok := t.(*GInt)
	if ok {
		return v.Bits == 1
	}
	return false
}

func (a *GArray) Equals(other GType) bool {
	o, ok := other.(*GArray)
	if !ok {
		return false
	}
	return a.Dim == o.Dim && o.SubType.Equals(a.SubType)
}

func (a *GArray) String() string {
	return fmt.Sprintf("[%d]%s", a.Dim, a.SubType.String())
}

func (p *GPointer) Equals(other GType) bool {
	o, ok := other.(*GPointer)
	if !ok {
		return false
	}
	return o.PointsTo.Equals(p.PointsTo)
}

func (p *GPointer) String() string {
	return fmt.Sprintf("*%s", p.PointsTo.String())
}

func (*GConstant) Equals(other GType) bool {
	_, ok := other.(*GConstant)
	return ok
}

func (c *GConstant) String() string {
	return "constant"
}

func (*GVoid) Equals(other GType) bool {
	_, ok := other.(*GVoid)
	return ok
}

func (c *GVoid) String() string {
	return "void"
}

func (i *GInt) Equals(other GType) bool {
	oint, ok := other.(*GInt)
	if !ok {
		return false
	}
	return i.Bits == oint.Bits && i.Signed == oint.Signed
}

func (s *GStruct) String() string {
	return "GStruct"
}

func (s *GStruct) Equals(other GType) bool {
	o, ok := other.(*GStruct)
	if !ok {
		return false
	}

	if len(o.Names) != len(s.Names) {
		return false
	}

	for idx, name := range s.Names {
		if o.Names[idx] != name {
			return false
		}
		if !s.Types[idx].Equals(o.Types[idx]) {
			return false
		}
	}

	return true
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
