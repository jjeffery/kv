/*
Package kv makes it easy to work with lists of key/value pairs.

Many structured logging APIs make use of a "keyvals" API, where key/value
pairs are passed as a variadic list of interface{} arguments.
For example, the Go kit logger interface is probably the simplest:

 type Logger interface {
     Log(keyvals ...interface{})
 }

Package kv provides a way to pass key/value pairs with clarity and
type-safety:

 // for example
 logger.Log("method", "GetAddress", "profileID", profileID, "addressID", addressID, "took", time.Since(begin), "err",

 // can be written as
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

Using kv results in more verbose code, but in many situations the additional
clarity and type safety is worth the effort.

Extensibility

If an item implements the following interface, it will be treated as if it
is a list of key/value pairs.

 type keyvalser interface {
     Keyvals() []interface{}
 }


*/
package kv
