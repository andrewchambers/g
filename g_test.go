package main

// Test the compiler with the standard test tools. This lets us get coverage,
// profiling, and other nice things.

import (
	"github.com/andrewchambers/g/driver"
	"github.com/andrewchambers/g/target"
	"github.com/andrewchambers/g/util"
	"io/ioutil"
	"os"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func checkClangIsInstalled() error {

    tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempdir)
	
    var progsrc []byte = []byte("int main() { return 0; }\n")
    srcPath := path.Join(tempdir,"prog.c")
    err = ioutil.WriteFile(srcPath,progsrc,0777)
    if err != nil {
        return err
    }
    cmd := exec.Command("gcc",srcPath, "-o", path.Join(tempdir,"prog"))
    err = cmd.Run()
    if err != nil {
        return err
    }
    return nil
}

type testResult struct {
    name string
    err error
}

func makeFailedTestResult(name string, err string, args ...interface{}) testResult {
    return testResult{
        name,
        fmt.Errorf(err,args...),
    }
}

const retzerotestdir = "./gtestcases/retzero/singlefile/"

func runSingleFileRetZero(itestpath interface {}) (result interface{}) {
	testpath := path.Join(retzerotestdir,itestpath.(os.FileInfo).Name())
	// Recover and log failure on panic.
	defer func () {
	    v := recover()
	    if v != nil {
	        result = makeFailedTestResult(testpath,"test panic (%v)", v)
	    }
	}()
	
	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		result = makeFailedTestResult(testpath,"failed to create tempdir (%s)", err)
		return
	}
	defer os.RemoveAll(tempdir)

	llPath := path.Join(tempdir, "test.ll")
	_ = path.Join(tempdir, "test")
	outfile, err := os.Create(llPath)
	if err != nil {
		result = makeFailedTestResult(testpath,"failed to create file %s (%s)", outfile, err)
		return
	}
	tm := target.GetTarget()
	err = driver.CompileFileToLLVM(tm, testpath, outfile)
	if err != nil {
		 result = makeFailedTestResult(testpath,"failed to compile file (%s)", err)
		 return
	}
	result = testResult{testpath,nil}
	return
}


// Run all the SingleFileRetZero tests in parallel.
func TestSingleFileRetZero(t *testing.T) {

    err := checkClangIsInstalled()
    if err != nil {
        t.Fatalf("clang failed to run %s",err)
        return
    }
    
	files, err := ioutil.ReadDir(retzerotestdir)
	if err != nil {
		t.Fatal("failed to read directory containing single file retzero tests.")
		return
	}
	iter := util.CreateIterator(files)
	iter = util.FilterIterator(iter,func (v interface{}) bool {
	    info,ok := v.(os.FileInfo)
	    if !ok {
	        return false
	    }
	    if info.IsDir() {
	        return false
	    }
	    if !strings.HasSuffix(info.Name(), ".g") {
            return false
		}
		return true
	})
	results := util.PMap(iter,runSingleFileRetZero)
	
	for {
	    v,done := results.Next()
	    if done {
	        break
	    }
	    tr := v.(testResult)
        if tr.err != nil {
            t.Errorf("Failed %s %s",tr.name,tr.err)
        }
	}
}
