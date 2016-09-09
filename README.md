# kv [![GoDoc](https://godoc.org/github.com/jjeffery/kv?status.svg)](https://godoc.org/github.com/jjeffery/kv) [![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/kv/master/LICENSE.md)

Package kv improves type safety when working with variadic key value pairs.

## Structured logging

Structured logging is a popular logging technique where log messages
are acknowleged as data, and should be machine parseable. Log entries
consist of a stricter key/value oriented message format.

A consequence is that logging APIs accept key/value pairs as an alternative
to the more traditional printf-style APIs.

```go
// unstructured
log.Printf("HTTP server listening on %s", addr)

// structured
logger.Log("msg", "server listening", "transport", "HTTP", "addr", addr)
```

## Keyvals

Many structured logging APIs make use of a "keyvals" API, where keys
and values are passed as alternating arguments in a variadic array of
interface{}. The simplest API comes from 
[Go kit](https://github.com/go-kit/kit/tree/master/log) 
(as does some of the text in the previous paragraphs):

```go
type Logger interface {
    Log(keyvals ...interface{})
}
```

## Package kv

While there is great flexibility and power in this variadic interface,
one downside is the loss of any strict type checking. Package kv goes
some way towards restoring type safety.

The following example (taken from Go kit) is a fairly typical example.
It is not obvious at first site if any arguments have been omitted and
the compiler does not provide any assistance.
```go
logger.Log("method", "GetAddress", "profileID", profileID, "addressID", addressID, "took", time.Since(begin), "err", err)
```

The same call is written more verbosely, but with more clarity and better
type safety:
```go
logger.Log(kv.Map{
    "method":    "GetAddress",
    "profileID": profileID,
    "addressID": addressID,
    "took":      time.Since(begin),
    "err":       err,
})
```
Many alternatives are possible, depending on preferences:

```go
logger.Log("method", "GetAddress",
    kv.P("profileID", profileID),
    kv.P("addressID", addressID"),
    kv.Keyvals{
        "took", time.Since(begin),
        err, // missing key
    })
```

The kv package is pretty good at picking up errors: in the last example
it will figure out that `err` is missing its keyword and will insert one.

This is only a very simple introduction to what is possible with this
package. There are many more examples in the 
[GoDoc](https://godoc.org/github.com/jjeffery/kv#example-Flatten) documentation.


# Logging facade

To use kv types with Go kit or other logging libraries, you need to 
write a logging facade. This is not a difficult task. As an example, 
the logging facade for the Go kit `Logger` interface looks like this:

```go
type logFacade struct {
	logger log.Logger
}

func (f *logFacade) Log(keyvals ...interface{}) {
	f.logger.Log(kv.Flatten(keyvals)...)
}
```
