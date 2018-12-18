/*
Package kv provides support for working with collections of key/value pairs.
The package provides types for a pair, list and map of key/value pairs. It
also provides support for easily creating messages and errors with associated
key/value pairs.

Lists, maps pairs

The types Message, Pair, List and Map all implement the fmt.Stringer interface
and the encoding.TextMarshaler interface, and so they can render themselves as
text. For example if you like the simplicity of logging with key value pairs but
are not ready to move away from the standard library log package you can
use this package to render your key value pairs.
  log.Println("this is a log message", kv.List{
      "key1", "value 1",
      "key2", 2,
  })

  // Output: (not including prefixes added by the log package):
  // this is a log message key1="value 1" key2=2

Messages, errors, context

A message is some optional free text followed by zero or more key/value pairs:
 example message with key/value pairs key1=1 key2="second value"
 message text with no key/value pairs
 message=without any="free-text"

Messages are easily constructed with message text and/or key/value pairs.
 // create a message
 msg1 := kv.Msg("first message").With("key1", "value 1")
 fmt.Println(msg1)

 // Output:
 // first message key1="value 1"

Key/value pairs can be stored in the context:
 ctx := context.Background()

 // associate some key/value pairs with the context
 ctx = kv.NewContext(ctx).With("url", "/api/widgets", "method", "get")

 // create another message with values from the context
 msg2 := kv.Msg("second message").With("key2", "value 2").Ctx(ctx)

 fmt.Println(msg2)

 // Output:
 // second message key2="value 2" url="/api/widgets" method=get

Errors can be constructed easily with key/value pairs:

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

Flattening and fixing

The kv.Flatten function accepts a slice of interface{} and "flattens" it to
return a slice with an even-numbered length where the value at every even-numbered
index is a keyword string. It can flatten arrays:

 keyvals := kv.Flatten({"k1", 1, []interface{}{"k2", 2, "k3", 3}, "k4", 4})
 // ["k1", 1, "k2", 2, "k3", 3, "k4", 4]

Flatten is reasonably good at working out what to do when
the input length is not an even number, or when one of the items at an even-numbered
index is not a string value.

 keyvals := kv.Flatten("message 1", kv.Map{
     "key1": 1,
     "key2": 2,
 }))
 // ["msg", "message 1", "key1", 1, "key2", 2]

See the Flatten tests for more examples of how kv.Flatten will attempt to
fix non-conforming key/value lists. (https://github.com/jjeffery/kv/blob/master/flatten_test.go)

The keyvalser interface

The List, Map and Pair types all implement the following interface:

 type keyvalser interface {
     Keyvals() []interface{}
 }

The Flatten function recognises types that implement this interface and treats
them as a slice of key/value pairs when flattening a slice.
*/
package kv
