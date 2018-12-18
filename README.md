# kv [![GoDoc](https://godoc.org/github.com/jjeffery/kv?status.svg)](https://godoc.org/github.com/jjeffery/kv) [![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/jjeffery/kv/master/LICENSE.md) [![Build Status](https://travis-ci.org/jjeffery/kv.svg?branch=master)](https://travis-ci.org/jjeffery/kv) [![Coverage Status](https://coveralls.io/repos/github/jjeffery/kv/badge.svg?branch=master)](https://coveralls.io/github/jjeffery/kv?branch=master)

Package kv provides support for working with collections of key/value pairs.

### Lists, maps, pairs

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

### Messages, errors, context

A message is some optional free text followed by zero, one or more key value pairs. Examples:
```
message text with key/value pairs key1=value1 key1=value2
message text with no key/value pairs
key=value1 key=value2 key3=value3
```

Key/values can be stored in the context.
```
ctx := context.Background()

// associate some key/value pairs with the context
ctx = kv.NewContext(ctx).With("url", "/api/widgets", "method", "get")

// create a message
msg1 := kv.Msg("first message").With("key1", "value 1")

// create another message with values from the context
msg1 := kv.Msg("second message).With("key2", "value 2").Ctx(ctx)

fmt.Println(msg1)
fmt.Println(msg2)

// Output:
// first message key1="value 1"
// second message key2="value 2" url="/api/widgets" method=get
```

See the [GoDoc](https://godoc.org/github.com/jjeffery/kv) for more details.
