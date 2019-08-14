package kv_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/jjeffery/kv"
)

func init() {
	// Use fmt.Println for logging messages so that Output works
	// in example test cases.
	kv.LogOutput = func(calldepth int, s string) error {
		_, err := fmt.Println(strings.TrimSpace(s))
		return err
	}
}

func ExampleContext() {
	// start with a context
	ctx := context.Background()

	// attach some key/value pairs
	ctx = kv.From(ctx).With("method", "get", "url", "/api/widgets")

	// ... later on log a message ...

	fmt.Println("permission denied", kv.From(ctx).With("user", "alice"))

	// Output:
	// permission denied user=alice method=get url="/api/widgets"
}

func ExampleError() {
	// setup the user
	user := "alice"

	// ... later on ...

	// create an error and log it
	err := kv.NewError("permission denied").With("user", user)
	fmt.Println(err)

	// alternatively, wrap an existing error
	err = kv.Wrap(err, "cannot open file").With("file", "/etc/passwd")
	fmt.Println(err)

	// Output:
	// permission denied user=alice
	// cannot open file: permission denied file="/etc/passwd" user=alice
}

func ExampleLog() {
	kv.Log("message 1")
	kv.Log("message 2", kv.With("a", 1))

	kv := kv.With("p", "q")

	kv.Log("message 3")
	kv.Log("message 4", kv.With("a", 1, "b", 2))

	ctx := kv.From(context.Background()).With("c1", 100, "c2", 200)

	kv.Log(ctx, "message 5")
	kv.Log("message 6:", ctx, "can be in any order")
	kv.Log(ctx, "message 7", kv.With("a", 1, "b", 2))

	// Output:
	// message 1
	// message 2 a=1
	// message 3 p=q
	// message 4 p=q a=1 b=2
	// message 5 p=q c1=100 c2=200
	// message 6: can be in any order p=q c1=100 c2=200
	// message 7 p=q a=1 b=2 c1=100 c2=200
}
