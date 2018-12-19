package kv

import (
	"bytes"
	"context"
	"strings"
)

// Error implements the builtin error interface.
type Error struct {
	Text        string
	List        List
	ContextList List
	Err         error
}

// Err returns an error that formats as the given text.
func Err(text string) *Error {
	return &Error{
		Text: text,
	}
}

// Wrap returns an error that wraps err, optionally annotating
// with the message text.
func Wrap(err error, text ...string) *Error {
	e := &Error{
		Text: strings.Join(text, " "),
		Err:  err,
	}

	return e
}

func (e *Error) clone() *Error {
	e2 := *e
	return &e2
}

// Error implements the error interface.
//
// The string returned prints the error text of this error
// any any wrapped errors, each separated by a colon and a space (": ").
// After the error message (or messages) comes the key/value pairs.
func (e *Error) Error() string {
	var texts []string
	var list List
	var ctxList List

	addError := func(e *Error) {
		if e.Text != "" {
			texts = append(texts, e.Text)
		}
		list = list.With(e.List...)
		ctxList = ctxList.With(e.ContextList...)
	}

	addError(e)

	for err := e.Err; err != nil; {
		if e2, ok := err.(*Error); ok {
			addError(e2)
			err = e2.Err
			continue
		}
		if keyvals, ok := err.(keyvalser); ok {
			kvlist := Flatten(keyvals.Keyvals())
			for i := 0; i < len(kvlist); i += 2 {
				key := kvlist[i].(string)
				val := kvlist[i+1]
				if strings.EqualFold(key, "msg") || strings.EqualFold(key, "message") {
					texts = append(texts, valueString(val))
				} else {
					list = append(list, key, val)
				}
			}
			err = nil
			continue
		}
		if text := err.Error(); text != "" {
			texts = append(texts, text)
		}
		err = nil
	}

	list = list.With(ctxList...).dedup()
	text := strings.TrimSpace(strings.Join(texts, ": "))
	var buf bytes.Buffer
	buf.WriteString(text)
	list.writeToBuffer(&buf)
	return buf.String()
}

// Unwrap implements the Wrapper interface.
//
// See https://go.googlesource.com/proposal/+/master/design/go2draft-error-inspection.md
func (e *Error) Unwrap() error {
	return e.Err
}

// With returns an error based on e, but with additional key/value
// pairs associated.
func (e *Error) With(keyvals ...interface{}) *Error {
	e = e.clone()
	e.List = e.List.With(keyvals...)
	return e
}

// Ctx returns an error based on e, but with additional key/value
// pairs extracted from the context.
func (e *Error) Ctx(ctx context.Context) *Error {
	e = e.clone()
	e.ContextList = fromContext(ctx)
	return e
}
