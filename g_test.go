package main

// Test the compiler with the standard test tools. This lets us get coverage,
// profiling, and other nice things.

import (
	"fmt"
	"github.com/andrewchambers/g/driver"
	"github.com/andrewchambers/g/target"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func checkClangIsWorking() error {
	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempdir)

	var progsrc []byte = []byte("int main() { return 0; }\n")
	srcPath := path.Join(tempdir, "prog.c")
	err = ioutil.WriteFile(srcPath, progsrc, 0777)
	if err != nil {
		return err
	}
	cmd := exec.Command("clang", srcPath, "-o", path.Join(tempdir, "prog"))
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

type testResult struct {
	name string
	err  error
}

func makeFailedTestResult(name string, err string, args ...interface{}) testResult {
	return testResult{
		name,
		fmt.Errorf(err, args...),
	}
}

const retzerotestdir = "./gtestcases/retzero/singlefile/"

func runSingleFileRetZero(t *testing.T, testpath string) (result testResult) {

	t.Logf("running test %s testpath")

	// Recover and log failure on panic.
	defer func() {
		v := recover()
		if v != nil {
			result = makeFailedTestResult(testpath, "test panic (%v)", v)
		}
	}()

	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		result = makeFailedTestResult(testpath, "failed to create tempdir (%s)", err)
		return
	}
	defer os.RemoveAll(tempdir)

	llPath := path.Join(tempdir, "test.ll")
	_ = path.Join(tempdir, "test")
	outfile, err := os.Create(llPath)
	if err != nil {
		result = makeFailedTestResult(testpath, "failed to create file %s (%s)", outfile, err)
		return
	}
	tm := target.GetTarget()
	err = driver.CompilePackageToLLVM(tm, testpath, outfile)
	if err != nil {
		result = makeFailedTestResult(testpath, "failed to compile file (%s)", err)
		return
	}
	binPath := path.Join(tempdir, "test.ll")
	err = driver.LinkLLVMToBinary(llPath, binPath)
	if err != nil {
		result = makeFailedTestResult(testpath, "failed to link file (%s)", err)
		return
	}

	cmd := exec.Command(binPath)
	err = cmd.Run()
	if err != nil {
		result = makeFailedTestResult(testpath, "non zero exit code (%s)", err)
		return
	}

	result = testResult{testpath, nil}
	return
}

// Run all the SingleFileRetZero tests in parallel.
func TestSingleFileRetZero(t *testing.T) {

	t.Skip("skipping due to refactor.")

	err := checkClangIsWorking()
	if err != nil {
		t.Fatalf("clang failed to run %s", err)
		return
	}

	files, err := ioutil.ReadDir(retzerotestdir)
	if err != nil {
		t.Fatal("failed to read directory containing single file retzero tests.")
		return
	}

	predicate := func(info os.FileInfo) bool {
		if info.IsDir() {
			return false
		}
		if !strings.HasSuffix(info.Name(), ".g") {
			return false
		}
		return true
	}

	for _, info := range files {
		if !predicate(info) {
			continue
		}

		tr := runSingleFileRetZero(t, path.Join(retzerotestdir, info.Name()))

		if tr.err != nil {
			t.Errorf("%s failed. %s", tr.name, tr.err)
		}

	}

}
