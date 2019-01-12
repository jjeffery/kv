package logfmt

import (
	"bytes"
	"errors"
	"testing"
)

func TestWriteKeyValue(t *testing.T) {
	tests := []struct {
		key   interface{}
		value interface{}
		want  string
	}{
		{
			key:   "key",
			value: "value",
			want:  "key=value",
		},
		{
			key:   "key=",
			value: `some value`,
			want:  `key_="some value"`,
		},
		{
			key:   "the key",
			value: "the value",
			want:  `the_key="the value"`,
		},
		{
			key:   "the\tkey\x03",
			value: "the\tvalue\x02",
			want:  `the_key_="the\tvalue\x02"`,
		},
		{
			key:   17,
			value: 25,
			want:  "17=25",
		},
		{
			key:   complex(1, 4),
			value: complex(float64(100.99), float64(1.19e23)),
			want:  "(1+4i)=(100.99+1.19e+23i)",
		},
		{
			key:   "",
			value: "",
			want:  `EMPTY=""`,
		},
		{
			key:   nil,
			value: nil,
			want:  "null=null",
		},
		{
			key:   []byte(nil),
			value: []byte(nil),
			want:  "null=null",
		},
		{
			key:   errors.New("key"),
			value: errors.New("value"),
			want:  "key=value",
		},
		{
			key:   testTextMarshaler("key"),
			value: testTextMarshaler("value"),
			want:  "key=value",
		},
		{
			key:   testStringer("key"),
			value: testStringer("value"),
			want:  "key=value",
		},
		{
			key:   panicingTextMarshaler("key"),
			value: panicingTextMarshaler("value"),
			want:  "PANIC=PANIC",
		},
		{
			key:   "key",
			value: panicingTextMarshaler("value"),
			want:  "key=PANIC",
		},
		{
			key:   panicingTextMarshaler("key"),
			value: "value",
			want:  "PANIC=value",
		},
		{
			key:   failingTextMarshaler("key"),
			value: failingTextMarshaler("value"),
			want:  "ERROR=ERROR",
		},
		{
			key:   "key",
			value: failingTextMarshaler("value"),
			want:  "key=ERROR",
		},
		{
			key:   failingTextMarshaler("key"),
			value: "value",
			want:  "ERROR=value",
		},
		{
			key:   func() *string { s := "key"; return &s }(),
			value: func() *string { s := "value"; return &s }(),
			want:  "key=value",
		},
		{
			key:   func() *string { return nil }(),
			value: func() *string { return nil }(),
			want:  "null=null",
		},
		{
			key:   struct{ v int }{v: 25},
			value: struct{ v int }{v: 17},
			want:  `{25}="{17}"`,
		},
		{
			key:   "key",
			value: "value:",
			want:  `key="value:"`,
		},
	}
	for i, tt := range tests {
		doTest := func(key interface{}, value interface{}, want string) {
			var tt interface{} // hides outer tt to avoid error
			_ = tt

			var buf bytes.Buffer
			WriteKeyValue(&buf, key, value)
			if got := buf.String(); got != want {
				t.Errorf("%d: got `%s` want `%s`", i, got, want)
			}
		}
		doTest(tt.key, tt.value, tt.want)

		if s, ok := tt.key.(string); ok {
			// key is a string, test for []byte as well
			key := []byte(s)
			doTest(key, tt.value, tt.want)
		}
		if s, ok := tt.value.(string); ok {
			// value is a string, test for []byte as well
			value := []byte(s)
			doTest(tt.key, value, tt.want)
		}
	}
}

type testStringer string

func (t testStringer) String() string {
	return string(t)
}

type testTextMarshaler string

func (t testTextMarshaler) MarshalText() ([]byte, error) {
	return []byte(t), nil
}

type panicingTextMarshaler string

func (t panicingTextMarshaler) MarshalText() ([]byte, error) {
	panic(string(t))
}

type failingTextMarshaler string

func (t failingTextMarshaler) MarshalText() ([]byte, error) {
	return nil, errors.New(string(t))
}
