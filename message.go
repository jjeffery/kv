package kv

import (
	"bytes"
	"context"
	"log"
	"strconv"
	"strings"
)

var (
	// LogFunc is called by the Message.Log method to log a message.
	// By default it uses the Go standard library log package.
	LogFunc = log.Println
)

// A Message consists of message text followed by
// zero or more key/value pairs.
type Message struct {
	Text        string // message text
	List        List   // key value pairs
	ContextList List   // key value pairs from context
}

// From returns a message populated with key/values from the context.
func From(ctx context.Context) *Message {
	return &Message{
		ContextList: fromContext(ctx),
	}
}

// Msg returns a message with text.
func Msg(text string) *Message {
	return &Message{
		Text: text,
	}
}

// Parse parses the input to produce a message.
func Parse(input []byte) *Message {
	input = bytes.TrimSpace(input)
	lex := newLexer(input)
	msg := newMessage(lex)
	return msg
}

func (msg *Message) clone() *Message {
	m := *msg
	return &m
}

// From returns a new message based on msg, but populated with key/value pairs from the context.
func (msg *Message) From(ctx context.Context) *Message {
	msg = msg.clone()
	msg.ContextList = fromContext(ctx)
	return msg
}

// Msg returns a new message based on msg, but with text as its message text.
func (msg *Message) Msg(text string) *Message {
	msg = msg.clone()
	msg.Text = text
	return msg
}

// With returns a new message based on msg, but with the keyvals appended to its
// list of key/value pairs.
func (msg *Message) With(keyvals ...interface{}) *Message {
	msg = msg.clone()
	msg.List = msg.List.With(keyvals...)
	return msg
}

// Wrap returns an error based on the message that wraps err.
//
// This method can be useful when creating an error with
// key/values from the context. See the example.
func (msg *Message) Wrap(err error, text ...string) *Error {
	e := &Error{
		Text:        msg.Text,
		List:        msg.List,
		ContextList: msg.ContextList,
		Err:         err,
	}
	if len(text) > 0 {
		e.Text = strings.Join(text, " ")
	}
	return e
}

// Err returns an error with the specified text and
// key/value pairs copied from the message.
//
// This method can be useful when creating an error with
// key/values from the context. See the example.
func (msg *Message) Err(text string) *Error {
	return &Error{
		Text:        text,
		List:        msg.List,
		ContextList: msg.ContextList,
	}
}

// String returns a string representation of the message in
// the format format: "text key1=value1 key2=value2  ...".
func (msg *Message) String() string {
	var buf bytes.Buffer
	msg.writeToBuffer(&buf)
	return buf.String()
}

// MarshalText implements the TextMarshaler interface.
func (msg *Message) MarshalText() (text []byte, err error) {
	var buf bytes.Buffer
	msg.writeToBuffer(&buf)
	return buf.Bytes(), nil
}

// UnmarshalText implements the TextUnmarshaler interface.
func (msg *Message) UnmarshalText(text []byte) error {
	m := Parse(text)
	*msg = *m
	return nil
}

// Log the message using the LogFunc function. By default this
// uses the Go standard library log package.
func (msg *Message) Log() {
	LogFunc(msg)
}

func (msg *Message) writeToBuffer(buf *bytes.Buffer) {
	if msg == nil {
		return
	}
	if msg.Text == "" && len(msg.List) == 0 && len(msg.ContextList) == 0 {
		return
	}
	if msg.Text != "" {
		if buf.Len() > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(msg.Text)
	}
	if len(msg.ContextList) == 0 {
		// only the message list
		msg.List.dedup().writeToBuffer(buf)
	} else if len(msg.List) == 0 {
		msg.ContextList.dedup().writeToBuffer(buf)
	} else {
		var list List
		list = list.With(msg.List...)
		list = list.With(msg.ContextList...)
		list.dedup().writeToBuffer(buf)
	}
}

func newMessage(lex *lexer) *Message {
	// firstKeyPos is the position of the first key in the message
	//
	// consider the following example message:
	//
	//  this is a message key=1 key=2 more message stuff key=3
	//                                                   ^
	// if a message has key=val and then text that       |
	// does not match key=val, then the key=val is       |
	// not parsed for example, the first key is here ----+
	var firstKeyPos int

	// count kv pairs so that we can allocate once only
	var kvCount int

	// iterate through the message looking for the position
	// before which we will not be looking for key/val pairs
	for lex.token != tokEOF {
		for lex.notMatch(tokKey, tokQuotedKey, tokEOF) {
			firstKeyPos = 0
			lex.next()
		}
		if lex.token == tokEOF {
			break
		}
		firstKeyPos = lex.pos
		for lex.match(tokKey, tokQuotedKey) {
			kvCount += 2
			lex.next() // skip past key
			lex.next() // skip past value
			lex.skipWS()
		}
	}

	lex.rewind()
	lex.skipWS()
	var (
		text    []byte
		message Message
	)

	if firstKeyPos == 0 {
		// there are no key/value pairs
		text = lex.input
	} else {
		message.List = make(List, 0, kvCount)
		var pos int
		for lex.pos < firstKeyPos {
			pos = lex.pos
			lex.next()
		}
		text = lex.input[:pos]
		for lex.match(tokKey, tokQuotedKey) {
			if lex.token == tokKey {
				message.List = append(message.List, string(lex.lexeme()))
			} else {
				message.List = append(message.List, unquote(lex.lexeme()))
			}
			lex.next()

			switch lex.token {
			case tokQuoted:
				message.List = append(message.List, unquote(lex.lexeme()))
			default:
				message.List = append(message.List, string(lex.lexeme()))
			}

			lex.next()
			lex.skipWS()
		}
	}

	message.Text = string(bytes.TrimSpace(text))
	return &message
}

func unquote(input []byte) string {
	s, err := strconv.Unquote(string(input))
	if err != nil {
		return "?"
	}
	return s
}
