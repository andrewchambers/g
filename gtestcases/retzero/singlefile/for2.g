package main

func main() int {
    var x int
    for {
        x = x + 1
        if x == 5 {
            return x - 5
        }
    }
    return 1
}
