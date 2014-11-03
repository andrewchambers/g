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


# Status

Currently undergoing large refactor so broken, look back later.

# Build and test:
[![Build Status](https://travis-ci.org/andrewchambers/g.svg?branch=master)](https://travis-ci.org/andrewchambers/g)

install clang then run:

```
go get github.com/andrewchambers/g
cd $GOPATH/src/andrewchambers/g
go test ./...
```

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

Switch Improvements:

```
switch {
    case strcmp(s,"foo") == 0:
    case strcmp(s,"bar") == 0:
    default:
}

// Errors on duplicated cases.
switch v {
    case 0,1,2:
    case 4..10:
    default:
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

# Tentative examples

Typed new keyword?

```
 var x = new int
 var x = new [20]int
```

Refcount type or types (atomic ref for thread safety)?:

```
// No idea on syntax
func newRef () ref Foo {
    var r = newref Foo
    return r
}
//  what about ref arith?
XXX
```


Tagged union perhaps? might be too much, when people can just do it themselves.
The problem this solves is explicitly catching all cases if requirements change.
```
 
 type ASTNode tunion {
     
     x struct {
        foo int
     }
     
     y struct {
        bar int
     }
     
     z struct {
        baz int
     }
 }
 
 
 var v ASTNode = x{}
 
 // Catches unhandled cases with compiler error.
 match v {
    case x:
        v.foo
    case y:
        v.bar
    case z:
        v.baz
 }

```

Switch case on enums? Solves the problem above but is more general.:

```
type MyEnumType enum {
            X
            Y
            Z
}

var v = MyEnumType{X}

// Error catches unhandled enum cases.
switch v {
    case X,Y:
    case Z:
}

```

Deferred cleanup:
```
func doSomething () (int,*char) {
    var p *Foo = malloc(sizeof(Foo))
    if p == nil {
        return -1,"malloc failed"
    }
    //After a defer block is executed, the contained block will run on containing scope exit.
    defer {
        free((*void)p)
    }
    
    if cond() {
        return 0,""
    }
    
    return 1,""
}
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

