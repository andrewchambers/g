package main

import (
    "flag"
    "fmt"
    "parse"
)



func printVersion() {
	fmt.Println("g version 0.01")
}

func printUsage() {
	printVersion()
	fmt.Println()
	fmt.Println("This software is a g compiler - https://github.com/andrewchambers/g")
	fmt.Println("It was created with the goals of being the small and hackable.")
	fmt.Println("It is hopefully a playground for the language design")
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

	if *preprocessOnly {
		//preprocessFile(input, output)
	} else if *tokenizeOnly {
		//tokenizeFile(input, output)
	} else if *parseOnly {
		//parseFile(input, output)
	} else {
		//compileFile(input, nil, output)
	}
}
