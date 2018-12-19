package kv

import (
	"context"
	"time"
)

type ctxKeyT int

var ctxKey ctxKeyT

// Context implements the context.Context interface,
// and can create a new context with key/value pairs
// attached to it.
//
// It's a bit of a pity that this is the first type that appears in the
// Godoc documentation, because this really isn't used all that much.
// If you're reading the Godoc from the top, look at List and Message before
// coming back to contexts.
type Context interface {
	context.Context

	// With returns a new context with keyvals attached.
	With(keyvals ...interface{}) context.Context
}

type kvContext struct {
	ctx context.Context
}

// NewContext is used to create a new context. See the example
// for usage.
func NewContext(ctx context.Context) Context {
	return kvContext{ctx: ctx}
}

// Deadline implements the context.Context interface.
func (c kvContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

// Done implements the context.Context interface.
func (c kvContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err implements the context.Context interface.
func (c kvContext) Err() error {
	return c.ctx.Err()
}

// Value implements the context.Context interface.
func (c kvContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// With returns a context.Context with the keyvals attached.
func (c kvContext) With(keyvals ...interface{}) context.Context {
	return newContext(c.ctx, keyvals)
}

func newContext(ctx context.Context, keyvals []interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(keyvals) == 0 {
		return ctx
	}
	keyvals = Flatten(keyvals)
	keyvals = append(keyvals, fromContext(ctx)...)
	keyvals = keyvals[:len(keyvals):len(keyvals)] // set capacity
	return context.WithValue(ctx, ctxKey, keyvals)
}

func fromContext(ctx context.Context) []interface{} {
	var keyvals []interface{}
	if ctx != nil {
		keyvals, _ = ctx.Value(ctxKey).([]interface{})
	}
	return keyvals
}
