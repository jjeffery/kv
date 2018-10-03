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
	lex.token = tokQuoted
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
	if err == nil {
		switch ch {
		case ':':
			// remove any ':' separator after a quoted value
			break
		case '=':
			// an equals at the end of a quoted value means treat
			// it as a keyword
			lex.token = tokQuotedKey
		default:
			lex.reader.UnreadRune()
		}
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
			// Only consider this a keyword if the next character
			// after the equals is a non-space character. This picks
			// up cases where, for example, a base64 value is logged
			// that has one or more '=' chars at the end.
			ch, _, err = lex.reader.ReadRune()
			if err != nil {
				// eof, so the equals is just part of the word
				buf.WriteRune('=')
				break
			}
			lex.reader.UnreadRune()
			if unicode.IsSpace(ch) {
				// equals is part of the word
				buf.WriteRune('=')
			} else {
				// next char is non-space, so we consider
				// this to be a keyword
				token = tokKey
			}
			break
		}
		if lex.token == tokKey || lex.token == tokQuotedKey {
			// unquoted colon terminates a value
			if ch == ':' {
				break
			}
		}
		buf.WriteRune(ch)
	}
	lex.token = token
	lex.lexeme = buf.Bytes()
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
