package kv

import (
	"context"
)

type ctxKeyT int

var ctxKey ctxKeyT

// Context is used for creating a context.Context
// with key/value pairs attached.
type Context struct {
	Ctx context.Context
}

// NewContext is used to create a new context. See the example
// for usage.
func NewContext(ctx context.Context) Context {
	return Context{Ctx: ctx}
}

// With returns a context.Context with the keyvals attached.
func (c Context) With(keyvals ...interface{}) context.Context {
	ctx := c.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if len(keyvals) == 0 {
		return ctx
	}
	keyvals = Flatten(keyvals)
	keyvals = append(keyvals, fromContext(ctx)...)
	return context.WithValue(ctx, ctxKey, keyvals)
}

func fromContext(ctx context.Context) []interface{} {
	var keyvals []interface{}
	if ctx != nil {
		keyvals, _ = ctx.Value(ctxKey).([]interface{})
	}
	return keyvals
}
