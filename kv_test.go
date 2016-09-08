package kv

import (
	"io"
	"reflect"
	"testing"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		v    []interface{}
		want []interface{}
		alt  []interface{}
	}{
		{
			v: []interface{}{
				"key1", "val1", "key2", 2,
			},
			want: []interface{}{
				"key1", "val1", "key2", 2,
			},
		},
		{
			v: []interface{}{
				P("key1", "val1"),
				P("key2", 2),
			},
			want: []interface{}{
				"key1", "val1", "key2", 2,
			},
			alt: []interface{}{
				"key2", 2, "key1", "val1",
			},
		},
		{
			v: []interface{}{
				Map{
					"key1": "val1",
					"key2": 2,
				},
				P("key3", 3),
			},
			want: []interface{}{
				"key1", "val1", "key2", 2, "key3", 3,
			},
		},
		{
			v: []interface{}{
				Map{
					"key1": "val1",
				},
				"key2", 2,
				"key3", 3.0,
				P("key4", 4),
			},
			want: []interface{}{
				"key1", "val1", "key2", 2, "key3", 3.0, "key4", 4,
			},
		},
		{
			v: []interface{}{
				Keyvals{
					Map{
						"key1": "val1",
					},
					"key2", 2,
					"key3", 3.0,
				},
				"key4", "4",
				Keyvals{
					Keyvals{
						P("key5", 5),
					},
				},
			},
			want: []interface{}{
				"key1", "val1", "key2", 2, "key3", 3.0, "key4", "4", "key5", 5,
			},
		},
		{
			v:    []interface{}{io.EOF},
			want: []interface{}{"error", io.EOF},
		},
		{
			v:    []interface{}{"msg", "the message", io.EOF},
			want: []interface{}{"msg", "the message", "error", io.EOF},
		},
	}

	for i, tt := range tests {
		got := Flatten(tt.v)
		if !reflect.DeepEqual(got, tt.want) &&
			!reflect.DeepEqual(got, tt.alt) {
			t.Errorf("%d: want %v, got %v", i, tt.want, got)
		}
	}
}
