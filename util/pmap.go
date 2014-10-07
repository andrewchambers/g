package util

import (
	"runtime"
	"sync"
)

type chanCountIter struct {
	c chan interface{}
}

func (it *chanCountIter) Next() (interface{}, bool) {
	v := <-it.c
	if v == nil {
		return nil, true
	}
	return v, false
}

func PMap(it Iterator, transform func(v interface{}) interface{}, poolSize int) Iterator {

	if poolSize <= 0 {
		poolSize = runtime.NumCPU()
	}

	var wg sync.WaitGroup
	jobs := make(chan interface{})
	resultChan := make(chan interface{})

	workerProc := func() {
		for {
			v := <-jobs
			if v == nil {
				wg.Done()
				break
			}
			xformed := transform(v)
			resultChan <- xformed
		}
	}

	for poolSize >= 0 {
		poolSize -= 1
		wg.Add(1)
		go workerProc()
	}

	go func() {
		for {
			v, end := it.Next()
			if end {
				close(jobs)
				break
			}
			jobs <- v
		}
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return &chanCountIter{
		resultChan,
	}
}
