# G programming language

A new programming language - WIP

Imagine C with simpler syntax (based on Go) and a few cool things like packages and multiple returns.

G aims to...

* Be easy to parse and analyze like Go.
* Do anything C can do (unions, unsafe things).
* Be lean like C, should be suitable for microcontrollers.
* Have perfect interop with C.
* Have clear mapping to underlying machine code.
* Have inline assembly support.
* Use packages similar to Go and not headers.
* Have useful but minimalist extra features.

And more.

G is not trying to compete with big languages like rust or C++. It just trying to be as close to C as possible while mostly simplifying, and improving things.

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

* Multiple return values are just syntatic sugar over hidden pointer args. This allows C abi compatibility.
* Go style exports with case. But can be overridden with private or public keywords to allow c interop.
* Tagged unions supported explicitly
* No implicit casts like go.
* Less memory safety than go - can access arbitrary addresses.
* Directly output LLVM text assembly.
* support for inline assembly
* := syntax? it does save alot of typing. var x = is probably less confusing to new people.
* Bounds checking on arrays? optional or not?
* Macros as invoked subprograms? avoids needing special dsl, just a specified data format etc.
