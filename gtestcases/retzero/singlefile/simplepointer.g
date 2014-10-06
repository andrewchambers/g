package main

func main() int {
    var x int
    var y *int
    x = 1
    y = &x
    *y = *y - 1
    return *y
}
