package kv

// KeyvalsAppender appends key/value pairs to keyvals and returns
// the result.
type KeyvalsAppender interface {
	AppendKeyvals(keyvals []interface{}) []interface{}
}

type Keyvals []interface{}

func (k Keyvals) AppendKeyvals(keyvals []interface{}) []interface{} {
	// convert to an array of interface{}
	vals := []interface{}(k)

	const typeKVA = 1
	const typeString = 2
	const typeError = 3
	const typeOther = 4

	types := make([]byte, len(vals))

	for i, val := range vals {
		switch val.(type) {
		case KeyvalsAppender:
			types[i] = typeKVA
		case string:
			types[i] = typeString
		case error:
			types[i] = typeError
		default:
			types[i] = typeOther
		}
	}

	for len(vals) > 0 {
		switch types[0] {
		case typeKVA:
			keyvals = vals[0].(KeyvalsAppender).AppendKeyvals(keyvals)
			vals = vals[1:]
			types = types[1:]
		case typeString:
			if len(vals) == 1 || types[1] == typeKVA {
				// we have a string on its own, which usually means
				// there is a value without a key
				keyvals = append(keyvals, "MISSING_KEY", vals[0])
				vals = vals[1:]
				types = types[1:]
			} else {
				keyvals = append(keyvals, vals[0], vals[1])
				vals = vals[2:]
				types = types[2:]
			}
		case typeError:
			keyvals = append(keyvals, "error", vals[0])
			vals = vals[1:]
			types = types[1:]
		default:
			keyvals = append(keyvals, "MISSING_KEY", vals[0])
			vals = vals[1:]
			types = types[1:]
		}
	}

	return keyvals
}

type Pair struct {
	Key   string
	Value interface{}
}

func P(key string, value interface{}) Pair {
	return Pair{
		Key:   key,
		Value: value,
	}
}

func (p Pair) AppendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals, p.Key, p.Value)
}

type Map map[string]interface{}

func (m Map) AppendKeyvals(keyvals []interface{}) []interface{} {
	for key, value := range m {
		keyvals = append(keyvals, key, value)
	}
	return keyvals
}

type Error struct {
	Err error
}

func (e Error) AppendKeyvals(keyvals []interface{}) []interface{} {
	if e.Err == nil {
		return keyvals
	}
	return append(keyvals, "error", e.Err)
}

func Err(err error) Error {
	return Error{
		Err: err,
	}
}

func Flatten(keyvals []interface{}) []interface{} {
	return Keyvals(keyvals).AppendKeyvals(nil)
}

func Flatten__(keyvals []interface{}) []interface{} {
	var appenders []KeyvalsAppender

	startIndex := 0
	for i := 0; i < len(keyvals); i++ {
		v := keyvals[i]
		if appender, ok := v.(KeyvalsAppender); ok {
			if i > startIndex {
				appenders = append(appenders, Keyvals(keyvals[startIndex:i]))
			}
			appenders = append(appenders, appender)
			startIndex = i + 1
		}
	}

	if startIndex == 0 {
		// no flattening required
		return keyvals
	}

	if startIndex < len(keyvals) {
		appenders = append(appenders, Keyvals(keyvals[startIndex:]))
	}

	var keyvals2 []interface{}
	for _, appender := range appenders {
		keyvals2 = appender.AppendKeyvals(keyvals2)
	}

	return keyvals2
}
