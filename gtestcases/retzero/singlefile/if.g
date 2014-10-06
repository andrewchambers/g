package main

func main() int {
    
    var x int
    x = 0
    
    if 0 == 1 {
        x = 1
    } else if 1 == 2  {
        x = 2
    } else {
        x = 3
    }
    
    if 0 == 1 {
        x = 1
    } else if 1 == 1  {
        x = x - 2
    } else {
        x = 3
    }
    
    if 0 == 0 {
        x = x - 1
    } else if 1 == 2  {
        x = 2
    } else {
        x = 3
    }
    
    return x
}
