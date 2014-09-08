# G programming language

A new simple systems programming language - WIP

I am a big fan of both Go and C, however both have problems.  You would never write an OS in Go, and C has alot of legacy quirks and design descisions that may not be good in hindsight.

G aims to solves these problems.

* G is easy to parse syntax which isn't context sensitive, similar to Go.
* G aims to remove unnecessarily complicated and legacy parts of C.
* G has no garbage collector, it's runtime is exactly like C.
* G can call C code with the correct signatures and shares linkers and toolchains. G is ABI compatible with C.
* G has Inline assembly.
* G has zero initialized variables.
* G uses packages, not headers.
* G allows multiple return values from functions for error reporting.
* G assignment is not an expression
* G has no ternary operator.

G is not trying to compete with big languages like rust or c++. It just trying to be as close to C as possible while mostly removing and simplifying things.

Example:

```
package main

import "stdio"

func main() int {
    stdio.printf("hello world!\n")
    return 0
}
```


# brain storm

* Multiple return values are just syntatic sugar over hidden pointer args. This allows abi compatibility.
* Go style exports with case. But can be overridden with private or public keywords to allow c interop.
* Tagged unions supported explicitly
* No implicit casts like go.
* Less memory safety than go - can access arbitrary addresses.
* Directly output LLVM text assembly.
* support for inline assembly
* := syntax? it does save alot of typing.
* Bounds checking on arrays? optional or not?
* Macros as invoked subprograms? avoids needing special dsl, just a specified data format etc.
