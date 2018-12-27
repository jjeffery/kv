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
		fn   func() (err error, cause error)
		want string
	}{
		{
			fn: func() (err error, cause error) {
				return NewError("message text"), nil
			},
			want: "message text",
		},
		{
			fn: func() (error, error) {
				err := errors.New("first message")
				return Wrap(err, "second message"), err
			},
			want: "second message: first message",
		},
		{
			fn: func() (error, error) {
				return NewError("message text").With("a", 1, "b", 2), nil
			},
			want: "message text a=1 b=2",
		},
		{
			fn: func() (error, error) {
				err := NewError("first message").With("a", 1, "b", 2, "c", 3)
				return Wrap(err, "second message").With("a", 1, "b", 2, "d", 4), err
			},
			want: "second message: first message a=1 b=2 d=4 c=3",
		},
		{
			fn: func() (error, error) {
				err1 := keyvalserError{"msg", "first", "a", 1, "b", 2, "c", "3"}
				err2 := Wrap(err1, "second").With("a", 1, "b", "2")
				return Wrap(err2, "third").With("a", 2), err2
			},
			want: "third: second: first a=2 a=1 b=2 c=3",
		},
		{
			fn: func() (error, error) {
				ctx := From(context.Background()).With("c", 3, "d", 4)
				err1 := keyvalserError{"msg", "first", "a", 1, "b", 2, "c", "3"}
				err2 := Wrap(err1, "second").With("a", 1, "b", "2")
				return From(ctx).Wrap(err2, "third").With("a", 2), err2
			},
			want: "third: second: first a=2 a=1 b=2 c=3 d=4",
		},
		{
			fn: func() (error, error) {
				return From(nil).NewError("text"), nil
			},
			want: "text",
		},
		{
			fn: func() (error, error) {
				list := With("a", 1, "b", 2)
				err := list.NewError("text")
				return err, nil
			},
			want: "text a=1 b=2",
		},
		{
			fn: func() (error, error) {
				list := With("a", 1, "b", 2)
				cause := list.NewError("second")
				err := list.Wrap(cause, "first")
				return err, cause
			},
			want: "first: second a=1 b=2",
		},
		{
			fn: func() (error, error) {
				return nil, nil
			},
			want: "",
		},
	}
	for tn, tt := range tests {
		err, cause := tt.fn()
		var errText string
		if err != nil {
			errText = err.Error()
		}
		if got, want := errText, tt.want; got != want {
			t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
			continue
		}
		var unwrapped error
		if unwrap, ok := err.(interface{ Unwrap() error }); ok {
			unwrapped = unwrap.Unwrap()
		}
		if got, want := unwrapped, cause; got != want {
			t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
			continue
		}
		causer, canCause := err.(interface{ Cause() error })
		if cause != nil {
			if got, want := canCause, true; got != want {
				t.Errorf("%d: got=%v, want=%v", tn, got, want)
				continue
			}
			if got, want := causer.Cause(), cause; got != want {
				t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
				continue
			}
		} else {
			if got, want := canCause, false; got != want {
				t.Errorf("%d: got=%v, want=%v", tn, got, want)
				continue
			}
		}
	}
}

/*
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
				return NewError("error")
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
			t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
		}
	}
}
*/
