package kv_test

import (
	"github.com/jjeffery/kv"
)

type Logger interface {
	Log(v ...interface{})
}

func Example(logger Logger) {
	logger.Log(
		kv.P("key1", "value1"),
		kv.Map{
			"key2": 2,
			"key3": 3,
		},
		kv.Keyvals{"key4", 4, "key5", 5.0})

	// how to disambiguate
	logger.Log("message on its own", "user", "A12345678")
}
