package kv_test

import (
	"fmt"
)

import (
	"github.com/go-logfmt/logfmt"
	"github.com/jjeffery/kv"
)

type Logger interface {
	Log(v ...interface{})
}

func doLog(v ...interface{}) {
	data, _ := logfmt.MarshalKeyvals(kv.Flatten(v)...)
	fmt.Println(string(data))
}

var errNotFound error
var logger = struct {
	Log func(v ...interface{})
}{
	Log: doLog,
}

var id = 42

type recordT struct {
	sku, location, color, pickingCode string
}

func LookupSomething(id int) (*recordT, bool) {
	return nil, false
}

func Example() {
	// logger.Log(v ...interface{})

	record, ok := LookupSomething(id)
	if !ok {
		logger.Log("not found", kv.P("id", id))
		return
	}

	logger.Log("found article", kv.Map{
		"sku":         record.sku,
		"location":    record.location,
		"color":       record.color,
		"pickingCode": record.pickingCode,
	})

	// Output:
	// msg="not found" id=42
}

func ExampleKeyvals() {
	var result string
	var count int

	result, count = getResultAndCount()

	logger.Log(kv.List{
		"result", result,
		"count", count,
	})
}

func getResultAndCount() (string, int) {
	return "", 0
}

func ExamplePair() {
	var result string
	var count int

	result, count = getResultAndCount()

	logger.Log(
		kv.Pair{"result", result},
		kv.Pair{"count", count})

	// this alternative requires slightly less typing
	logger.Log(
		kv.P("result", result),
		kv.P("count", count))
}

func ExampleP() {
	var result string
	var count int

	result, count = getResultAndCount()

	logger.Log(
		kv.P("result", result),
		kv.P("count", count))
}

func ExampleMap() {
	var result string
	var count int

	result, count = getResultAndCount()

	logger.Log(kv.Map{
		"result": result,
		"count":  count,
	})

	// The output will be either
	//   result="result here" count=1
	// or
	//   count=1 result="result here"
}

func ExampleFlatten() {
	// Example of a log facade that prints to stdout in
	// logfmt format (uses package github.com/go-logfmt/logfmt).
	log := func(v ...interface{}) {
		keyvals := kv.Flatten(v)
		data, _ := logfmt.MarshalKeyvals(keyvals...)
		fmt.Println(string(data))
	}

	// Message can be split into individual key/value pairs.
	log(kv.P("msg", "message 1"), kv.P("key1", 1), kv.P("key2", 2))

	// Assumes the first string without a keyword is the message ("msg").
	// Use a map for key/value pairs where the order is not important.
	// Only including one entry in the map so that the output is predictable
	// for testing.
	log("message 2", kv.Map{
		"key1": "one",
	})

	// A more complex, and probably unrealistic example of mixing styles.
	log("message3",
		kv.List{
			"key1", 1,
			"key2", 2,
		},
		kv.Map{
			"key3": 3,
		},
		"key4", 4,
		kv.P("key5", 5),
	)

	// If a key is missing, one will be inserted to make the keyvals
	// slice valid.
	log("msg", "message 4", "key1", 1, 2)

	// Output:
	// msg="message 1" key1=1 key2=2
	// msg="message 2" key1=one
	// msg=message3 key1=1 key2=2 key3=3 key4=4 key5=5
	// msg="message 4" key1=1 _p1=2
}
