package main

// Test the compiler with the standard test tools. This lets us get coverage,
// profiling, and other nice things.


import (
    "testing"
    "io/ioutil"
)



// Compile a Single file to an llvm text file, returns the result to a channel.

func compileSingleFileToLLVM() {

}

func compileSingleFileToBinary() {

}

func assembleAndLink() {

}

func TestSingleFileRetZero(t *testing.T) {
    files,err := ioutil.ReadDir("gtestcases/retzero/singlefile/")
    if err != nil {
        t.Fatal("failed to read directory containing single file retzero tests.")
        return
    }
    for _,_ = range files {
        
    }
}
