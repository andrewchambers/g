package util

import (
	"testing"
)

func TestPMap(t *testing.T) {
	arr := [...]int{1, 2}
	it := CreateIterator(arr)
	it = PMap(it, func(i interface{}) interface{} { return i })
	v, done := it.Next()
	if v.(int) != 1 && done != false {
		t.Errorf("bad values %v %v", v, done)
	}
	v, done = it.Next()
	if v.(int) != 2 && done != false {
		t.Errorf("bad values %v %v", v, done)
	}
	v, done = it.Next()
	if v != nil && done != true {
		t.Errorf("expected end or iterator")
	}
}
