package kv

import (
	"reflect"
	"testing"
)

func TestListMarshal(t *testing.T) {
	tests := []struct {
		list        List
		text        string
		marshaled   string
		unmarshaled List
	}{
		{
			list:        List{"a", 1, "b", "value 2"},
			marshaled:   `a=1 b="value 2"`,
			unmarshaled: List{"a", "1", "b", "value 2"},
		},
		{
			list:        List{"a", 1, "b", "value 2"},
			marshaled:   `a=1 b="value 2"`,
			text:        "leading message ",
			unmarshaled: List{"msg", "leading message", "a", "1", "b", "value 2"},
		},
	}
	for tn, tt := range tests {
		b, err := tt.list.MarshalText()
		if err != nil {
			t.Error(err)
			continue
		}
		if got, want := string(b), tt.marshaled; got != want {
			t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
			continue
		}
		if tt.text != "" {
			var m = []byte(tt.text)
			m = append(m, b...)
			b = m
		}
		var l List
		if err = l.UnmarshalText(b); err != nil {
			t.Error(err)
			continue
		}
		if got, want := l, tt.unmarshaled; !reflect.DeepEqual(got, want) {
			t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
			continue
		}
	}
}

func TestListClone(t *testing.T) {
	tests := []struct {
		list List
		cap  int
	}{
		{
			list: List{},
			cap:  0,
		},
		{
			list: List{"a", 1},
			cap:  0,
		},
		{
			list: List{"a", 1},
			cap:  11,
		},
	}
	for tn, tt := range tests {
		clone := tt.list.clone(tt.cap)
		if got, want := clone, tt.list; !reflect.DeepEqual(got, want) {
			t.Errorf("%d:\n got=%+v\nwant=%+v", tn, got, want)
			continue
		}
		if got, want := clone.Keyvals(), tt.list.Keyvals(); !reflect.DeepEqual(got, want) {
			t.Errorf("%d:\n got=%+v\nwant=%+v", tn, got, want)
			continue
		}
		cloneCap := tt.cap
		if n := len(tt.list); cloneCap < n {
			cloneCap = n
		}
		if got, want := cap(clone), cloneCap; got != want {
			t.Errorf("%d:\n got=%+v\nwant=%+v", tn, got, want)
			continue
		}
	}
}

/*
func TestListWith(t *testing.T) {
	tests := []struct {
		list List
		text string
		err  *errorT
	}{
		{
			list: List{"a", 1, "b", 2},
			text: "message",
			err: &errorT{
				text: "message",
				list: List{"a", 1, "b", 2},
			},
		},
	}
	for tn, tt := range tests {
		if got, want := tt.list.NewError(tt.text).(*errorT), tt.err; !errEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}
*/
