// This needs significant optimization, mainly to reduce memory
// allocation.

// Package kvlog provides a writer intended for use with the
// Go standard library log package.
package kvlog

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/jjeffery/kv"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// headerRE matches the possible combinations of date and/or time that the
	// stdlib log package will produce. No attempt is made to match the file/line
	// but it could be added.
	headerRE = regexp.MustCompile(`^(\d{4}/\d\d/\d\d )?(\d\d:\d\d:\d\d(\.\d{0,6})? )?`)

	// newline is the newline sequence. Originally changed if windows, but seems
	// to work fine as the same value for all operating systems.
	// TODO: change to a const.
	newline = "\n"

	// verbosePrefixes is a list of prefixes that indicate the message should only
	// be displayed in verbose mode
	verbosePrefixes = []string{
		"debug:",
		"trace:",
	}

	whiteSpaceRE = regexp.MustCompile(`^\s+`)
	blackSpaceRE = regexp.MustCompile(`^[^\s,]+`)
)

// Writer implements io.Writer and can be used as the writer for
// log.SetOutput.
type Writer struct {
	Out     io.Writer
	Width   func() int
	Verbose bool
	mutex   sync.Mutex
}

// NewWriter returns a writer that can be used as a writer for the log.
func NewWriter(writer io.Writer) *Writer {
	w := &Writer{
		Out: writer,
	}
	const defaultWidth = 120
	if fder, ok := writer.(interface{ Fd() uintptr }); ok {
		fd := int(fder.Fd())

		if terminal.IsTerminal(fd) {
			w.Width = func() int {
				width, _, err := terminal.GetSize(fd)
				if err != nil {
					return defaultWidth
				}
				return width
			}
		}
	}
	if w.Width == nil {
		w.Width = func() int {
			return defaultWidth
		}
	}
	return w
}

// messageText contains the message parts and rune count required to
// display each part.
type messageText struct {
	text           string
	textRuneCount  int
	totalRuneCount int
	keyvals        []keyvalText
}

// keyvalText contains the key/value text and the rune count required
// to display.
type keyvalText struct {
	text          string
	textRuneCount int
}

func (w *Writer) Write(p []byte) (int, error) {
	msg := kv.Parse(p)

	var prefix string
	{
		match := headerRE.FindStringSubmatch(msg.Text)
		if len(match) > 0 {
			prefix = match[0]
			msg.Text = msg.Text[len(prefix):]
		}
	}
	if !w.Verbose {
		for _, vp := range verbosePrefixes {
			if strings.HasPrefix(msg.Text, vp) {
				// suppress verbose messages
				return 0, nil
			}
		}
	}
	width := w.Width()
	var indent string
	if len(prefix) > 0 {
		indent = strings.Repeat(" ", len(prefix))
	} else {
		indent = "    "
	}

	var msgTexts []messageText

	for {
		msgText := messageText{
			text:          msg.Text,
			textRuneCount: utf8.RuneCountInString(msg.Text),
		}
		// totalRuneCount is the number of columns required to display the entire message on one line
		msgText.totalRuneCount = msgText.textRuneCount

		for i := 0; i < len(msg.List); i += 2 {
			keyval := kv.P(msg.List[i].(string), msg.List[i+1]).String()
			runeCount := utf8.RuneCountInString(keyval)
			msgText.keyvals = append(msgText.keyvals, keyvalText{
				text:          keyval,
				textRuneCount: runeCount,
			})
			// the totalRuneCount includes a "+1" for a space because
			// it is used to determine if the message will fit all by
			// itself on one line.
			msgText.totalRuneCount += runeCount + 1
		}

		msgTexts = append(msgTexts, msgText)
		if msg.Next == nil {
			break
		}

		msg = *msg.Next
	}

	var sb bytes.Buffer
	sb.WriteString(prefix)
	col := len(prefix)
	var needSpace int
	var needColon int

	for i, msgText := range msgTexts {
		if i > 0 {
			if col+msgText.totalRuneCount+needSpace+needColon > width {
				// won't fit the message on the rest of the line, so start a new one
				sb.WriteString(newline)
				sb.WriteString(indent)
				col = len(indent)
				needSpace = 0
				needColon = 0
			}
			if needColon > 0 {
				sb.WriteRune(':')
				needColon = 0
			}
		}
		// TODO: check if the message text will not fit on one line, and if that is the case display the
		// text with line-wrapping (which will be slower).
		if msgText.textRuneCount > width {
			// this is where the message itself is too long to fit on one line, so we need to
			// line wrap
			in := []byte(msgText.text)
			for len(in) > 0 {
				ws := whiteSpaceRE.Find(in)
				if len(ws) > 0 {
					in = in[len(ws):]
				}
				var wsLen int
				if len(ws) > 0 {
					wsLen = 1
				}
				bs := blackSpaceRE.Find(in)
				if len(bs) > 0 {
					in = in[len(bs):]
				}
				bsLen := utf8.RuneCount(bs)

				// The black space RE will terminate before punctuation to handle very long
				// strings with no spaces but possibly punctuation. Detect if it has terminated
				// before punctuation, and if so include the punctuation char on the same line.
				var (
					punct    rune
					punctLen int
				)
				if len(in) > 0 {
					var size int
					punct, size = utf8.DecodeRune(in)
					if !unicode.IsSpace(punct) {
						punctLen = size
						in = in[size:]
					}
				}

				if bsLen+wsLen+punctLen+col > width {
					sb.WriteString(newline)
					sb.WriteString(indent)
					sb.Write(bs)
					col = len(indent) + bsLen
				} else {
					if len(ws) > 0 {
						sb.WriteRune(' ')
						col++
					}
					sb.Write(bs)
					col += bsLen
				}
				if punctLen > 0 {
					sb.WriteRune(punct)
					col += punctLen
				}
			}
			needSpace = 1
		} else if msgText.textRuneCount > 0 {
			if needSpace > 0 {
				sb.WriteRune(' ')
				col++
				needSpace = 0
			}
			sb.WriteString(msgText.text)
			col += msgText.textRuneCount
			needSpace = 1
		}

		for _, keyvalText := range msgText.keyvals {
			if keyvalText.textRuneCount+needSpace+col > width {
				sb.WriteString(newline)
				sb.WriteString(indent)
				col = len(indent)
				needSpace = 0
			}
			if needSpace > 0 {
				sb.WriteRune(' ')
				needSpace = 0
				col++
			}
			sb.WriteString(keyvalText.text)
			col += keyvalText.textRuneCount
			needSpace = 1
		}
		needColon = 1
	}
	sb.WriteString(newline)

	w.mutex.Lock()
	n, err := w.Out.Write(sb.Bytes())
	w.mutex.Unlock()
	return n, err
}
