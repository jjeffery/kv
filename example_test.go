package kv_test

import (
	"fmt"
	log "fmt" // sleight of hand so example looks like its using the standard log package

	"github.com/jjeffery/kv"
)

var id = 42

type Record struct {
	sku, location, color, pickingCode string
}

func LookupSomething(id int) (*Record, bool) {
	return &Record{
		sku:         "X1FWP",
		location:    "bin 31",
		color:       "green",
		pickingCode: "p1p",
	}, true
}

func Example() {
	log.Println("trace: lookup up something", kv.P("id", id))
	record, ok := LookupSomething(id)
	if !ok {
		log.Println("error: not found", kv.P("id", id))
		return
	}

	// log a message with lots of key/value pairs
	log.Println("found article", kv.List{
		"sku", record.sku,
		"location", record.location,
		"color", record.color,
		"pickingCode", record.pickingCode,
	})

	// Log again, this time with a map. The main difference
	// is that with a map the keys are sorted. With a list the
	// keys are in the order given.
	log.Println("found article", kv.Map{
		"sku":         record.sku,
		"location":    record.location,
		"color":       record.color,
		"pickingCode": record.pickingCode,
	})

	// Output:
	// trace: lookup up something id=42
	// found article sku=X1FWP location="bin 31" color=green pickingCode=p1p
	// found article color=green location="bin 31" pickingCode=p1p sku=X1FWP
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

func ExampleP() {
	result, count := "the result", 1

	fmt.Println(
		kv.P("result", result),
		kv.P("count", count),
	)

	// Output:
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
	// log function prints to stdout.
	log := func(v ...interface{}) {
		list := kv.List(kv.Flatten(v))
		fmt.Println(list)
	}

	// Typical usage: structured logging
	log("msg", "message 1", "key1", 1, "key2", 2)

	// Message can be split into individual key/value pairs.
	log(kv.P("msg", "message 1"), kv.P("key1", 1), kv.P("key2", 2))

	// Assumes the first string without a keyword is the message ("msg").
	// Use a map for key/value pairs where the order is not important.
	log("message 2", kv.Map{
		"key1": "one",
		"key2": "two",
	})

	// A more complex, and probably unrealistic example of mixing styles.
	log("message3",
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

	// If a key is missing, one will be inserted to make the keyvals
	// slice valid.
	log("msg", "message 4", "key1", 1, 2)

	// Output:
	// msg="message 1" key1=1 key2=2
	// msg="message 1" key1=1 key2=2
	// msg="message 2" key1=one key2=two
	// msg=message3 key.1=1 key2=2 key3.1=3.1 key3.2=3.2 key4=4 key5=5
	// msg="message 4" key1=1 _p1=2
}
