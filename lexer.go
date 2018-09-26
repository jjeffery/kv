package kv

import (
	"bytes"
	"io"
	"unicode"
)

const (
	tokEOF = iota
	tokWord
	tokWS
	tokEquals
	tokKey
	tokQuoted
	tokQuotedKey
)

type lexer struct {
	reader *bytes.Reader
	token  int
	lexeme []byte
}

func newLexer(input []byte) *lexer {
	lex := &lexer{
		reader: bytes.NewReader(input),
	}
	lex.next()
	return lex
}

func (lex *lexer) next() bool {
	ch, _, err := lex.reader.ReadRune()
	if err == io.EOF {
		return lex.eof()
	}
	if unicode.IsSpace(ch) {
		return lex.whiteSpace(ch)
	}
	if ch == '"' {
		return lex.quoted(ch)
	}
	if ch == '=' {
		return lex.equals(ch)
	}

	return lex.word(ch)
}

func (lex *lexer) eof() bool {
	lex.token = tokEOF
	lex.lexeme = nil
	return false
}

func (lex *lexer) whiteSpace(ch rune) bool {
	var buf bytes.Buffer
	buf.WriteRune(ch)
	for {
		var err error
		ch, _, err = lex.reader.ReadRune()
		if err != nil {
			break
		}
		if !unicode.IsSpace(ch) {
			lex.reader.UnreadRune()
			break
		}
		buf.WriteRune(ch)
	}
	lex.token = tokWS
	lex.lexeme = buf.Bytes()
	return true
}

func (lex *lexer) quoted(quote rune) bool {
	var buf bytes.Buffer
	var escaped bool
	buf.WriteRune(quote)
	for {
		ch, _, err := lex.reader.ReadRune()
		if err != nil {
			// premature end of input
			buf.WriteRune(quote)
			break
		}
		buf.WriteRune(ch)
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == quote {
			break
		}
	}

	// lose any ":" separator after a quoted value
	ch, _, err := lex.reader.ReadRune()
	if err == nil && ch != ':' {
		lex.reader.UnreadRune()
	}
	if ch == '=' {
		lex.token = tokQuotedKey
	} else {
		lex.token = tokQuoted
	}
	lex.lexeme = buf.Bytes()
	return true
}

func (lex *lexer) word(ch rune) bool {
	var buf bytes.Buffer
	buf.WriteRune(ch)
	token := tokWord
	for {
		var err error
		ch, _, err = lex.reader.ReadRune()
		if err != nil {
			break
		}
		if unicode.IsSpace(ch) {
			lex.reader.UnreadRune()
			break
		}
		if ch == '=' {
			lex.reader.UnreadRune()
			token = tokKey
			break
		}
		// <HACK>
		if lex.token == tokEquals {
			// unquoted colon terminates a value
			if ch == ':' {
				break
			}
		}
		// </HACK>
		buf.WriteRune(ch)
	}
	lex.token = token
	lex.lexeme = buf.Bytes()
	return true
}

func (lex *lexer) equals(ch rune) bool {
	lex.lexeme = []byte{byte(ch)}
	lex.token = tokEquals
	return true
}

func (lex *lexer) skipWS() {
	for lex.token == tokWS {
		if !lex.next() {
			return
		}
	}
}

func (lex *lexer) notMatch(toks ...int) bool {
	for _, tok := range toks {
		if tok == lex.token {
			return false
		}
	}
	return true
}

func (lex *lexer) match(toks ...int) bool {
	for _, tok := range toks {
		if tok == lex.token {
			return true
		}
	}
	return false
}
