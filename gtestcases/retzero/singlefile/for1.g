package main

func main() int {
    var x int
    var y int
    x = 0
    y = 0
    for ; x < 10 ; x = x + 1 {
        y = y + 1
    }
    return x + y - 20
}
