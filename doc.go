/*
Package kv makes it easy to work with lists of key/value pairs.

Structured logging

Many structured logging APIs make use of a "keyvals" API, where key/value
pairs are passed as a variadic list of interface{} arguments.
For example, the Go kit (http://gokit.io/) logger interface looks like this:

 type Logger interface {
     Log(keyvals ...interface{})
 }

While there is flexibility and power in this variadic API,
one downside is the loss of any strict type checking. The following
example (taken from Go kit) is a fairly typical example.
It is not obvious at a glance whether any arguments have been accidentally
omitted.

 logger.Log("method", "GetAddress", "profileID", profileID, "addressID", addressID, "took", time.Since(begin), "err", err)

Package kv goes some way towards restoring type safety and improving clarity:

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

The kv alternatives are more verbose, but in many situations the additional
clarity and type safety is worth the effort.

Flattening

The key to using the kv API is to use the kv.Flatten function to flatten
the keyvals before logging.

 // Flatten converts the contents of v into a slice
 // of alternating key/value pairs.
 func Flatten(v ...interface{}) []interface{}

A logging facade for the Go kit log.Logger interface looks like this:

 type logFacade struct {
     logger log.Logger
 }

 func (f *logFacade) Log(keyvals ...interface{}) {
     f.logger.Log(kv.Flatten(keyvals)...)
 }

Fixing

The kv.Flatten function is reasonably good at working out what to do when
the input is not strictly conformant. It will infer a message value without
a key and give it a "msg" key.


 // ["msg", "message 1", "key1", 1, "key", 2]
 keyvals = kv.Flatten("message 1", kv.Map{
     "key1": 1,
     "key2": 2,
 })

If a value is present without a key it will assign it an arbitrary one.

 // ["msg", "message 2", "key3", 3, "_p1", 4]
 keyvals = kv.Flatten("msg", "message 3", "key3", 3, 4)

A single error gets turned into a message (but see "Extending" below).

 // ["msg", "the error message"]
 keyvals = kv.Flatten(err)

See the Flatten tests for more examples of how kv.Flatten will attempt to
fix non-conforming keyvals lists.

Extending

If an item implements the following interface, it will be treated as if it
is a list of key/value pairs.

 type keyvalser interface {
     Keyvals() []interface{}
 }

For example, errors generated using the "github.com/jjeffery/errorv" package
implement the keyvalser interface:

 if err := doSomethingWith(theThing); err != nil {
     return errorv.Wrap(err, "cannot do something", kv.P("theThing", theThing))
 }

 // ...later on when we log the error ...

 // msg="cannot do something" cause="file not found" theThing="the thing"
 logger.Log(err)
*/
package kv
