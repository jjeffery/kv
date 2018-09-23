package kv

import (
	"bytes"
	"context"
	"strconv"
)

// A Message consists of message text followed by
// zero or more key/value pairs, optionally followed by
// another message.
type Message struct {
	List        List
	ContextList List
	Text        string
	Next        *Message
}

// Ctx returns a message populated with key/values from the context.
//
// See the example for NewContext.
func Ctx(ctx context.Context) Message {
	return Message{
		ContextList: fromContext(ctx),
	}
}

// Msg returns a message with text.
func Msg(text string) Message {
	return Message{
		Text: text,
	}
}

// Parse parses the input to produce a message.
func Parse(input []byte) Message {
	lex := newLexer(input)
	msg := newMessage(lex)
	if msg == nil {
		return Message{}
	}
	return *msg
}

// With returns a list with keyvals as contents.
func With(keyvals ...interface{}) Message {
	keyvals = Flatten(keyvals)
	return Message{
		List: keyvals,
	}
}

// Ctx returns a message populated with key/values from the context.
func (msg Message) Ctx(ctx context.Context) Message {
	msg.ContextList = fromContext(ctx)
	return msg
}

// Msg returns a message with text.
func (msg Message) Msg(text string) Message {
	msg.Text = text
	return msg
}

// With returns a message with the keyvals appended.
func (msg Message) With(keyvals ...interface{}) Message {
	msg.List = msg.List.With(keyvals...)
	return msg
}

// String returns a string representation of the message in
// the format format: "text key1=value1 key2=value2  ...".
func (msg Message) String() string {
	var buf bytes.Buffer
	msg.writeToBuffer(&buf)
	return buf.String()
}

// MarshalText implements the TextMarshaler interface.
func (msg Message) MarshalText() (text []byte, err error) {
	var buf bytes.Buffer
	msg.writeToBuffer(&buf)
	return buf.Bytes(), nil
}

func (msg *Message) writeToBuffer(buf *bytes.Buffer) {
	if msg == nil {
		return
	}
	if buf.Len() > 0 && len(msg.Text) > 0 {
		buf.WriteRune(' ')
	}
	buf.WriteString(msg.Text)
	msg.List.writeToBuffer(buf)
	msg.ContextList.writeToBuffer(buf)
	msg.Next.writeToBuffer(buf)
}

func newMessage(lex *lexer) *Message {
	var message Message
	var text []byte

	lex.skipWS()
	for lex.notMatch(tokKey, tokQuotedKey, tokEOF) {
		text = append(text, lex.lexeme...)
		lex.next()
	}

	if len(text) == 0 && lex.token == tokEOF {
		// nothing to read
		return nil
	}

	message.Text = string(bytes.TrimSpace(text))

	for lex.match(tokKey, tokQuotedKey) {
		if lex.token == tokKey {
			message.List = append(message.List, string(lex.lexeme))
		} else {
			message.List = append(message.List, unquote(lex.lexeme))
		}
		// move past equals
		lex.next()
		if lex.token == tokEquals {
			lex.next()
		}

		switch lex.token {
		case tokQuoted:
			message.List = append(message.List, unquote(lex.lexeme))
		case tokWS:
			message.List = append(message.List, "")
		default:
			message.List = append(message.List, string(lex.lexeme))
		}

		lex.next()
		lex.skipWS()
	}

	message.Next = newMessage(lex)
	return &message
}

func unquote(input []byte) string {
	s, _ := strconv.Unquote(string(input))
	return s
}
