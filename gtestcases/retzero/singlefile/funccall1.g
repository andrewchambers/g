package main

func foo() int {
    return 0
}

func main() int {
    var x int
    x = 1
    x = foo()
    return x
}
