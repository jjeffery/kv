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
		{
			v:    []interface{}{List{"a", 1, "b", 2}},
			want: []interface{}{"a", 1, "b", 2},
		},
		{
			v:    []interface{}{testKeyvalser{}, "5", 6},
			want: []interface{}{"1", "2", "3", "4", "5", 6},
		},
	}

	for i, tt := range tests {
		got := flattenFix(tt.v)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d: want %v, got %v", i, tt.want, got)
		}
	}
}

type testKeyvalser struct{}

func (tkv testKeyvalser) Keyvals() []interface{} {
	return []interface{}{"1", "2", "3", "4"}
}
