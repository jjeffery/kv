package kv_test

import (
	"context"
	"errors"
	"fmt"
	log "fmt" // sleight of hand so example looks like its using the standard log package

	"github.com/jjeffery/kv"
)

func init() {
	kv.LogFunc = func(args ...interface{}) { fmt.Println(args...) }
}

func ExampleList() {
	result, count := "the result", 1

	fmt.Println("results are in:", kv.List{
		"result", result,
		"count", count,
	})

	// Output:
	// results are in: result="the result" count=1
}

func getResultAndCount() (string, int) {
	return "result here", 1
}

func ExampleWith() {
	result, count := "the result", 1

	fmt.Println("results are in:", kv.With(
		"result", result,
		"count", count,
	))

	// Output:
	// results are in: result="the result" count=1
}

func ExamplePair() {
	result, count := "the result", 1

	// go vet warns: "composite literal uses unkeyed fields"
	fmt.Println(
		kv.Pair{"result", result},
		kv.Pair{"count", count},
	)

	// this alternative requires slightly less typing
	// and avoids the go vet warning
	fmt.Println(
		kv.P("result", result),
		kv.P("count", count),
	)

	// Output:
	// result="the result" count=1
	// result="the result" count=1
}

func ExampleMap() {
	result, count := "the result", 1

	fmt.Println(kv.Map{
		"result": result,
		"count":  count,
	})

	// Output:
	// count=1 result="the result"
}

func ExampleFlatten() {
	// printKV flattens, fixes and prints args as key/value pairs
	printKV := func(args ...interface{}) {
		list := kv.List(kv.Flatten(args))
		fmt.Println(list)
	}

	// flatten: multiple lists, maps, pairs are flattened
	printKV("message1",
		kv.List{
			"key.1", 1,
			"key2", 2,
		},
		kv.Map{
			"key3.1": 3.1,
			"key3.2": 3.2,
		},
		"key4", 4,
		kv.P("key5", 5),
	)

	// fix: if a key is missing, one will be inserted to make the keyvals slice valid.
	printKV("msg", "message 2", "key1", 1, 2)

	// Output:
	// msg=message1 key.1=1 key2=2 key3.1=3.1 key3.2=3.2 key4=4 key5=5
	// msg="message 2" key1=1 _p1=2
}

func ExampleContext() {
	ctx := context.Background()

	// create a new context with keyvals attached.
	ctx = kv.NewContext(ctx).With("method", "GET", "url", "/api/widgets/1")

	// ... and then later on ...

	// add more context
	ctx = kv.NewContext(ctx).With("userid", "alice")

	// ... and then later on ...

	// use the values for logging later on
	log.Println("something happened", kv.From(ctx).With(
		"code", "red",
	))

	// Output:
	// something happened code=red userid=alice method=GET url="/api/widgets/1"
}

func ExampleMsg() {
	log.Println(kv.Msg("logging a message").With(
		"key1", 1,
		"key2", "two",
	))

	// Output:
	// logging a message key1=1 key2=two
}

func ExampleMessage_From() {
	ctx := context.Background()

	// create a new context with keyvals attached.
	ctx = kv.NewContext(ctx).With("method", "GET", "url", "/api/widgets/1")

	// use the values for logging later on
	log.Println(kv.Msg("something happened").From(ctx))

	// Output:
	// something happened method=GET url="/api/widgets/1"
}

func ExampleMessage_Msg() {
	ctx := context.Background()

	// create a new context with keyvals attached.
	ctx = kv.NewContext(ctx).With("method", "GET", "url", "/api/widgets/1")

	// use the values for logging later on
	log.Println(kv.From(ctx).Msg("something happened"))

	// Output:
	// something happened method=GET url="/api/widgets/1"
}

func ExampleMessage_Wrap() {
	// create a context with key/value pairs
	ctx := kv.NewContext(context.Background()).With("user", "scott")

	// ... later on there is an error ...
	err := errors.New("permission denied")

	// wrap the error with context key/value pairs
	err = kv.From(ctx).Wrap(err)

	log.Println("error:", err)

	// Output:
	// error: permission denied user=scott
}

func ExampleMessage_Err() {
	// create a context with key/value pairs
	ctx := kv.NewContext(context.Background()).With("user", "tiger")

	// ... later on there is an error ...
	err := kv.From(ctx).Err("permission denied")

	log.Println("error:", err)

	// Output:
	// error: permission denied user=tiger
}

func ExampleMessage_Log() {
	// An alternative way to log messages
	kv.Msg("something happened").With("key1", 1, "key2", "value 2").Log()

	// Output:
	// something happened key1=1 key2="value 2"
}

func ExampleError() {
	file := "testing-file"
	err1 := errors.New("elf header corrupted")
	err2 := kv.Wrap(err1, "emit macho dwarf").With("file", file)
	fmt.Println(err2)

	// Output:
	// emit macho dwarf: elf header corrupted file="testing-file"
}
