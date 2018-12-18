# kv [![GoDoc](https://godoc.org/github.com/jjeffery/kv?status.svg)](https://godoc.org/github.com/jjeffery/kv) [![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/kv/master/LICENSE.md) [![Build Status](https://travis-ci.org/jjeffery/kv.svg?branch=master)](https://travis-ci.org/jjeffery/kv) [![Coverage Status](https://coveralls.io/repos/github/jjeffery/kv/badge.svg?branch=master)](https://coveralls.io/github/jjeffery/kv?branch=master)

Package kv provides support for working with collections of key/value pairs.

### Lists, maps, pairs

The types `Message`, `Pair`, `List` and Map all implement the `fmt.Stringer` interface
and the `encoding.TextMarshaler` interface, and so they can render themselves as text.
```go
// key/value list
l := kv.List{
    "key1", "value 1",
    "key2", 2,
})

// key/value map
m := kv.Map{
    "key3": "value 3",
    "key4": 4,
})

// key/value pair
p := kv.Pair{key: "key5", value: 5}) // alternatively kv.P("key5", 5)

fmt.Println(l, m, p)

// Output:
// key1="value 1" key2=2 key3="value 3" key4=4 key5=5
```

If you like the simplicity of logging with key value pairs but are not ready to
move away from the standard library `log` package you can use this package to 
render your key value pairs.
```go
log.Println("this is a log message", kv.List{
    "key1", "value 1",
    "key2", 2,
})

// Output:
// this is a log message key1="value 1" key2=2
```

### Messages, errors, context

A message is some optional free text followed by zero or more key/value pairs:
```
example message with key/value pairs key1=1 key2="second value"
message text with no key/value pairs
message=without any="free-text"
```

Messages are easily constructed with message text and/or key/value pairs.
```go
// create a message
msg1 := kv.Msg("first message").With("key1", "value 1")
fmt.Println(msg1)

// Output:
// first message key1="value 1"
```

Key/value pairs can be stored in the context:
```go
ctx := context.Background()

// associate some key/value pairs with the context
ctx = kv.NewContext(ctx).With("url", "/api/widgets", "method", "get")

// create another message with values from the context
msg2 := kv.Msg("second message").With("key2", "value 2").Ctx(ctx)

fmt.Println(msg2)

// Output:
// second message key2="value 2" url="/api/widgets" method=get
```

Errors are easily constructed with key/value pairs:
```go
// Create a new error
err := kv.Err("composite literal uses unkeyed fields").With("file", filename, "line", lineno)
fmt.Println(err)

// Output:
// an error has occurred file="example_test.go" line=92

// Wrap an existing error
err = kv.Wrap(err, "vet").With("severity", severity)
fmt.Println(err)

// Output:
// vet: composite literal uses unkeyed fields severity=warning file="example_test.go" line=92
```

See the [GoDoc](https://godoc.org/github.com/jjeffery/kv) for more details.
