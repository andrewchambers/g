package driver

import (
	"bufio"
	"fmt"
	"github.com/andrewchambers/g/emit"
	"github.com/andrewchambers/g/parse"
	"github.com/andrewchambers/g/target"
	"github.com/andrewchambers/g/util"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func TokenizeFile(sourceFile string, out io.WriteCloser) error {
	defer out.Close()
	f, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("Failed to open source file %s for lexing: %s\n", sourceFile, err)
	}
	defer f.Close()
	tokChan, _ := parse.Lex(sourceFile, f)
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

func GetPackageName(sourceFile string) (string, error) {
	f, err := os.Open(sourceFile)
	if err != nil {
		return "", fmt.Errorf("Failed to open source file %s for lexing: %s\n", sourceFile, err)
	}
	tokChan, cancel := parse.Lex(sourceFile, f)
	defer func() {
		cancel <- struct{}{}
	}()
	tok := <-tokChan
	if tok == nil || tok.Kind != parse.PACKAGE {
		return "", fmt.Errorf("malformed package statement")
	}
	tok = <-tokChan
	if tok == nil || tok.Kind != parse.IDENTIFIER {
		return "", fmt.Errorf("malformed package name")
	}
	return tok.Val, nil
}

func ParseFolder(folder string) ([]*parse.File, error) {
	rfile, rerr := make([]*parse.File, 0, 16), make(util.ErrorList, 0, 16)
	filePaths, err := util.GFilesInDir(folder)
	if err != nil {
		return rfile, err
	}
	for _, path := range filePaths {
		f, err := ParseFile(path)
		if err != nil {
			rerr = append(rerr, err)
		} else {
			rfile = append(rfile, f)
		}
	}
	if len(rerr) == 0 {
		return rfile, nil
	}
	return rfile, rerr
}

func ParseFile(sourceFile string) (*parse.File, error) {
	f, err := os.Open(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open source file %s for lexing: %s\n", sourceFile, err)
	}
	tokChan, _ := parse.Lex(sourceFile, f)
	ast, err := parse.Parse(tokChan)
	if err != nil {
		return nil, fmt.Errorf("parse error: %s", err)
	}
	return ast, nil
}

func CompilePackageToLLVM(machine target.TargetMachine, sourcePackage string, out io.WriteCloser) error {
	defer out.Close()
	ast, err := ParseFolder(sourcePackage)
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
	clangErrorPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	errorBytes, err := ioutil.ReadAll(clangErrorPipe)
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("%s", errorBytes)
	}
	return nil
}
