package main

func sub(z *int, x *int,y *int) {
    *z = *x - *y
    return;
}

func main() int {
    var x int = 3
    var y int = 4
    var z int = 0
    sub(&z, &x, &y);
    return z + 1
}
