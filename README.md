# sectiontrace

sectiontrace is a Golang library to generate performance
profiling and latency-debuggging data by measuring time
spent in sections of code. It generates data in the
`trace_event` format, compatible with `about:tracing`
in Chromium.

## How to use sectiontrace

### Importing the library

```
import (
  "github.com/steinarvk/sectiontrace"
)
```

### Declaring sections

Sections should have names that are globally unique
throughout your program, in order to generate unambiguous
data, but the library will not enforce this unless you
enable `DebugMode`. Because they are supposed to be
globally unique, sections should usually be declared on
the global scope.

```
  var sectionFunctionA = sectiontrace.New("FunctionA")
  var sectionFunctionB = sectiontrace.New("FunctionB")
  var sectionFunctionC_CallA = sectiontrace.New("FunctionC.CallA")
  var sectionFunctionC_CallB = sectiontrace.New("FunctionC.CallB")
```

### Annotating sections of code

There are several ways to annotate sections of code for
profiling:

```
  // FunctionA is annotated with a callback function.
  func FunctionA(ctx context.Context, ...) error {
    return sectionFunctionA.Do(ctx, func(ctx context.Context) error {
      ...
    })
  }

  // FunctionB is annotated with a Begin() / End() pair of calls.
  func FunctionB(ctx context.Context, ...) (err error) {
    ctx, sec := sectionFunctionB.Begin()
    defer func() { sec.End(err) }()

    ...
    err = ...
  }

  // FunctionC moves through multiple phases.
  func FunctionC(ctx context.Context, ...) error {
    ctx, sec := sectionFunctionC_Setup.Begin(ctx)
    defer func() { sec.End(nil) }()

    ...

    ctx, sec = sec.NextPhase(sectionFunctionC_CallA)

    FunctionA(ctx)

    ctx, sec = sec.NextPhase(sectionFunctionC_CallB)

    FunctionB(ctx)

    return nil
  }
```

### Collecting and exporting the data

Lastly, you must declare what you want to do with the data.
The library simply calls its hooks `OnBegin` and `OnEnd`;
the wider program must override these in order to collect
the data and make it useful. Here's a simple example that
collects all the records and prints them to stdout in
a format (the JSON object format) that can directly be
loaded by `about:tracing`:

```
  func main() {
	  var records []*sectiontrace.Record
    defer func() {
      json.NewEncoder(os.Stdout).Encode(sectiontrace.Export(records))
    }()

	  sectiontrace.OnBegin = func(begin *sectiontrace.Record) {
		  records = append(records, begin)
	  }
	  sectiontrace.OnEnd = func(begin, end *sectiontrace.Record) {
		  records = append(records, end)
	  }

    ...
  }
```

### Extra data

sectiontrace's trace data includes extra information to be
able to fully reconstruct the section tree. The `a`
(ancestor) and `p` (parent) fields in `args` reference
the `id` fields of other sections in the same scope.

### Pitfalls

#### Use contexts

For maximum usefulness,  code to be profiled should be
using contexts. sectiontrace should still work without them,
but it will lack the extra data connecting different
sections.

Contexts should be propagated correctly; i.e. always use the
context returned from sectiontrace within a section and
never after it.

### Be careful when deferring the End call

When using the `Begin` and `End` style, you may have noted
that the example code above does the defer like this:

```
  ctx, sec := sectionFunctionC_Setup.Begin()
  defer func() { sec.End(err) }()

  ctx, sec = sec.NextPhase(sectionFunctionC_CallA)
```

This may look redundant, and it could be tempting to write
it like this:

```
  ctx, sec := sectionFunctionC_Setup.Begin()
  defer sec.End(err)  // WRONG

  ctx, sec = sec.NextPhase(sectionFunctionC_CallA)
```

Unfortunately, this will cause `sec` (and `err`!) to be
resolved immediately. Although the variables are rebound
later, the call will just close the original section again
(and `err` will probably not be set).

Also look out for shadowing variables instead of rebinding
them:
```
  ctx, sec := sectionFunctionC_Setup.Begin()
  defer func() { sec.End(err) }()

  ctx, sec := sec.NextPhase(sectionFunctionC_CallA)  // WRONG
```

### Handling usage errors and panics

sectiontrace will panic by default on failed sanity checks
and usage errors.

Panicking on failed sanity checks (e.g. if you alter _its_
context keys to contain invalid values).

Panicking on less-serious usage errors (like closing a
section twice) may or may not be the behaviour you want.

To make the library do something else instead, override
`OnUsageError` and/or `OnPanic`. For instance, to log
the errors and ignore them:

```
  sectiontrace.OnUsageError = func(err error) {
    log.Printf("sectiontrace usage error: %v", err)
  }
  sectiontrace.OnPanic = func(err error) {
    log.Printf("sectiontrace panic: %v", err)
  }
```

## Authorship

sectiontrace was written by me, Steinar V. Kaldager.

This is a hobby project written primarily for my own usage.
Don't expect support for it. It was developed in my spare
time and is not affiliated with any employer of mine.

It is released as open source under the MIT license. Feel
free to make use of it however you wish under the terms
of that license.

## License

This project is licensed under the MIT License - see the
LICENSE file for details.
