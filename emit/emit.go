package emit

import (
	"bufio"
	"github.com/andrewchambers/g/parse"
	"github.com/andrewchambers/g/resolve"
	"github.com/andrewchambers/g/target"
)

func EmitModule(machine target.TargetMachine, out *bufio.Writer, files []*parse.File) error {
	r := resolve.New()
	r.ResolvePackage(files)
	return nil
}
