package driver

import (
	"bufio"
	"fmt"
	"github.com/andrewchambers/g/emit"
	"github.com/andrewchambers/g/parse"
	"github.com/andrewchambers/g/target"
	"io"
	"ioutil"
	"os"
	"os/exec"
)

func TokenizeFile(sourceFile string, out io.WriteCloser) error {
	defer out.Close()
	f, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("Failed to open source file %s for lexing: %s\n", sourceFile, err)
	}
	tokChan := parse.Lex(sourceFile, f)
	for tok := range tokChan {
		if tok == nil {
			return nil
		}
		if tok.Kind == parse.ERROR {
			return fmt.Errorf(tok.Val)
		}
		fmt.Fprintf(out, "%s:%s:%d:%d\n", tok.Kind, tok.Val, tok.Span.Start.Line, tok.Span.Start.Col)
	}
	return nil
}

func ParseFile(sourceFile string) (*parse.File, error) {
	f, err := os.Open(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open source file %s for lexing: %s\n", sourceFile, err)
	}
	tokChan := parse.Lex(sourceFile, f)
	ast, err := parse.Parse(tokChan)
	if err != nil {
		return nil, fmt.Errorf("parse error: %s\n", err)
	}
	return ast, nil
}

func CompileFileToLLVM(machine target.TargetMachine, sourceFile string, out io.WriteCloser) error {
	defer out.Close()
	ast, err := ParseFile(sourceFile)
	if err != nil {
		return err
	}
	err = emit.EmitModule(machine, bufio.NewWriter(out), ast)
	if err != nil {
		return err
	}
	return nil
}

func LinkLLVMToBinary(llvmFile string, outFile string) error {
	cmd := exec.Command("clang", llvmFile, "-o", outFile)
	clangErrors, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	errStrings, err := ioutil.ReadAll(clangErrors)

	cmd.Wait()
	return nil
}
