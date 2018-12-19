package kv

import (
	"bytes"
	"context"
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

type testKeyvalPairer struct {
	key   string
	value interface{}
}

func (p testKeyvalPairer) keyvalPair() (string, interface{}) {
	return p.key, p.value
}

func TestKeyvals(t *testing.T) {
	tests := []struct {
		keyvalser keyvalser
		want      []interface{}
	}{
		{
			keyvalser: List{"k1", 1, "k2", 2},
			want:      []interface{}{"k1", 1, "k2", 2},
		},
		{
			keyvalser: Map{"k1": 1},
			want:      []interface{}{"k1", 1},
		},
		{
			keyvalser: Map{"k1": 1, "k2": 2, "k3": 3, "a1": "1", "a3": "3", "a2": "2"},
			want:      []interface{}{"a1", "1", "a2", "2", "a3", "3", "k1", 1, "k2", 2, "k3", 3},
		},
		{
			keyvalser: Pair{"k1", 1},
			want:      []interface{}{"k1", 1},
		},
	}

	for i, tt := range tests {
		if got := tt.keyvalser.Keyvals(); !reflect.DeepEqual(tt.want, got) {
			t.Errorf("%d: want=%v, got=%v", i, tt.want, got)
		}
	}
}

func TestKeyvalPair(t *testing.T) {
	tests := []struct {
		keyvalPairer keyvalPairer
		wantKey      string
		wantValue    interface{}
	}{
		{
			keyvalPairer: testKeyvalPairer{"k1", 1},
			wantKey:      "k1",
			wantValue:    1,
		},
		{
			keyvalPairer: Pair{"k1", 1},
			wantKey:      "k1",
			wantValue:    1,
		},
	}

	for i, tt := range tests {
		gotKey, gotValue := tt.keyvalPairer.keyvalPair()
		if gotKey != tt.wantKey || !reflect.DeepEqual(tt.wantValue, gotValue) {
			t.Errorf("%d: want=[%s, %v], got=[%s, %v]", i, tt.wantKey, tt.wantValue, gotKey, gotValue)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{
			input: Pair{"key", "value"},
			want:  "key=value",
		},
		{
			input: Map{"key": "value"},
			want:  "key=value",
		},
		{
			input: List{"key1", "value1", "key2", "value2"},
			want:  "key1=value1 key2=value2",
		},
		{
			input: Parse([]byte("this is a message key1=value1 key2=value2")),
			want:  "this is a message key1=value1 key2=value2",
		},
		{
			input: Parse([]byte("message 1 key1=value1 message 2   key2=value2 message 3   key3=value3")),
			want:  "message 1 key1=value1 message 2   key2=value2 message 3 key3=value3",
		},
	}
	for i, tt := range tests {
		stringer, ok := tt.input.(fmt.Stringer)
		if !ok {
			t.Errorf("expected fmt.Stringer")
		} else {
			if got, want := stringer.String(), tt.want; got != want {
				t.Errorf("%d: got=%s want=%s", i, got, want)
			}
		}
		marshaler, ok := tt.input.(encoding.TextMarshaler)
		if !ok {
			t.Errorf("expected encoding.TextMarshaler")
		} else {
			b, err := marshaler.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := string(b), tt.want; got != want {
				t.Errorf("%d: got=%s want=%s", i, got, want)
			}
		}
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		input string
		msg   Message
	}{
		{
			input: "error: this is the message key1=value1 key2=value2: file not found\n",
			msg: Message{
				Text: "error: this is the message key1=value1 key2=value2: file not found",
			},
		},
		{
			input: `one`,
			msg: Message{
				Text: `one`,
				List: nil,
			},
		},
		{
			input: "    one\t\n   ",
			msg: Message{
				Text: `one`,
				List: nil,
			},
		},
		{
			input: `select "id","name","location" from "table" where "id" = $1 [25]`,
			msg: Message{
				Text: `select "id","name","location" from "table" where "id" = $1 [25]`,
				List: nil,
			},
		},
		{
			input: `this is the message key1=1 key2="2"`,
			msg: Message{
				Text: "this is the message",
				List: List{
					"key1", "1",
					"key2", "2",
				},
			},
		},
		{
			input: `this is the message "key1"="1" "key2"="2"`,
			msg: Message{
				Text: "this is the message",
				List: List{
					"key1", "1",
					"key2", "2",
				},
			},
		},
		{
			input: `this is the message key1=`,
			msg: Message{
				Text: "this is the message key1=",
				List: List{},
			},
		},
		{
			input: `message key1==`,
			msg: Message{
				Text: "message key1==",
				List: List{},
			},
		},
		{
			input: `message a8r5t= key1== key2="" key3=x`,
			msg: Message{
				Text: "message a8r5t= key1==",
				List: List{
					"key2", "",
					"key3", "x",
				},
			},
		},
		{ // trailing whitspace, multiple white space
			input: `message    key1=1    key2=2   `,
			msg: Message{
				Text: "message",
				List: List{
					"key1", "1",
					"key2", "2",
				},
			},
		},
		{ // missing quote
			input: `message key1="1`,
			msg: Message{
				Text: "message",
				List: List{
					"key1", "?",
				},
			},
		},
		{ // escapes
			input: `message key1="a\r\n" key2="\x41"`,
			msg: Message{
				Text: "message",
				List: List{
					"key1", "a\r\n",
					"key2", "A",
				},
			},
		},
		{ // nested message
			input: `message 1 key1=1 message 2 key2=2`,
			msg: Message{
				Text: "message 1 key1=1 message 2",
				List: List{
					"key2", "2",
				},
			},
		},
		{ // nested message with colon
			input: `message 1 key1="1": message 2 key2=2`,
			msg: Message{
				Text: `message 1 key1="1": message 2`,
				List: List{
					"key2", "2",
				},
			},
		},
		{ // nested message with colon
			input: `message 1 key1=1: message 2 key2=2`,
			msg: Message{
				Text: "message 1 key1=1: message 2",
				List: List{
					"key2", "2",
				},
			},
		},
		{ // empty input
			input: ``,
			msg:   Message{},
		},
	}

	for tn, tt := range tests {
		if got, want := Parse([]byte(tt.input)), tt.msg; !msgEqual(got, &want) {
			t.Errorf("%d, got=%v\nwant=%v", tn, got, want)
		}
		var msg Message
		if err := msg.UnmarshalText([]byte(tt.input)); err != nil {
			t.Errorf("%d: got=%v\nwant=nil", tn, err)
		}
		if got, want := msg, tt.msg; !msgEqual(&got, &want) {
			t.Errorf("%d: got=%v\nwant=%v", tn, got, want)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	if got, want := From(nil).With(), (Message{}); !msgEqual(got, &want) {
		t.Errorf("got=%v, want=%v", got, want)
	}

	if got, want := NewContext(nil).With(), context.Background(); got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
}

func TestListMsg(t *testing.T) {
	tests := []struct {
		list List
		text string
		msg  *Message
	}{
		{
			list: List{"a", 1, "b", 2},
			text: "message",
			msg: &Message{
				Text: "message",
				List: List{"a", 1, "b", 2},
			},
		},
	}
	for tn, tt := range tests {
		if got, want := tt.list.Msg(tt.text), tt.msg; !msgEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestListErr(t *testing.T) {
	tests := []struct {
		list List
		text string
		err  *Error
	}{
		{
			list: List{"a", 1, "b", 2},
			text: "message",
			err: &Error{
				Text: "message",
				List: List{"a", 1, "b", 2},
			},
		},
	}
	for tn, tt := range tests {
		if got, want := tt.list.Err(tt.text), tt.err; !errEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestListWrap(t *testing.T) {
	err1 := errors.New("test")
	tests := []struct {
		list List
		text string
		err  error
		e    *Error
	}{
		{
			list: List{"a", 1, "b", 2},
			text: "message",
			err:  err1,
			e: &Error{
				Text: "message",
				List: List{"a", 1, "b", 2},
				Err:  err1,
			},
		},
	}
	for tn, tt := range tests {
		if got, want := tt.list.Wrap(tt.err, tt.text), tt.e; !errEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestListFrom(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		list List
		ctx  context.Context
		msg  *Message
	}{
		{
			list: List{"a", 1, "b", 2},
			ctx:  NewContext(ctx).With("c", 3, "d", 4),
			msg: &Message{
				List:        List{"a", 1, "b", 2},
				ContextList: List{"c", 3, "d", 4},
			},
		},
	}
	for tn, tt := range tests {
		if got, want := tt.list.From(tt.ctx), tt.msg; !msgEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestMessageWrap(t *testing.T) {
	err1 := errors.New("test")
	ctx := NewContext(context.Background()).With("c", 3, "d", 4)
	tests := []struct {
		msg  *Message
		text []string
		err  error
		e    *Error
	}{
		{
			msg:  From(ctx).With("a", 1),
			text: []string{"message"},
			err:  err1,
			e: &Error{
				Text:        "message",
				List:        List{"a", 1},
				ContextList: List{"c", 3, "d", 4},
				Err:         err1,
			},
		},
		{
			msg:  From(ctx).With("a", 1),
			text: nil,
			err:  err1,
			e: &Error{
				Text:        "",
				List:        List{"a", 1},
				ContextList: List{"c", 3, "d", 4},
				Err:         err1,
			},
		},
	}
	for tn, tt := range tests {
		if got, want := tt.msg.Wrap(tt.err, tt.text...), tt.e; !errEqual(got, want) {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestMessageLog(t *testing.T) {
	prevLogFunc := LogFunc
	defer func() {
		LogFunc = prevLogFunc
	}()

	var output string
	LogFunc = func(args ...interface{}) {
		output = strings.TrimSpace(fmt.Sprintln(args...))
	}

	tests := []struct {
		msg  *Message
		want string
	}{
		{
			msg: &Message{
				Text:        "message",
				List:        List{"a", 1, "b", 2},
				ContextList: List{"c", 3},
			},
			want: "message a=1 b=2 c=3",
		},
	}
	for tn, tt := range tests {
		tt.msg.Log()
		if got, want := output, tt.want; got != want {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestMessageWriteToBuffer(t *testing.T) {
	tests := []struct {
		start string
		msg   *Message
		want  string
	}{
		{
			msg:  nil,
			want: "",
		},
		{
			msg:  &Message{},
			want: "",
		},
		{
			start: "xx",
			msg:   &Message{Text: "message"},
			want:  "xx message",
		},
	}
	for tn, tt := range tests {
		var buf bytes.Buffer
		buf.WriteString(tt.start)
		tt.msg.writeToBuffer(&buf)
		if got, want := buf.String(), tt.want; got != want {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestContext(t *testing.T) {
	ctx := context.Background()
	ctx, cancel1 := context.WithTimeout(ctx, time.Minute)
	defer cancel1()
	deadline, _ := ctx.Deadline()
	ctx, cancel2 := context.WithCancel(ctx)
	defer cancel2()
	ctx = context.WithValue(ctx, "key", 99)

	c := NewContext(ctx)
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

	v1 := c.Value("key").(int)
	v2 := ctx.Value("key").(int)
	if got, want := v1, v2; got != want {
		t.Fatalf("got=%v, want=%v", got, want)
	}
}

func msgEqual(m1, m2 *Message) bool {
	if m1 == nil && m2 == nil {
		return true
	}
	if m1 == nil || m2 == nil {
		return false
	}
	if m1.Text != m2.Text {
		return false
	}
	if len(m1.List) > 0 || len(m2.List) > 0 {
		if !reflect.DeepEqual(m1.List, m2.List) {
			return false
		}
	}
	if len(m1.ContextList) > 0 || len(m2.ContextList) > 0 {
		if !reflect.DeepEqual(m1.ContextList, m2.ContextList) {
			return false
		}
	}
	return true
}

func errEqual(e1, e2 *Error) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if e1.Text != e2.Text {
		return false
	}
	if e1.Err != e2.Err {
		return false
	}
	if len(e1.List) > 0 || len(e2.List) > 0 {
		if !reflect.DeepEqual(e1.List, e2.List) {
			return false
		}
	}
	if len(e1.ContextList) > 0 || len(e2.ContextList) > 0 {
		if !reflect.DeepEqual(e1.ContextList, e2.ContextList) {
			return false
		}
	}
	return true
}
