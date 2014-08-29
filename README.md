# G programming language

A new simple systems programming language - WIP

I am a big fan of both Go and C, however both have problems.  You would never  write an OS in Go, and C has alot of legacy quirks and design descisions that may not be good in hindsight.

G aims to solves these problems.
* Easy to parse syntax which isn't context sensitive, similar to Go.
* Unnecessarily complicated and legacy parts of C removed.
* No garbage collector and minimal runtime like C.
* ABI compatible with C.
* Inline assembly.
* Zero initialized variables.
* Packages, not headers.

G is not trying to compete with big languages like rust or c++. It just trying to be as close to C as possible while removing bad features.

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
* := syntax needs to be implemented later.
