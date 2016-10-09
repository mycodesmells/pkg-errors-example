# Smart Golang Error Handling

This blog post is based on a great [GopherCon 2016 talk](https://www.youtube.com/watch?v=lsBF58Q-DnY) by Dave Cheney.

Handling errors in Golang seems to be a pretty easy thing to do. As a convention, whenever a function can result in some incorrect way, some error is returned as one of its values, which then can be passed further or displayed to the user. But how exactly should you do it? And is the standard package enough for all this? According to Dave Cheney, it's not.

### Standard package approach

Handling errors using the standard package, you are limited to two approaches when it comes to handling errors - returning an error as it is to the caller, or creating new error by adding something to its message:

    func ReturnAsItIs() error {
        err := SomethingReturningError()
        return err
    }

    func AddToTheMessage() error {
        err := SomethingReturningError()
        return fmt.Errorf("some additional context: %v", err)
    }

There are some problems with this approach, though. Returning an error as it results in a complete lack of error context. For example, if some function at the bottom of the execution stack returns some _file not found _ error, and you just pass it through, at the end you just get the same information - it's very hard to debug this way, right? Adding some text to the message at each level, as it is done in the second approach is slightly better, but you face another problem - except for having some text at the end of the message, it's not possible to determine an original error's type.

### Solution by Dave Cheney

Cheney created a package that can replace `errors` from the standard package because it extends its functionality and can be used under the same name. It's `github.com/pkg/errors` and allows you to handle errors in a much better way. It allows you to (spoiler alert) wrap errors with additional information, without losing its original context. It also gives us the possibility to see the whole stack trace, which may be useful for debugging more complicated parts of your software.

### Step-by-step comparison

First, we create three similar stack of functions, one calling another. The last one returns a custom error (of `common.MyError` type), which is handled differently by each of the three error handling approach. First, we just pass the original error:

    // bare/bare.go
    func CallA() error {
        return CallB()
    }

    func CallB() error {
        return CallC()
    }

    func CallC() error {
        return common.MyError{Msg: "Error from CallC"}
    }

Second approach is adding a bit of context with some text message and `fmt.Errorf(..)`:

    // concat/concat.go
    func CallA() error {
        return fmt.Errorf("Error from CallA: %v", CallB())
    }

    func CallB() error {
        return fmt.Errorf("Error from CallB: %v", CallC())
    }

    func CallC() error {
        return common.MyError{Msg: "Error from CallC"}
    }


Finally, in our third approach we take advantage of `pkg/errors` and enhance our error with additional text context, this time using `errors.Wrap(..)`:

    // wrap/wrap.go
    func CallA() error {
        return errors.Wrap(CallB(), "Error from CallA")
    }

    func CallB() error {
        return errors.Wrap(CallC(), "Error from CallB")
    }

    func CallC() error {
        return common.MyError{Msg: "Error from CallC"}
    }

In our `main.go` file we'll execute all three `CallA` functions and see what information do we get with the error at the top level:

    // main.go
    func main() {
        bareErr := bare.CallA()
        printErr("Bare", bareErr)

        concatErr := concat.CallA()
        printErr("Concat", concatErr)

        wrapErr := wrap.CallA()
        printErr("Wrap", wrapErr)
    }

    func printErr(name string, err error) {
        fmt.Printf("== %s ==\n", name)
        ...
        fmt.Println()
    }

**Error message**

Let's add printing the final error's message and see how much do we know about it:

    // main.go
    ...
    func printErr(name string, err error) {
        fmt.Printf("== %s ==\n", name)
        fmt.Printf("Message: %v\n", err)
        fmt.Println()
    }    

The output reveals the biggest problem with passing bare error through - we don't get any context and it's extremely hard to find where the error came from:

== Bare ==
Message: Errrrr: Error from CallC

== Concat ==
Message: Error from CallA: Error from CallB: Errrrr: Error from CallC

== Wrap ==
Message: Error from CallA: Error from CallB: Errrrr: Error from CallC

**Error type**

Once we get the error, we might want to check its type, right? So what do we get there?

    func printErr(name string, err error) {
        fmt.Printf("== %s ==\n", name)
        fmt.Printf("Type: %T\n", err)
        fmt.Println()
    }

This time, bare error is the winner. The solution from `concat` package returns some kind of `errorString`, while our favourite gives us `withStack` one. As you'll see in a second, it's not the preferred way to check for error type in `pkg/errors`, so don't worry too much:

    == Bare ==
    Type: common.MyError

    == Concat ==
    Type: *errors.errorString

    == Wrap ==
    Type: *errors.withStack

**Original cause of error**

Now we're entering previously unknown territory: checking for error's cause. This won't do any difference for the standard approaches, but gives us some magic in the other one:

    func printErr(name string, err error) {
        fmt.Printf("== %s ==\n", name)
        fmt.Printf("Original error? %v\n", errors.Cause(err))
        fmt.Println()
    }

This `errors.Cause(..)` method, imported from our 3rd party dependency looks for the first non-wrapped error and return it as the originator of the error stack. This shows us that we can get much more information now:

    == Bare ==
    Original error? Errrrr: Error from CallC

    == Concat ==
    Original error? Error from CallA: Error from CallB: Errrrr: Error from CallC

    == Wrap ==
    Original error? Errrrr: Error from CallC

Same goes for the type of an original error. We dive deep into the error to see where it all started:

    == Bare ==
    Original type? common.MyError

    == Concat ==
    Original type? *errors.errorString

    == Wrap ==
    Original type? common.MyError

**Stack trace**

Since we saw that the error returned by `pkg/errors` is something called `withStack`, we'd really like to see that stack, right? We surely can, but it requires us to create an interface that includes `StackTrace()` function. The package doesn't export `stackTracer` interface but it _can be regarded as a part of its API_:

    type stackTracer interface {
        StackTrace() errors.StackTrace
    }

Then, we can cast our error to this interface, loop through `StackTrace()` result and print its contents:

    func printStack(err error) {
        if err, ok := err.(stackTracer); ok {
            for i, f := range err.StackTrace() {
                fmt.Printf("%+s:%d", f, i)
            }
        } else {
            fmt.Println("No stack trace...")
        }
    }

    func printErr(name string, err error) {
        fmt.Printf("== %s ==\n", name)
        printStack(err)
        fmt.Println()
    }

It's not a surprise, that only our third approach returns anything:

    == Bare ==
    No stack trace...

    == Concat ==
    No stack trace...

    == Wrap ==
    github.com/mycodesmells/pkg-errors-example/wrap.CallA
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/wrap/wrap.go:0main.main
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/main.go:1runtime.main
        /usr/local/go/src/runtime/proc.go:2runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:3

We can also `fmt.Printf("%+v", err)` the error and see stack trace as well:

    Errrrr: Error from CallC
    Error from CallB
    github.com/mycodesmells/pkg-errors-example/wrap.CallB
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/wrap/wrap.go:13
    github.com/mycodesmells/pkg-errors-example/wrap.CallA
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/wrap/wrap.go:9
    main.main
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/main.go:19
    runtime.main
        /usr/local/go/src/runtime/proc.go:183
    runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:2086
    Error from CallA
    github.com/mycodesmells/pkg-errors-example/wrap.CallA
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/wrap/wrap.go:9
    main.main
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/main.go:19
    runtime.main
        /usr/local/go/src/runtime/proc.go:183
    runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:2086    

### Summary

Complete list of checks:

    func printErr(name string, err error) {
        fmt.Printf("== %s ==\n", name)
        fmt.Printf("Message: %v\n", err)
        fmt.Printf("Type: %T\n", err)
        fmt.Printf("Original error? %v\n", errors.Cause(err))
        fmt.Printf("Original type? %T\n", errors.Cause(err))
        printStack(err)

        fmt.Println()
    }

Its output:

    == Bare ==
    Message: Errrrr: Error from CallC
    Type: common.MyError
    Original error? Errrrr: Error from CallC
    Original type? common.MyError
    No stack trace...

    == Concat ==
    Message: Error from CallA: Error from CallB: Errrrr: Error from CallC
    Type: *errors.errorString
    Original error? Error from CallA: Error from CallB: Errrrr: Error from CallC
    Original type? *errors.errorString
    No stack trace...

    == Wrap ==
    Message: Error from CallA: Error from CallB: Errrrr: Error from CallC
    Type: *errors.withStack
    Original error? Errrrr: Error from CallC
    Original type? common.MyError
    github.com/mycodesmells/pkg-errors-example/wrap.CallA
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/wrap/wrap.go:0main.main
        /home/slomek/go/src/github.com/mycodesmells/pkg-errors-example/main.go:1runtime.main
        /usr/local/go/src/runtime/proc.go:2runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:3

Complete source code of this example is available [on Github](https://github.com/mycodesmells/pkg-errors-example).
