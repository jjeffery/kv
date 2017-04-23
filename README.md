# kv [![GoDoc](https://godoc.org/github.com/jjeffery/kv?status.svg)](https://godoc.org/github.com/jjeffery/kv) [![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/kv/master/LICENSE.md) [![Build Status](https://travis-ci.org/jjeffery/kv.svg?branch=master)](https://travis-ci.org/jjeffery/kv) [![Coverage Status](https://coveralls.io/repos/github/jjeffery/kv/badge.svg?branch=master)](https://coveralls.io/github/jjeffery/kv?branch=master)

Package kv makes it easy to work with lists of key/value pairs.

- [Structured logging APIs](#structured-logging-apis)
- [Key value pairs with standard library logging](#key-value-pairs-with-standard-library-logging)
- [Flattening](#flattening)
- [Fixing](#fixing)
- [Extending](#extending)
- [Performance](#performance)

## Structured logging APIs

Many structured logging APIs make use of a "keyvals" API, where key/value
pairs are passed as a variadic list of interface{} arguments.
For example, the [Go kit](https://github.com/go-kit/kit/tree/master/log) 
logger interface looks like this:

```go
type Logger interface {
    Log(keyvals ...interface{}) error
}
```

While there is flexibility and power in this variadic API,
one downside is the loss of any strict type checking. The following 
example (taken from Go kit) is a fairly typical example.
It is not obvious at a glance whether any arguments have been accidentally 
omitted.

```go
logger.Log("method", "GetAddress", "profileID", profileID, "addressID", addressID, "took", time.Since(begin), "err", err)
```
Package kv goes some way towards restoring type safety and improving clarity:
```go
// previous example can be written as
logger.Log(kv.Map{
    "method":    "GetAddress",
    "profileID": profileID,
    "addressID": addressID,
    "took":      time.Since(begin),
    "err":       err,
 })

// or alternatively
logger.Log(kv.P("method", "GetAddress"),
    kv.P("profileID", profileID),
    kv.P("addressID", addressID),
    kv.P("took", time.Since(begin)),
    kv.P("err", err))
```

The kv alternatives are more verbose, but in many situations the additional
clarity and type safety is worth the effort.

## Key value pairs with standard library logging

The `kv` package types `Pair`, `List` and `Map` all implement the `fmt.Stringer` interface and can render their values as text. For example if you like the simplicity of logging with key value pairs but are not ready to move to a structured logging package you can use the standard library logging. This approach works well with packages like [comail.io/go/colog](https://github.com/comail/colog):

```go
import "log"

func doSomething(s, t string) {
    // log a single key/value pair
    doOneThingWith(s)
    log.Println("did one thing", kv.P("s", s))
    
    // log multiple key/value pairs with a list
    u := getSometing(s)
    log.Println("got something", kv.List{
        "s", s,
        "u", u,
    })
     
    // log multiple key/value pairs with a map
    doAnotherThingWith(s, t, u)
    log.PrintLn("did another thing", kv.Map{
        "s": s,
        "t": t,
        "u", u,
    })
}
```

```
Output:
did one thing s="value for s"
got something s="value for s" u="value for u"
did another thing u="value for u" s="value for s" t="value for t"
```

## Flattening

The key to using the kv API is to use the `kv.Flatten` function to flatten
the keyvals before logging.

```go
// Flatten converts the contents of v into a slice
// of alternating key/value pairs.
func Flatten(v ...interface{}) []interface{}
```

A logging facade for the Go kit `Logger` interface looks like this:

```go
type logFacade struct {
    logger log.Logger
}

func (f *logFacade) Log(keyvals ...interface{}) error {
    return f.logger.Log(kv.Flatten(keyvals)...)
}
```

## Fixing

The `Flatten` function is reasonably good at working out what to do when
the input is not strictly conformant. It will infer a message value without 
a key and give it a "msg" key.

```go
// ["msg", "message 1", "key1", 1, "key", 2]
keyvals = kv.Flatten("message 1", kv.Map{
    "key1": 1,
    "key2": 2,
})
```

If a value is present without a key it will assign it an arbitrary one.
```go
// ["msg", "message 2", "key3", 3, "_p1", 4]
keyvals = kv.Flatten("msg", "message 3", "key3", 3, 4)
```

A single error gets turned into a message (but see *Extending* below).
```go
// ["msg", "the error message"]
keyvals = kv.Flatten(err)
```

See the [Flatten tests](https://github.com/jjeffery/kv/blob/master/flatten_test.go)
for more examples of how `kv.Flatten` will attempt to fix non-conforming 
keyvals lists.

## Extending

If an item implements the following interface, it will be treated as if it
is a list of key/value pairs.

```go
type keyvalser interface {
    Keyvals() []interface{}
}
```

For example, errors generated using the 
[github.com/jjeffery/errors](https://github.com/jjeffery/errors) package 
implement the `keyvalser` interface:

```go
if err := doSomethingWith(theThing); err != nil {
    return errors.Wrap(err, "cannot do something").With(kv.P("theThing", theThing))
}

// ...later on when we log the error ...

// msg="cannot do something" cause="file not found" theThing="the thing"
logger.Log(err)

```

*EXPERIMENTAL*: The following interfaces are also recognized, and are used in preference
to the keyvalser interface, as they indicate a more memory efficient
method for extracting one or more key/value pairs from an object.

```go
type keyvalPairer interface {
	KeyvalPair() (key string, value interface{})
}
```

```go
type keyvalMapper interface {
	KeyvalMap() map[string]interface{}
}
```

This makes it easy to define standard logging for types. For example:

```go
type User struct {
	ID string
	
	// ... other fields ...
}

func (u *User) KeyvalPair() (string, interface{}) {
	return "User.ID", u.ID
}

// ... later on ...

func doSomethingWithUser(u *User) {
	if !hasPermission(u) {
		// msg="permission denied" User.ID=u1234 
		logger.Log("permission denied", u)
	}
}
```

> The `keyvalPairer` and `keyvalMapper` interfaces seem like a good idea,
but have not been used all that much. If they do not prove all that useful
they *might* be removed in favor of simplicity.

## Performance

To date, the `kv` package has only been used in applications where peak 
logging rates are of the order of a few log messages per second. Because
performance has not been an issue, no serious time has been spent
analyzing and tuning for performance. The following discussion is based
on best guesses.

When constructing the variadic key/value lists, the `kv.P` function is the
fastest because it does not require any memory allocation. Using  `kv.Map` 
and `kv.Keyvals` has the overhead of allocating a map and array respectively.

The `kv.Flatten` function does not allocate any memory if the input keyvals
do not require any modification: it will return the input slice unchanged.
If high volume messages are already in valid "keyvals" format then using
package `kv` for lower volume error messages should not impact performance
significantly.

When the keyvals input has to be flattened and/or fixed, the `kv.Flatten` 
package has to allocate a new slice with a new backing array for the output.
Under some circumstances it also allocates temporary slices. It would be 
straightforward to create a new flattening function that delegates memory
allocation. Something like:

```go
type Pool interface {
	Get() []interface{} // fixed size; say length 64
	Put([]interface{})
}

func FlattenPool(input ...interface{}, pool Pool) []interface{}
```

The `keyvalser` interface may involve memory allocations depending on the
structure of the object that implements it. The `kv.List`, `kv.Map` and 
`kv.Pair` types all implement the `keyvalser` interface, but the
`kv.Flatten` function makes special exceptions for them and extracts 
their contents without allocating memory.

Type switches are another source of overhead. The `kv.Flatten` function
performs a type switch on every item in the input slice in order to decide 
if flattening or fixing is required. This type switching occurs even if the 
input does not require any modification. (TODO: run some benchmarks to
quantify this overhead).

## License

[MIT](https://raw.githubusercontent.com/jjeffery/kv/master/LICENSE.md)
