package kv

import (
	"regexp"
	"strconv"
)

/*
var debug debugT = true

type debugT bool

func (d debugT) Printf(format string, v ...interface{}) {
	if d {
		s := fmt.Sprintf(format, v...)
		fmt.Println(s)
	}
}
*/

// The keyvalsAppender interface is used for appending key/value pairs
type keyvalsAppender interface {
	appendKeyvals(keyvals []interface{}) []interface{}
}

// The keyvalser interface returns a slice of alternating keys
// and values.
type keyvalser interface {
	Keyvals() []interface{}
}

// Keyvals is a variadic slice of alternating keys and values.
type Keyvals []interface{}

// Pair represents a single key/value pair.
type Pair struct {
	Key   string
	Value interface{}
}

// P returns a key/value pair.
func P(key string, value interface{}) Pair {
	return Pair{
		Key:   key,
		Value: value,
	}
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

func (m Map) appendKeyvals(keyvals []interface{}) []interface{} {
	for key, value := range m {
		keyvals = append(keyvals, key, value)
	}
	return keyvals
}

func isEven(i int) bool {
	return (i & 0x01) == 0
}

func isOdd(i int) bool {
	return (i & 0x01) != 0
}

// Flatten accepts a keyvals slice and "flattens" it into a slice
// of alternating key/value pairs. See the examples.
//
// Flatten will also check the result to ensure it is a valid
// slice of key/value pairs according to the following rules.
//  * Must have an even number of items in the slice
//  * Items at even indexes must be strings
// Flatten will insert keys into the array to ensure that the returned
// slice conforms.
func Flatten(keyvals []interface{}) []interface{} {
	const keyMsg = "msg"
	const keyError = "error"
	const keyMissingPrefix = "_p"

	// Indicates whether the keyvals slice needs to be flattened.
	// Start with true if it has an odd number of items.
	requiresFlattening := isOdd(len(keyvals))

	// Used for estimating the size of the flattened slice
	// in an attempt to use one memory allocation.
	var estimatedLen int

	// Do the keyvals include a "msg" key. This is not entirely
	// reliable if the "msg" is supposed to be a value.
	var haveMsg bool

	for i, val := range keyvals {
		switch v := val.(type) {
		case Map:
			requiresFlattening = true
			estimatedLen += len(v) * 2
		case Pair:
			requiresFlattening = true
			estimatedLen += 2
		case Keyvals:
			requiresFlattening = true
			// TODO(jpj): recursively descending into the keyvals
			// will come up with a reasonably length estimate, but
			// for now just double the number of elements in the slice
			// and this is probably accurate enough.
			estimatedLen = len(v) * 2
		case keyvalsAppender:
			requiresFlattening = true
			// some unknown Keyvals appender: not possible to estimate
			// so just use a constant.
			estimatedLen += 8
		case keyvalser:
			requiresFlattening = true
			estimatedLen += 8
		case string:
			if v == keyMsg {
				// Remember that we already have a "msg" key, which
				// will be used for inferring missing key names later.
				haveMsg = true
			}
		default:
			estimatedLen++
			if isEven(i) {
				// will probably need another item in the array
				estimatedLen++
				requiresFlattening = true
			}
		}
	}

	if !requiresFlattening {
		return keyvals
	}

	var hasMissingKeys bool
	missingKey := func(v interface{}) interface{} {
		hasMissingKeys = true
		return missingKeyT("MISSING")
	}

	// In most circumstances this output slice will have the
	// required capacity.
	output := make([]interface{}, 0, estimatedLen)
	output = flatten(output, keyvals, missingKey)

	if hasMissingKeys {
		counter := 0
		for i, v := range output {
			if _, ok := v.(missingKeyT); ok {
				// assign a name for the missing key, depends on the type
				// of the value associated with the key
				var keyName string
				switch output[i+1].(type) {
				case string:
					if !haveMsg {
						// If there is no 'msg' key, the first string
						// value gets 'msg' as its key.
						haveMsg = true
						keyName = keyMsg
					}
				case error:
					if haveMsg {
						// If there is already a 'msg' key, then an
						// error gets 'error' as the key.
						keyName = keyError
					} else {
						// If there is no 'msg' key, the first error
						// value gets 'msg' as its key.
						haveMsg = true
						keyName = keyMsg
					}
				}
				if keyName == "" {
					// Otherwise, missing keys all have a prefix that is
					// unlikely to clash with others key names, and are
					// numbered from 1.
					counter++
					keyName = keyMissingPrefix + strconv.Itoa(counter)
				}
				output[i] = keyName
			}
		}
	}

	return output
}

// The missingKeyT type is used as a placeholder for missing keys.
// Once all the missing keys are inserted, they are numbered from left
// to right.
type missingKeyT string

func flatten(output []interface{}, input []interface{}, missingKeyName func(interface{}) interface{}) []interface{} {
	for len(input) > 0 {
		if i := countNonKeyvalAppenders(input); i > 0 {
			output = flattenScalars(output, input[:i], missingKeyName)
			input = input[i:]
			continue
		}

		// At this point the first item in the input is a KeyvalsAppender.
		switch v := input[0].(type) {
		case Keyvals:
			// Treat the Keyvals type as an interface{} slice on its own
			// because it may need flattening and inserting of keys.
			output = flatten(output, []interface{}(v), missingKeyName)
		case keyvalsAppender:
			output = v.appendKeyvals(output)
		case keyvalser:
			output = flatten(output, v.Keyvals(), missingKeyName)
		default:
			//panic("cannot happen")
		}

		input = input[1:]
	}

	return output
}

// countNonKeyvalAppenders returns the count of items in input up to but
// not including the first KeyvalsAppender.
func countNonKeyvalAppenders(input []interface{}) int {
	for i := 0; i < len(input); i++ {
		switch input[i].(type) {
		case keyvalsAppender:
			return i
		case Keyvals:
			return i
		case keyvalser:
			return i
		}
	}
	return len(input)
}

// flattenScalars adjusts a list of items, none of which are KeyvalsAppenders.
// Ideally the list will have an even number of items, with strings in the
// even indices. If it doesn't, this method will fix it.
func flattenScalars(output []interface{}, input []interface{}, missingKeyName func(interface{}) interface{}) []interface{} {
	for len(input) > 0 {
		var needsFixing bool

		if isOdd(len(input)) {
			needsFixing = true
		} else {
			// check for non-string in an even position
			for i := 0; i < len(input); i += 2 {
				switch input[i].(type) {
				case string, missingKeyT:
					break
				default:
					needsFixing = true
				}
			}
		}

		if !needsFixing {
			output = append(output, input...)
			input = nil
			continue
		}

		// Build a classification of items in the array. This will be used
		// to determine the most likely position of missing key(s).
		// TODO(jpj): this could be allocated from a sync.Pool
		type classificationT byte
		const (
			stringKey classificationT = iota
			stringPossibleKey
			stringValue
			errorVar
			otherType
		)
		classifications := make([]classificationT, len(input))
		getKeyName := func(i int) interface{} {
			return missingKeyName(input[i])
		}

		for i := 0; i < len(input); i++ {
			switch v := input[i].(type) {
			case string:
				if _, ok := knownKeys[v]; ok {
					classifications[i] = stringKey
				} else if possibleKeyRE.MatchString(v) {
					classifications[i] = stringPossibleKey
				} else {
					classifications[i] = stringValue
				}
			case missingKeyT:
				classifications[i] = stringKey
			default:
				classifications[i] = otherType
			}
		}

		if len(input) == 1 {
			// Only one parameter, give it a key name. If it is a string it might
			// be the 'msg' parameter.
			output = append(output, getKeyName(0))
			output = append(output, input[0])
			input = nil
			continue
		}

		// Function to insert a key before an item that is either unlikely
		// or impossible to be a key. Returns true if something was inserted.
		// Note that this function assumes that there are at least two items
		// in the slice, which is guaranteed at this point.
		insertKeyFromBack := func(c classificationT) bool {
			// Start at the second last item
			for i := len(input) - 2; i > 0; i -= 2 {
				if classifications[i] == c {
					if isEven(len(input)) {
						input = insertKeyAt(input, i, getKeyName(i))
					} else {
						input = insertKeyAt(input, i+1, getKeyName(i+1))
					}
					return true
				}
			}
			return false
		}
		if insertKeyFromBack(otherType) {
			continue
		}
		if insertKeyFromBack(errorVar) {
			continue
		}
		if insertKeyFromBack(stringValue) {
			continue
		}
		insertKeyFromFront := func(c classificationT) bool {
			for i := 0; i < len(input); i += 2 {
				if classifications[i] == c {
					input = insertKeyAt(input, i, getKeyName(i))
					return true
				}
			}
			return false
		}
		if insertKeyFromFront(otherType) {
			continue
		}
		if insertKeyFromFront(errorVar) {
			continue
		}
		if insertKeyFromFront(stringValue) {
			continue
		}
		if insertKeyFromFront(stringPossibleKey) {
			continue
		}
		input = insertKeyAt(input, 0, getKeyName(0))
	}
	return output
}

var possibleKeyRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func insertKeyAt(input []interface{}, index int, keyName interface{}) []interface{} {
	newInput := make([]interface{}, 0, len(input)+1)
	if index > 0 {
		newInput = append(newInput, input[0:index]...)
	}
	newInput = append(newInput, keyName)
	newInput = append(newInput, input[index:]...)
	return newInput
}

// this could be public and configurable
var knownKeys = map[string]struct{}{
	"msg":   struct{}{},
	"level": struct{}{},
	"id":    struct{}{},
}
