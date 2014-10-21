# G programming language

A new programming language - WIP

G is a refined C (based on Go) with a few cool things like packages and multiple returns.

G aims to...

* Be easy to parse and analyze like Go.
* Do anything C can do (unions, unsafe things).
* Be lean like C, should be suitable for things like kernels, etc.
* Have perfect interop with C.
* Have clear mapping to underlying machine code.
* Have inline assembly support.
* Use packages similar to Go and not headers.
* Have useful but minimalist extra features.

And more.

G is not trying to compete with big languages like rust or C++. It just trying to be as close to C as possible while mostly simplifying, and improving things. It is still just portable assembly.

# Examples (not all implemented):


Hello world:
```
package main

import "stdio"

func main() int {
    stdio.printf("hello world!\n")
    return 0
}
```


Multiple return:

```
type errcode int

func Foo() (int,errcode) {
    return 0,-1
}

...
  var v,err = Foo()
```
Structs and unions:
```
type s struct {
   x int
   y int
}

type u union {
   x int
   y int
}
```

Simple type inference:

```
type s struct {
   x int
   y int
}

...

var v = &s{x : 0 , y :1}

```
Tuple types and simple destructuring:
```
    var x = 1,2 // x is a tuple
    ...
    var x,y = 1,2 // x and y are destructured
    ...
    var x,y (int,byte) = 1,2 // explicit typing of tuple
    ...
    var x = (1,"foo")
    var y = x[1] // y is now type *char
```

Saner left to right declaration syntax:
```
// x is a function pointer which takes an int and a byte and returns a pointer to an array of 32 ints.
 var x func (int,byte) *[32]int
```


# Status

For a general idea of what currently works, look inside the gtestcases folder.

# Build and test:
[![Build Status](https://travis-ci.org/andrewchambers/g.svg?branch=master)](https://travis-ci.org/andrewchambers/g)

install clang then run:

```
go get github.com/andrewchambers/g
cd $GOPATH/src/andrewchambers/g
go test ./...
```

# brain storm

* gfmt tool.
* Package layouts like go.
* Multiple return values are just syntatic sugar over hidden pointer args. This allows C abi compatibility.
* Go style exports with case. But can be overridden with private or public keywords to allow c interop.
* Tagged unions supported explicitly
* No implicit casts like go.
* Less memory safety than go - can access arbitrary addresses.
* Directly output LLVM text assembly.
* support for inline assembly
* no := syntax. it does save alot of typing. var x = is probably less confusing to new people and less redundant.
* Bounds checking on arrays? optional or not?
* Macros as invoked subprograms? avoids needing special dsl, just a specified data format etc.
* Tuples + destructuring for multiple return?
