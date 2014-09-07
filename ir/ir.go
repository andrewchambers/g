package ir

import (
	"github.com/andrewchambers/g/types"
)

type Module struct {
	ReturnType types.GType
	Functions []Function
}

type Function struct {
	Entry *BasicBlock
}

type BasicBlock struct {
	Instructions []Instruction
}

type Instruction interface {
}

type Value interface {
	GetGType()
}