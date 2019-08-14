package parse

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jjeffery/kv/internal/logfmt"
)

func b(s string) []byte {
	return []byte(s)
}

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		msg   Message
	}{
		{
			input: "error: this is the message key1=value1 key2=value2: file not found\n",
			msg: Message{
				Text: b("error: this is the message key1=value1 key2=value2: file not found"),
			},
		},
		{
			input: `one`,
			msg: Message{
				Text: b(`one`),
				List: nil,
			},
		},
		{
			input: "    one\t\n   ",
			msg: Message{
				Text: b(`one`),
				List: nil,
			},
		},
		{
			input: `select "id","name","location" from "table" where "id" = $1 [25]`,
			msg: Message{
				Text: b(`select "id","name","location" from "table" where "id" = $1 [25]`),
				List: nil,
			},
		},
		{
			input: `this is the message key1=1 key2="2"`,
			msg: Message{
				Text: b("this is the message"),
				List: [][]byte{
					b("key1"), b("1"),
					b("key2"), b("2"),
				},
			},
		},
		{
			input: `this is the message "key1"="1" "key2"="2"`,
			msg: Message{
				Text: b("this is the message"),
				List: [][]byte{
					b("key1"), b("1"),
					b("key2"), b("2"),
				},
			},
		},
		{
			input: `this is the message "key1"= "1" "key2"="2"`,
			msg: Message{
				Text: b("this is the message \"key1\"= \"1\""),
				List: [][]byte{
					b("key2"), b("2"),
				},
			},
		},
		{
			input: `this is the message key1=`,
			msg: Message{
				Text: b("this is the message key1="),
				List: [][]byte{},
			},
		},
		{
			input: `message key1==`,
			msg: Message{
				Text: b("message key1=="),
				List: [][]byte{},
			},
		},
		{
			input: `message a8r5t= key1== key2="" key3=x`,
			msg: Message{
				Text: b("message a8r5t= key1=="),
				List: [][]byte{
					b("key2"), b(""),
					b("key3"), b("x"),
				},
			},
		},
		{ // trailing whitspace, multiple white space
			input: `message    key1=1    key2=2   `,
			msg: Message{
				Text: b("message"),
				List: [][]byte{
					b("key1"), b("1"),
					b("key2"), b("2"),
				},
			},
		},
		{ // missing quote
			input: `message key1="1`,
			msg: Message{
				Text: b("message"),
				List: [][]byte{
					b("key1"), b("1"),
				},
			},
		},
		{ // escapes
			input: `message key1="a\r\n" key2="\x41\u0042"`,
			msg: Message{
				Text: b("message"),
				List: [][]byte{
					b("key1"), b("a\r\n"),
					b("key2"), b("AB"),
				},
			},
		},
		{ // nested message
			input: `message 1 key1=1 message 2 key2=2`,
			msg: Message{
				Text: b("message 1 key1=1 message 2"),
				List: [][]byte{
					b("key2"), b("2"),
				},
			},
		},
		{ // nested message with colon
			input: `message 1 key1="1": message 2 key2=2`,
			msg: Message{
				Text: b(`message 1 key1="1": message 2`),
				List: [][]byte{
					b("key2"), b("2"),
				},
			},
		},
		{ // nested message with colon
			input: `message 1 key1=1: message 2 key2=2`,
			msg: Message{
				Text: b("message 1 key1=1: message 2"),
				List: [][]byte{
					b("key2"), b("2"),
				},
			},
		},
		{ // invalid utf-8 is left as-is
			input: "invalid message \xfe a=2",
			msg: Message{
				Text: b("invalid message \xfe"),
				List: [][]byte{
					b("a"), b("2"),
				},
			},
		},
		{ // empty input
			input: ``,
			msg:   Message{},
		},
		{
			input: `text a=1 b=2 c=3 d=4 e=5 f=6 g=7 h=8 i=9 j=10`,
			msg: Message{
				Text: b("text"),
				List: [][]byte{
					b("a"), b("1"),
					b("b"), b("2"),
					b("c"), b("3"),
					b("d"), b("4"),
					b("e"), b("5"),
					b("f"), b("6"),
					b("g"), b("7"),
					b("h"), b("8"),
					b("i"), b("9"),
					b("j"), b("10"),
				},
			},
		},
	}

	for tn, tt := range tests {
		t.Run(fmt.Sprint(tn), func(t *testing.T) {
			msg := Bytes([]byte(tt.input))
			defer releaseMessage(t, msg)

			if got, want := msg, &tt.msg; !msgEqual(got, want) {
				t.Errorf("%d:\n got=%v\nwant=%v", tn, got, want)
			}
		})
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		input    []byte
		unquoted string
		before   int
		after    int
	}{
		{
			input:    nil,
			unquoted: "???",
			before:   8,
			after:    8,
		},
		{
			input:    b(`"no escape"`),
			unquoted: "no escape",
			before:   32,
			after:    32,
		},
		{
			input:    b(`"\"escape\""`),
			unquoted: `"escape"`,
			before:   32,
			after:    24,
		},
		{
			input:    b(`"unicode \u0041"`),
			unquoted: `unicode A`,
			before:   32,
			after:    23,
		},
		{
			input:    b(`"unicode \u20Ac"`),
			unquoted: "unicode \u20ac",
			before:   32,
			after:    21,
		},
		{
			input:    b(`"too long\n to fit into\n the buffer"`),
			unquoted: "too long\n to fit into\n the buffer",
			before:   8,
			after:    0,
		},
		{
			input:    b(`"invalid\"`),
			unquoted: "???",
			before:   8,
			after:    8,
		},
	}
	for tn, tt := range tests {
		input := tt.input
		buf := make([]byte, tt.before)
		unquoted, buf := unquote(input, buf)
		if got, want := string(unquoted), tt.unquoted; got != want {
			t.Errorf("%d: got=%v want=%v", tn, got, want)
			continue
		}
		if got, want := len(buf), tt.after; got != want {
			t.Errorf("%d: got=%v want=%v", tn, got, want)
		}
	}
}

// releaseMessage releases a message and ensures that its
// contents have been erased
func releaseMessage(t *testing.T, msg *Message) {
	t.Helper()
	msg.Release()

	if got := msg.Text; got != nil {
		t.Errorf("got=%v want=nil", got)
	}
	for i, v := range msg.List {
		if got := v; got != nil {
			t.Errorf("%d: got=%v want=nil", i, got)
		}
	}

	for i, ch := range msg.buf[:] {
		if got, want := ch, byte(0); got != want {
			t.Errorf("%d: got=%v want=%v", i, got, want)
		}
	}

	if got, want := msg.used, 0; got != want {
		t.Errorf("got=%v want=%v", got, want)
	}
}

func msgEqual(m1, m2 *Message) bool {
	if string(m1.Text) != string(m2.Text) {
		return false
	}
	if len(m1.List) != len(m2.List) {
		return false
	}
	for i, v1 := range m1.List {
		v2 := m2.List[i]
		if string(v1) != string(v2) {
			return false
		}
	}
	return true
}

// String implements the Stringer interface to help with
// understanding test failures.
func (m *Message) String() string {
	var buf bytes.Buffer
	buf.Write(m.Text)
	for i := 0; i < len(m.List); i += 2 {
		key := m.List[i]
		value := m.List[i+1]
		logfmt.WriteKeyValue(&buf, key, value)
	}
	return buf.String()
}

func BenchmarkParseBytes(b *testing.B) {
	input := []byte(`message text a=1 b="value 2" c="3" d="value\n\tfour"`)
	benchmarkParseBytes(input, b)
}

func BenchmarkParseBytesNoQuotes1(b *testing.B) {
	input := []byte(`message text a=1`)
	benchmarkParseBytes(input, b)
}

func BenchmarkParseBytesNoQuotes4(b *testing.B) {
	input := []byte(`message text a=1 b=12 c=abcdef fourth=last_value`)
	benchmarkParseBytes(input, b)
}

func BenchmarkParseBytesQuotesNoEscape1(b *testing.B) {
	input := []byte(`message text a="1"`)
	benchmarkParseBytes(input, b)
}

func BenchmarkParseBytesQuotesNoEscape10(b *testing.B) {
	input := []byte(`message text a="1" b="2" c="3" d="4" e="5" f="6" g="7" h="8" i="9" j="10"`)
	benchmarkParseBytes(input, b)
}

func BenchmarkParseBytesQuotesEscape10(b *testing.B) {
	input := []byte(`message text a="\"1" b="\"2" c="\"3" d="\"4" e="\"5" f="\"6" g="\"7" h="\"8" i="\"9" j="\"10"`)
	benchmarkParseBytes(input, b)
}

func BenchmarkParseBytesQuotesEscape10Long(b *testing.B) {
	input := []byte(`message text a="\"1234567890" b="\"2345678901" c="\"3456789012" d="\"4567890123" e="\"5678901234"` +
		` f="\"6789012345" g="\"7890123456" h="\"8901234567" i="\"9012345678" j="\"0123456789"`)
	benchmarkParseBytes(input, b)
}

func benchmarkParseBytes(input []byte, b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		msg := Bytes(input)
		msg.Release()
	}
}
