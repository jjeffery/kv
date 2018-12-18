package kv

import (
	"context"
	"errors"
	"testing"
)

// keyvalserError is an error that implements the keyvalser interface.
// Used for testing only.
type keyvalserError []interface{}

func (e keyvalserError) Error() string {
	return List(e).String()
}

func (e keyvalserError) Keyvals() []interface{} {
	return e
}

func TestError(t *testing.T) {
	tests := []struct {
		fn   func() error
		want string
	}{
		{
			fn: func() error {
				return Err("message text")
			},
			want: "message text",
		},
		{
			fn: func() error {
				err := errors.New("first message")
				return Wrap(err, "second message")
			},
			want: "second message: first message",
		},
		{
			fn: func() error {
				return Err("message text").With("a", 1, "b", 2)
			},
			want: "message text a=1 b=2",
		},
		{
			fn: func() error {
				err := Err("first message").With("a", 1, "b", 2, "c", 3)
				return Wrap(err, "second message").With("a", 1, "b", 2, "d", 4)
			},
			want: "second message: first message a=1 b=2 d=4 c=3",
		},
		{
			fn: func() error {
				err1 := keyvalserError{"msg", "first", "a", 1, "b", 2, "c", "3"}
				err2 := Wrap(err1, "second").With("a", 1, "b", "2")
				return Wrap(err2, "third").With("a", 2)
			},
			want: "third: second: first a=2 a=1 b=2 c=3",
		},
		{
			fn: func() error {
				ctx := NewContext(context.Background()).With("c", 3, "d", 4)
				err1 := keyvalserError{"msg", "first", "a", 1, "b", 2, "c", "3"}
				err2 := Wrap(err1, "second").With("a", 1, "b", "2")
				return Wrap(err2, "third").With("a", 2).Ctx(ctx)
			},
			want: "third: second: first a=2 a=1 b=2 c=3 d=4",
		},
		{
			fn: func() error {
				return nil
			},
			want: "",
		},
	}
	for tn, tt := range tests {
		err := tt.fn()
		var got string
		if err != nil {
			got = err.Error()
		}
		if want := tt.want; got != want {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestUnwrap(t *testing.T) {
	err1 := errors.New("error 1")
	tests := []struct {
		fn   func() error
		want error
	}{
		{
			fn: func() error {
				return Wrap(err1)
			},
			want: err1,
		},
		{
			fn: func() error {
				return Err("error")
			},
			want: nil,
		},
		{
			fn: func() error {
				return nil
			},
			want: nil,
		},
	}
	for tn, tt := range tests {
		err := tt.fn()
		var got error
		if unwrap, ok := err.(interface{ Unwrap() error }); ok {
			got = unwrap.Unwrap()
		}
		if want := tt.want; got != want {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}
