package main

import (
	"flag"
	"fmt"
	"github.com/andrewchambers/g/driver"
	"github.com/andrewchambers/g/parse"
	"github.com/andrewchambers/g/target"
	"github.com/andrewchambers/g/util"
	"io"
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
	fmt.Println("It emits text llvm bytecode as plain text.")
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

	isInputDir, err := util.IsDirectory(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with input. %s\n", err)
	}

	var output io.WriteCloser

	if *outputPath == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(*outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open output file %s\n", err)
			os.Exit(1)
		}
		defer output.Close()
	}
	if *tokenizeOnly {
		err := driver.TokenizeFile(input, output)
		if err != nil {
			fmt.Println(err)
			fmt.Println("tokenizer failed.")
			os.Exit(1)
		}
	} else if *parseOnly {
		var astList []*parse.File
		var err error
		if isInputDir {
			astList, err = driver.ParseFolder(input)
		} else {
			var ast *parse.File
			ast, err = driver.ParseFile(input)
			astList = []*parse.File{ast}
		}
		if err != nil {
			fmt.Println(err)
			fmt.Println("parsing failed.")
			os.Exit(1)
		}
		for _, ast := range astList {
			parse.DebugDump(output, ast)
		}
	} else {
		t := target.GetTarget()
		err := driver.CompilePackageToLLVM(t, input, output)
		if err != nil {
			fmt.Println(err)
			fmt.Println("compilation to llvm failed.")
			os.Exit(1)
		}
	}
}
