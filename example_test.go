package kv_test

import (
	"context"
	log "fmt"

	"github.com/jjeffery/kv"
)

func ExampleContext() {
	// start with a context
	ctx := context.Background()

	// attach some key/value pairs
	ctx = kv.From(ctx).With("method", "get", "url", "/api/widgets")

	// ... later on log a message ...

	log.Println("permission denied", kv.From(ctx).With("user", "alice"))

	// Output:
	// permission denied user=alice method=get url="/api/widgets"
}

func ExampleError() {
	// setup the user
	user := "alice"

	// ... later on ...

	// create an error and log it
	err := kv.NewError("permission denied").With("user", user)
	log.Println(err)

	// alternatively, wrap an existing error
	err = kv.Wrap(err, "cannot open file").With("file", "/etc/passwd")
	log.Println(err)

	// Output:
	// permission denied user=alice
	// cannot open file: permission denied file="/etc/passwd" user=alice
}
