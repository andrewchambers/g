package main

// Test the compiler with the standard test tools. This lets us get coverage,
// profiling, and other nice things.


import (
    "testing"
    "io/ioutil"
)



// Compile a Single file to an llvm text file, returns the result to a channel.


func TestSingleFileRetZero(t *testing.T) {
    files,err := ioutil.ReadDir("gtestcases/retzero/singlefile/")
    if err != nil {
        t.Fatal("failed to read directory containing single file retzero tests.")
        return
    }
    for _,finfo = range files {
        if !finfo.IsDir() && {
        
        
        	f, err := os.Open(sourceFile)
	        if err != nil {
		        fmt.Fprintf(os.Stderr, "Failed to open source file %s for lexing: %s\n", sourceFile, err)
		        os.Exit(1)
	        }
	        tokChan := parse.Lex(sourceFile, f)
	        ast, err := parse.Parse(tokChan)
	        if err != nil {
		        fmt.Fprintf(os.Stderr, "parse error: %s\n", err)
		        os.Exit(1)
	        }
	        return ast
	    
	    
    }
}
