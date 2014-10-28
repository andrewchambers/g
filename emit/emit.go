package emit

import (
    	"github.com/andrewchambers/g/parse"
    	"github.com/andrewchambers/g/target"
    	"github.com/andrewchambers/g/resolve"
    	"bufio"
)

func EmitModule(machine target.TargetMachine, out *bufio.Writer, files []*parse.File) error {
    
    r := resolve.New()
    r.ResolvePackage(files)
    
    return nil
}
