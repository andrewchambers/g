package main

import (
	"flag"
	"fmt"
	"github.com/andrewchambers/g/parse"
	"github.com/andrewchambers/g/emit"
	"github.com/andrewchambers/g/llvm"
	"io"
	"bufio"
	"os"
	"runtime/pprof"
)

func printVersion() {
	fmt.Println("g version 0.01")
}

func printUsage() {
	printVersion()
	fmt.Println()
	fmt.Println("This software is a g compiler - https://github.com/andrewchambers/g")
	fmt.Println("It emits text llvm bytecode.")
	fmt.Println()
	fmt.Println("Software by Andrew Chambers 2014 - andrewchamberss@gmail.com")
	fmt.Println()
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	tokenizeOnly := flag.Bool("T", false, "Tokenize only (For debugging).")
	parseOnly := flag.Bool("A", false, "Print AST (For debugging).")
	doProfiling := flag.Bool("P", false, "Profile the compiler (For debugging).")
	version := flag.Bool("version", false, "Print version info and exit.")
	outputPath := flag.String("o", "-", "File to write output to, - for stdout.")
	flag.Parse()

	if *doProfiling {
		profile, err := os.Create("ccrun.prof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open profile file: %s\n", err)
			os.Exit(1)
		}
		pprof.StartCPUProfile(profile)
		defer pprof.StopCPUProfile()
	}

	if *version {
		printVersion()
		return
	}

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Bad number of args, please specify a single source file.\n")
		os.Exit(1)
	}

	input := flag.Args()[0]
	var output io.WriteCloser
	var err error

	if *outputPath == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(*outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open output file %s\n", err)
			os.Exit(1)
		}
	}
	defer output.Close()

	if *tokenizeOnly {
		tokenizeFile(input, output)
	} else if *parseOnly {
		ast := parseFile(input)
		fmt.Fprintln(output,ast.Dump(0))
	} else {
		compileFile(input, output)
	}
}

func tokenizeFile(sourceFile string, out io.WriteCloser) {
	f, err := os.Open(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open source file %s for lexing: %s\n", sourceFile, err)
		os.Exit(1)
	}
	tokChan := parse.Lex(sourceFile, f)
	for tok := range tokChan {
		if tok == nil {
			return
		}
		if tok.Kind == parse.ERROR {
			fmt.Fprintln(os.Stderr, tok.Val)
			os.Exit(1)
		}
		fmt.Fprintf(out, "%s:%s:%d:%d\n", tok.Kind, tok.Val, tok.Span.Start.Line, tok.Span.Start.Col)
	}
}

func parseFile(sourceFile string) *parse.File {
	f, err := os.Open(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open source file %s for lexing: %s\n", sourceFile, err)
		os.Exit(1)
	}
	tokChan := parse.Lex(sourceFile, f)
	ast,err := parse.Parse(tokChan)
	if err != nil {
	    fmt.Fprintf(os.Stderr, "parse error: %s\n", err)
	    os.Exit(1)
	}
	return ast
}

func compileFile(sourceFile string, out io.WriteCloser) {
	ast := parseFile(sourceFile)
	mod := emit.EmitModule(ast)
	llvm.EmitLLVM(out,mod)
}
