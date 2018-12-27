package kv

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestContextMethods(t *testing.T) {
	type keyT string
	key := keyT("key")
	ctx := context.Background()
	ctx, cancel1 := context.WithTimeout(ctx, time.Minute)
	defer cancel1()
	deadline, _ := ctx.Deadline()
	ctx, cancel2 := context.WithCancel(ctx)
	defer cancel2()
	ctx = context.WithValue(ctx, key, 99)

	c := From(ctx)
	d, _ := c.Deadline()
	if got, want := d, deadline; !got.Equal(want) {
		t.Fatalf("got=%v, want=%v", got, want)
	}

	if got, want := c.Done(), ctx.Done(); got != want {
		t.Fatalf("got=%v, want=%v", got, want)
	}

	cancel2()

	if got, want := c.Err(), ctx.Err(); got != want {
		t.Fatalf("got=%v, want=%v", got, want)
	}

	v1 := c.Value(key).(int)
	v2 := ctx.Value(key).(int)
	if got, want := v1, v2; got != want {
		t.Fatalf("got=%v, want=%v", got, want)
	}
}

func TestContext(t *testing.T) {
	tests := []struct {
		fn   func() context.Context
		text string
	}{
		{
			fn: func() context.Context {
				ctx := context.Background()
				return From(ctx)
			},
			text: "",
		},
		{
			fn: func() context.Context {
				return From(nil)
			},
			text: "",
		},
		{
			fn: func() context.Context {
				return From(nil).With()
			},
			text: "",
		},
		{
			fn: func() context.Context {
				return With("a", 1, "b", "two").From(context.Background())
			},
			text: "a=1 b=two",
		},
	}

	for tn, tt := range tests {
		ctx := tt.fn()
		if got, want := fmt.Sprint(ctx), tt.text; got != want {
			t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
		}
		stringer := ctx.(fmt.Stringer)
		if got, want := stringer.String(), tt.text; got != want {
			t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
		}
	}
}

func TestContextFormat(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	want := fmt.Sprintf("%v", ctx)
	ctx = From(ctx).With("a", 1)
	got := fmt.Sprintf("%+v", ctx)
	if !strings.HasPrefix(got, want) {
		t.Errorf("\n got=%v\nwant=%v", got, want)
	}
}
