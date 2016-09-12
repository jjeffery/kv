package kv

// The keyvalser interface returns a slice of alternating keys
// and values.
type keyvalser interface {
	Keyvals() []interface{}
}

// The keyvalPairer interface returns a single key/value pair.
type keyvalPairer interface {
	KeyvalPair() (key string, value interface{})
}

// The keyvalMapper interface returns a map of keys to values.
type keyvalMapper interface {
	KeyvalMap() map[string]interface{}
}

// The keyvalsAppender interface is used for appending key/value pairs.
// This is an internal interface: the promise is that it will only
// append valid key/value pairs.
//
// This internal interface can really be removed, but at time of writing
// I'm not sure that the keyvalPairer and keyvalMapper interfaces are
// going to stay. If they do get removed, then the keyvalsAppender provides
// a sneaky way to reduce memory allocations for Pair and Map types.
// So... remove the keyvalsAppender interface if keyvalPairer and keyvalMapper
// do in fact prove their worth and are made permanent.
type keyvalsAppender interface {
	appendKeyvals(keyvals []interface{}) []interface{}
}

// List is a slice of alternating keys and values.
type List []interface{}

// Keyvals returns the list cast as []interface{}.
// It implements the keyvalser interface described in the package summary.
func (l List) Keyvals() []interface{} {
	return []interface{}(l)
}

// Pair represents a single key/value pair.
type Pair struct {
	Key   string
	Value interface{}
}

// P returns a key/value pair. The following alternatives are equivalent:
//  kv.Pair{key, value}
//  kv.P(key, value)
// The second alternative is slightly less typing, and avoids
// the following go vet warning:
//  composite literal uses unkeyed fields
func P(key string, value interface{}) Pair {
	return Pair{
		Key:   key,
		Value: value,
	}
}

// Keyvals returns the pair's key and value as a slice of interface{}.
// It implements the keyvalser interface described in the package summary.
func (p Pair) Keyvals() []interface{} {
	return []interface{}{p.Key, p.Value}
}

// KeyvalPair returns the pair's key and value. This implements
// the keyvalsPairer interface described in the package summary.
func (p Pair) KeyvalPair() (key string, value interface{}) {
	return p.Key, p.Value
}

func (p Pair) appendKeyvals(keyvals []interface{}) []interface{} {
	return append(keyvals, p.Key, p.Value)
}

// Map is a map of keys to values.
//
// Note that when a map is appended to a keyvals list of alternating
// keys and values, there is no guarantee of the order that the key/value
// pairs will be appended.
type Map map[string]interface{}

// Keyvals returns the contents of the map as a list of alternating
// key/value pairs. It implements the keyvalser interface described
// in the package summary.
func (m Map) Keyvals() []interface{} {
	keyvals := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		keyvals = append(keyvals, k, v)
	}
	return keyvals
}

// KeyvalMap returns the map cast as a map[string]interface{}.
// It implements the keyvalMapper interface described in the package summary.
func (m Map) KeyvalMap() map[string]interface{} {
	return map[string]interface{}(m)
}

func (m Map) appendKeyvals(keyvals []interface{}) []interface{} {
	for key, value := range m {
		keyvals = append(keyvals, key, value)
	}
	return keyvals
}
