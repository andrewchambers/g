package util

import (
	"fmt"
	"reflect"
)

type Iterator interface {
	Next() (interface{}, bool)
}

type arrayIter struct {
	idx uint
	end uint
	arr reflect.Value
}

type FilterFunc func(interface{}) bool

type filterIter struct {
	sub  Iterator
	filt FilterFunc
}

func (it *arrayIter) Next() (interface{}, bool) {
	if it.idx == it.end {
		return nil, true
	}
	ret := it.arr.Index(int(it.idx))
	it.idx += 1
	return ret.Interface(), false
}

func CreateIterator(arr interface{}) Iterator {
	v := reflect.ValueOf(arr)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return &arrayIter{
			0,
			uint(v.Len()),
			v,
		}
	default:
		panic(fmt.Sprintf("cannot create Iterator type out of %s",v.Kind()))
	}
}

func (it *filterIter) Next() (interface{}, bool) {
	for {
		v, done := it.sub.Next()
		if done {
			return v, done
		}
		if it.filt(v) {
			return v, done
		}
	}
}

func FilterIterator(it Iterator, filt FilterFunc) Iterator {
	return &filterIter{it, filt}
}
