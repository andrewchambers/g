package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func GFilesInDir(dir string) ([]string, error) {
	ret := make([]string, 0, 16)
	finfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return ret, err
	}
	for _, f := range finfo {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".g") {
			continue
		}
		ret = append(ret, path.Join(dir, f.Name()))
	}
	return ret, nil
}

func IsDirectory(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		return true, nil
	}
	return false, nil
}

type ErrorList []error

func (e ErrorList) Error() string {
	ret := ""
	for idx := range e {
		ret += fmt.Sprintf("%s", e[idx])
		if idx != len(e)-1 {
			ret += "\n"
		}
	}
	return ret
}
