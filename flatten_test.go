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
			alt: []interface{}{
				"key2", 2, "key1", "val1", "key3", 3,
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
				List{
					Map{
						"key1": "val1",
					},
					"key2", 2,
					"key3", 3.0,
				},
				"key4", "4",
				List{
					List{
						P("key5", 5),
					},
				},
			},
			want: []interface{}{
				"key1", "val1", "key2", 2, "key3", 3.0, "key4", "4", "key5", 5,
			},
		},
		{
			v: []interface{}{
				"message text",
				testKeyvalPairer{"k1", 1},
				testKeyvalPairer{"k2", "2"},
			},
			want: []interface{}{
				"msg", "message text", "k1", 1, "k2", "2",
			},
		},
		{
			v: []interface{}{
				"message text",
				Map{"k1": 1, "k2": "2"},
				testKeyvalPairer{"k3", 3},
			},
			want: []interface{}{
				"msg", "message text", "k1", 1, "k2", "2", "k3", 3,
			},
			alt: []interface{}{
				"msg", "message text", "k2", "2", "k1", 1, "k3", 3,
			},
		},
		{
			v:    []interface{}{io.EOF},
			want: []interface{}{"msg", io.EOF},
		},
		{
			v:    []interface{}{"msg", "the message", io.EOF},
			want: []interface{}{"msg", "the message", "error", io.EOF},
		},
		{
			v:    []interface{}{"not found", "id", "A12345678"},
			want: []interface{}{"msg", "not found", "id", "A12345678"},
		},
		{
			v:    []interface{}{"not_found", "id", "A12345678"},
			want: []interface{}{"msg", "not_found", "id", "A12345678"},
		},
		{
			v:    []interface{}{1, 2, 3},
			want: []interface{}{"_p1", 1, "_p2", 2, "_p3", 3},
		},
		{
			v:    []interface{}{1, 2, 3, 4},
			want: []interface{}{"_p1", 1, "_p2", 2, "_p3", 3, "_p4", 4},
		},
		{
			v:    []interface{}{io.EOF, 2, 3, 4},
			want: []interface{}{"msg", io.EOF, "_p1", 2, "_p2", 3, "_p3", 4},
		},
		{
			v:    []interface{}{"msg", "message 4", "key1", 1, 2},
			want: []interface{}{"msg", "message 4", "key1", 1, "_p1", 2},
		},
		{
			v:    []interface{}{"msg", "server listening", "HTTP", "addr", ":6060"},
			want: []interface{}{"msg", "server listening", "_p1", "HTTP", "addr", ":6060"},
		},
		{
			v:    []interface{}{"msg", "listening", "transport", "HTTP", ":6060"},
			want: []interface{}{"msg", "listening", "transport", "HTTP", "_p1", ":6060"},
		},
		{
			v:    []interface{}{"msg", "level", "id"},
			want: []interface{}{"_p1", "msg", "level", "id"},
		},
		{
			v:    []interface{}{"installing", "level", "id"},
			want: []interface{}{"msg", "installing", "level", "id"},
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
