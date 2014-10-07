package util

import (
	"testing"
)

func TestArrayIterator(t *testing.T) {
	arr := [...]int{1, 2}
	it := CreateIterator(arr)
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

func TestFilterIterator(t *testing.T) {
	arr := [...]int{1, 2}
	it := CreateIterator(arr)
	it = FilterIterator(it, func(v interface{}) bool {
		if v.(int) == 2 {
			return false
		}
		return true
	})
	v, done := it.Next()
	if v.(int) != 1 && done != false {
		t.Errorf("bad values %v %v", v, done)
	}
	v, done = it.Next()
	if v != nil && done != true {
		t.Errorf("expected end or iterator")
	}
}
