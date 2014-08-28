# G programming language

A new systems programming language - WIP

I am a big fan of both Go and C, however both have problems. Go aims at servers,
but there is a niche for something better than C, but still simple, portable, and minimal with overhead.
I think G can fill this gap.

C
* Hard to parse with alot of unnecessary syntax.
* Alot of undefined behaviour and funny things like implicit casts.
* Compiles are slow due to preprocessing

Go
* Requires GC and runtime.
* Less suitable for embedded applications due to runtime and binary requirements..

G aims to solves these problems.
* Easy to parse syntax, similar to Go.
* Complicated parts of C removed.
* No garbage collector and minimal runtime like C.
* abi compatible with C.

Examples:

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
