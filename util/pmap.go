package util

type chanCountIter struct {
    n uint
    c chan interface{}
}

func (it *chanCountIter) Next() (interface{},bool) {
    if it.n == 0 {
        return nil,true
    }
    it.n -= 1
    return <-it.c,false
}


func PMap(it Iterator, transform func (v interface {}) interface{}) Iterator {
    resultChan := make(chan interface{})
    count := 0
    for {
        v,end := it.Next()
        if end {
            break
        }
        go func (v interface{}) {
            result := transform(v)
            resultChan <- result
        } (v)
        count += 1
    }
    
    return &chanCountIter {
        uint(count),
        resultChan,
    }
}
