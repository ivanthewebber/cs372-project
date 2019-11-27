# ![GOlang](https://golang.org/lib/godoc/images/go-logo-blue.svg)
### Introduction
[GOlang](https://golang.org/) (a.k.a. Go) is relatively new langauge (2009) created by google. This document gives a brief introduction to Go, focusing mostly on Go's characteristics and classification as a programming language.

Go was written to be more friendly to development than other languages. However, this doesn't mean it's a breeze. Go's tools require that developers follow a set of rules regulating file-trees, repositories, distribution, naming, and documentation. Of course, to a degreee one can opt out, but by keeping things consistent the language can work with itself much better.

Another goal of the language designers was to reduce verbosity and inconsistencies as experienced in Java and C/C++ respectively. In general the langauge feels naked due to the lack of parenthesis around for-loops (and etc.) and absence of classes. Like c this language is very close to the memory, but it's a very different feel.

## History and Current Status
Go was first released in November of 2009. Afterwards the Go team at Google continued to work on developing the project and building a community to provide feedback as the language developed. At [golang.org] and [blog.golang.org] one can find documentation, guides, and compilers for Go.

Go is [currently on version](https://golang.org/project/#go1) 1.13 (as of November 11, 2019). The standard library is still expanding and the development community is being careful to make smart choices for Go's future.

Another emerging hot technology is Web Assembly. Go's compiler can compile to Web Assembly, so this might be another adoption point.

## Paradigm
Go truly escapes the imperative paradigm with built-in concurrency via goroutines and channels, but Go doesn't really fit in with the object-oriented (OO) languages either. Unlike the common OO languages Go doesn't have a class, but it has two tools that serve the same purpose. First Go's functions have an optional acts-on parameter (i.e. a parameter which comes before a function connected with a dot). Additionally Go has interfaces which describe a group of methods that could act on a type. Any type that has functions of the same signature to act on it implicitly implements the interface.

Because Go requires a file-system naming scheme a class that implements an interface is internally given the same name, but outside is refered to by the enclosing package. For an example of this see [our rot13 package](rot13) which declares a Reader type that implements Reader by having a Read method. [In our rot13 program](main) we use this class for applying the rotation.

## Typing System
Unlike Java and C-family languages, Go doesn't require a type declaration every time a variable is declared. If the type can be implied from the return of a function, then it can be ommited. However, the `:=` operator must be used for declaration whereas the `=` is solely for assignment.

### Basic Types
Go's `int`, `uint`, `float`, and `bool` types are similar to those in the C-family, but Go also has specific-size variants like ECMAScript's TypedArrays (e.g. `int32`, `float64`). These are useful when precision is necessary or bitwise-wizardry is desired. Go also has built-in `complex` and `string` types simmilar to those in C++.

Go also has c-style pointers, a struct, and a type definition that is very simmilar to C's typedef.

### Composite Types
A composite type is defined based on relationships with basic and composite types. Like C, Go has `pointers`, `structs`, and `arrays`. Go also has `slice` types which are dynamic containers (like vectors), and `map` types which are (hashed) associative arrays. Go also has channel types for synchronization among goroutines and interfaces.

Go also has a built-in string type (it's very simmilar to the slice). For representing unicode characters Go has `rune`, an alias for `uint32`. Arrays, slices, strings, and maps can be iterated over (the string class takes care of the unicode stuff).

Finally, Go's functions are first-class objects meaning they can be parameters, return values, and functions can be on the righthand-side of assignments.

In the future Go will have custom-type generics.

 - <cite>[Type System Overview](https://go101.org/article/type-system-overview.html)</cite>

## Control Structures
Go has a single loop, the for loop, with all the clauses optional. Go has C-style `if`, and a switch which is only a shorthand for an if-else-if...-else block. Go also has a defer keyword that is simmilar to finally in Java. Go also has labeled break statements and goto.

```Go
if length := len(data); length == 0 {
	// do something...
} else if (length > 100) {
	// do something...
} else {
	// length is visible...
} // and length is invisible

for i := 0; i < 10; i += 1 {
	// Go's for loops are so naked...
	fmt.Printf("%d", i)
}

for i := 0; i < 10 {
	// Go's for loops can act like the while loops of other languages since all the clauses are optional...
	fmt.Printf("%d", i)
	i++
}

for index, value := range myArray {
	// do something...
}
```

## Semantics
Go is statically scoped; has static constants; has a garbage collector; uses static, stack-dynamic, and heap-dynamic allocation; and supports enclosures.

## Desirable Language Characteristics
### Efficiency -> Go Is FAST
Built-in concurrency, an awesome compiler, and other features make Go a fast (i.e. high-performance) language. Unlike other languages Go's goroutines allow developers to easily write platform-independent parallel code.

<cite>[Achieving Concurrency in Go](https://medium.com/rungo/achieving-concurrency-in-go-3f84cbf870ca)</cite>
<cite>[Five Things that Make Go Fast](https://dave.cheney.net/2014/06/07/five-things-that-make-go-fast)</cite>

### Extensibility
In Go you are able to declare new data types using the type keword which is comparable to typedef in C99, but Go currently does not allow you to extend a keyword or operator to work on user-defined types (unless, or course, the keyword or operation is defined for the underlying alias). For instance, Go has the keyword `range` for iterating over the elements in maps, arrays, slices, and strings. Currently iteration via `range` cannot be implemented on user-defined types.

### Regularity / Uniformity
The designers of Go did a lot of work to clean up the C-family syntax to ensure the language would be more uniform. Additionaly, Go's rules about package and repo organization guarante that the url will always match the file tree making it easy to prevent conflicts and trace the source of a package. Also, Go enforces minimal documentation standards and other things which at first are hard; but overall are for the good.

### Security/Reliability
Go handles exceptions in three steps: defer, panic, recover. First, the defer keyword defers the execution of a function until the end of a function. This is simmilar to a finally block, and is useful for keeping file opening and closing together.
```Go
func CopyFile(dstName, srcName string) (written int64, err error) {
    src, err := os.Open(srcName)
    if err != nil {
        return
    }
    defer src.Close() // now you don't have to worry about forgetting!

    dst, err := os.Create(dstName)
    if err != nil {
        return
    }
    defer dst.Close()

    return io.Copy(dst, src)
}
```
Second, panic is a built in function which acts similarly to an excetion in C++ or Java. Once invoked the current function's defered statements are executed normally and the function returns. This continues up the call tree until the program crashes or the built-in function recover is used within a defer block. Recover is able to catch the value given to panic and resume normal execution.

Go also has a built-in error interface which is often returned alongside other values in a method. If the value within the interface is not nil something has gone wrong. In order to compile code the error's must be checked and handled appropriately.

Go is also statically typed since the types of all variables can be inferred and does not restrict aliasing as you can have as many references to the same object as you want like in C.

<cite>[Defer, Panic, and Recover](https://blog.golang.org/defer-panic-and-recover)</cite>

### Slices
The array and slice types are as almost as easy-to-use as Python's list, but feel as close to the memory as C's pointers. The same data structure easly works as a stream, and can eithr be backed by a static array or dynamically resized. 

## Data Abstractions
For the most part Go's compiler decides for the programmer what variables will be stack-dynamic and which will be heap-dynamic.

## Syntax & Minimalism
A common fault in OO languages is verbosity. Consider this Java program:
```Java
package helloworld;

public class HelloWorld {

    public static void main(String[] args) {
        System.out.println("Hello World!");
    }

}
```
 - <cite>[Java's HelloWorld tutorial](https://docs.oracle.com/javase/tutorial/getStarted/cupojava/index.html)</cite>

And compare it to this:
```GO
package main

import "fmt"

func main() {
	fmt.Println("Hello, 世界")
}
```
 - <cite>[GO's Hello World widget](https://golang.org/)</cite>

In general Go attempts to be as minimal as possible to avoid unneeded characters, semicolons, and types (indeed, the type is only needed if it can't be discerned by other means). While minimalism sometimes has negative consequences (e.g. less readable, requireing knowledge of sub-standard acronyms and abbreviations), overall it removes a lot of fluff from the language. 

To truly see the verbose vs. minimalist comparison between Java and Go check out each language's hello-world tutorials.

### Type and Declaration
Most C-family languages use a syntax for type and declaration that is similar to a normal expression, but with types explicit. This is most clear for function pointers: the declaration `int (*fp)(int a, int b);` is meant to look like the expression `fp(a,b);` which evaluates to an int.

However, the designers of Go decided to use a seperate syntax for types and declarations in order to improve orthogonality and consistency. This is especially important when dealing with higher-order functions (with multiple returns). Consider the C function taking and returning a function pointer: 
```C
int (*(*fp)(int (*)(int, int), int))(int, int);
```
And the Go equivilent: 
```GO
f func(func(int,int) int, int) func(int, int) int
```
One of goals of the seperate syntax was to make a GO declaration left->right readable. 
The above can be read: 
`f is a func taking a func taking two ints and two ints and returning a function taking two ints and returning an int.` 
It's still a mouthful--but it's far less obsfucated if you ask me!

 - <cite>[Go's Declaration Syntax](https://blog.golang.org/gos-declaration-syntax)</cite>
 - <cite>[Go's Specification (with Extended Backus-Naur Form)](https://golang.org/ref/spec#Notation)</cite>
