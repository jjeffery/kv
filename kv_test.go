package kv

import (
	"context"
	"encoding"
	"fmt"
	"reflect"
	"testing"
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
			want:  "message 1 key1=value1 message 2 key2=value2 message 3 key3=value3",
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
				Text: "this is the message",
				List: List{
					"key1", "",
				},
			},
		},
		{
			input: `message key1==`,
			msg: Message{
				Text: "message",
				List: List{
					"key1", "=",
				},
			},
		},
		{
			input: `message key1== key2= key3=x`,
			msg: Message{
				Text: "message",
				List: List{
					"key1", "=",
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
					"key1", "1",
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
				Text: "message 1",
				List: List{
					"key1", "1",
				},
				Next: &Message{
					Text: "message 2",
					List: List{
						"key2", "2",
					},
				},
			},
		},
		{ // nested message with colon
			input: `message 1 key1="1": message 2 key2=2`,
			msg: Message{
				Text: "message 1",
				List: List{
					"key1", "1",
				},
				Next: &Message{
					Text: "message 2",
					List: List{
						"key2", "2",
					},
				},
			},
		},
		{ // nested message with colon
			input: `message 1 key1=1: message 2 key2=2`,
			msg: Message{
				Text: "message 1",
				List: List{
					"key1", "1",
				},
				Next: &Message{
					Text: "message 2",
					List: List{
						"key2", "2",
					},
				},
			},
		},
		{ // empty input
			input: ``,
			msg:   Message{},
		},
	}

	for tn, tt := range tests {
		if got, want := Parse([]byte(tt.input)), tt.msg; !msgEqual(&got, &want) {
			t.Errorf("%d, got=%v, want=%v", tn, got, want)
		}
		var msg Message
		if err := msg.UnmarshalText([]byte(tt.input)); err != nil {
			t.Errorf("%d: got=%v, want=nil", tn, err)
		}
		if got, want := msg, tt.msg; !msgEqual(&got, &want) {
			t.Errorf("%d: got=%v, want=%v", tn, got, want)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	if got, want := Ctx(nil).With(), (Message{}); !msgEqual(&got, &want) {
		t.Errorf("got=%v, want=%v", got, want)
	}

	if got, want := NewContext(nil).With(), context.Background(); got != want {
		t.Errorf("got=%v, want=%v", got, want)
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
	if !msgEqual(m1.Next, m2.Next) {
		return false
	}
	return true
}
