package main

// Test the compiler with the standard test tools. This lets us get coverage,
// profiling, and other nice things.

import (
	"github.com/andrewchambers/g/driver"
	"github.com/andrewchambers/g/target"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestSingleFileRetZero(t *testing.T) {
	testdir := "./gtestcases/retzero/singlefile/"
	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		t.Fatal("failed to read directory containing single file retzero tests.")
		return
	}
	for _, finfo := range files {

		if finfo.IsDir() {
			continue
		}
		if strings.HasSuffix(finfo.Name(), ".g") {
			sourceFile := path.Join(testdir, finfo.Name())
			t.Logf("running test %s", sourceFile)
			tempdir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Errorf("failed to create tempdir for test %s (%s)", sourceFile, err)
				continue
			}
			defer os.RemoveAll(tempdir)

			llPath := path.Join(tempdir, "test.ll")
			_ = path.Join(tempdir, "test")
			outfile, err := os.Create(llPath)
			if err != nil {
				t.Errorf("failed to create file llPath for test %s (%s)", sourceFile, err)
				continue
			}
			tm := target.GetTarget()
			err = driver.CompileFileToLLVM(tm, sourceFile, outfile)
			if err != nil {
				t.Errorf("failed to compile file %s (%s)", sourceFile, err)
			}
		}
	}
}
