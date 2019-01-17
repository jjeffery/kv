package kvlog

import (
	"bytes"
	"io"
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/jjeffery/kv/internal/logfmt"
	"github.com/jjeffery/kv/internal/pool"
	"github.com/jjeffery/kv/internal/terminal"
)

const (
	// defaultTerminalWidth is the width to use for a terminal if
	// the attempt to query the terminal width fails.
	defaultTerminalWidth = 120
)

var (
	whiteSpaceRE = regexp.MustCompile(`^\s+`)
	blackSpaceRE = regexp.MustCompile(`^[^\s,]+`)
)

type printer interface {
	Print(*logEntry)
}

func newPrinter(w io.Writer) printer {
	if fd, ok := fileDescriptor(w); ok {
		if terminal.IsTerminal(fd) {
			terminal.EnableVirtualTerminalProcessing(fd)
			return &terminalPrinter{
				w: w,
				width: func() int {
					width, _, err := terminal.GetSize(fd)
					if err != nil {
						return defaultTerminalWidth
					}
					return width
				},
			}
		}
	}

	return &simplePrinter{w: w}
}

// simplePrinter prints to a non-terminal device
type simplePrinter struct {
	w io.Writer
}

func (p *simplePrinter) Print(msg *logEntry) {
	buf := pool.AllocBuffer()
	if len(msg.Prefix) > 0 {
		buf.WriteString(msg.Prefix)
		// no space here to match the way prefixes
		// work in the log package
	}
	if len(msg.Date) > 0 {
		buf.Write(msg.Date)
		buf.WriteRune(' ')
	}
	if len(msg.Time) > 0 {
		buf.Write(msg.Time)
		buf.WriteRune(' ')
	}
	if len(msg.File) > 0 {
		buf.Write(msg.File)
		buf.WriteString(": ")
	}
	if msg.Level != "" {
		buf.WriteString(msg.Level)
		buf.WriteString(": ")
	}
	buf.Write(msg.Text)
	for i := 0; i < len(msg.List); i += 2 {
		buf.WriteRune(' ')
		logfmt.WriteKeyValue(buf, msg.List[i], msg.List[i+1])
	}
	buf.WriteRune('\n')
	p.w.Write(buf.Bytes())
	pool.ReleaseBuffer(buf)
}

// terminalPrinter is used to write log messages to an ANSI terminal.
type terminalPrinter struct {
	w       io.Writer
	width   func() int
	nocolor bool

	buf    *bytes.Buffer
	indent int
	col    int  // current column number
	infmt  bool // inside a format
}

func (p *terminalPrinter) reset() {
	pool.ReleaseBuffer(p.buf)
	p.buf = nil
	p.indent = 0
	p.col = 0
}

func (p *terminalPrinter) writeString(s string) {
	p.buf.WriteString(s)
	p.col += utf8.RuneCountInString(s)
}

func (p *terminalPrinter) writeRune(r rune) {
	p.buf.WriteRune(r)
	p.col++
}

func (p *terminalPrinter) write(b []byte) {
	p.buf.Write(b)
	p.col += utf8.RuneCount(b)
}

var ansiRE = regexp.MustCompile(`^[0-9]+(;[0-9]+)*$`)

func (p *terminalPrinter) startFormat(effect string) {
	if p.nocolor {
		return
	}
	if color, ok := colorEffects[effect]; ok {
		p.buf.WriteString("\x1b[0;")
		p.buf.WriteString(color)
		p.buf.WriteRune('m')
		p.infmt = true
	} else if ansiRE.MatchString(effect) {
		p.buf.WriteString("\x1b[0;")
		p.buf.WriteString(effect)
		p.buf.WriteRune('m')
		p.infmt = true
	}
}

func (p *terminalPrinter) newline() {
	p.buf.WriteRune('\n')
	for i := 0; i < p.indent; i++ {
		p.buf.WriteRune(' ')
	}
	p.col = p.indent
}

func (p *terminalPrinter) resetFormat() {
	if p.infmt {
		p.buf.WriteString("\x1b[0m")
		p.infmt = false
	}
}

func (p *terminalPrinter) Print(msg *logEntry) {
	p.buf = pool.AllocBuffer()

	if len(msg.Prefix) > 0 {
		p.writeString(msg.Prefix)
		// no space here to match the way prefixes
		// work in the log package
	}
	if len(msg.Date) > 0 {
		p.write(msg.Date)
		p.writeRune(' ')
	}
	if len(msg.Time) > 0 {
		p.write(msg.Time)
		p.writeRune(' ')
	}

	// indent is the hanging indent for messages that span multiple lines
	p.indent = p.col
	if p.indent == 0 {
		p.indent = 4
	}

	if len(msg.File) > 0 {
		p.startFormat("bright black")
		p.write(msg.File)
		p.writeString(": ")
		p.resetFormat()
	}

	if msg.Level != "" {
		p.startFormat(msg.Effect)
		p.writeString(msg.Level)
		p.writeString(": ")
		p.resetFormat()
	}

	// print to one less than the terminal width because some terminals
	// don't format nicely otherwise (eg git bash)
	width := p.width() - 1
	if width <= 0 {
		width = defaultTerminalWidth
	}

	// print message text with line wrapping
	for in := msg.Text; len(in) > 0; {
		var (
			wsLen, bsLen, punctLen int
			punct                  rune
		)
		ws := whiteSpaceRE.Find(in)
		if n := len(ws); n > 0 {
			in = in[n:]
			wsLen = 1
		}
		bs := blackSpaceRE.Find(in)
		if len(bs) > 0 {
			in = in[len(bs):]
			bsLen = utf8.RuneCount(bs)
		}

		// The black space RE will terminate before punctuation to handle very long
		// strings with no spaces but possibly punctuation. Detect if it has terminated
		// before punctuation, and if so include the punctuation char on the same line.
		if len(in) > 0 {
			var size int
			punct, size = utf8.DecodeRune(in)
			if !unicode.IsSpace(punct) {
				punctLen = size
				in = in[size:]
			}
		}

		if bsLen+punctLen == 0 {
			// trailing white space
			continue
		}

		if bsLen+wsLen+punctLen+p.col > width {
			p.newline()
			p.write(bs)
		} else {
			if len(ws) > 0 {
				p.writeRune(' ')
			}
			p.write(bs)
		}
		if punctLen > 0 {
			p.writeRune(punct)
		}
	}

	// print key/value pairs with line wrapping
	for i := 0; i < len(msg.List); i += 2 {
		key := msg.List[i]
		val := msg.List[i+1]
		keyLen := utf8.RuneCount(key)
		valLen := utf8.RuneCount(val)
		const equalsLen = 1
		var wsLen int
		if p.col > p.indent {
			wsLen = 1
		}
		if keyLen+valLen+equalsLen+wsLen+p.col > width {
			p.newline()
			wsLen = 0
		}
		if wsLen > 0 {
			p.writeRune(' ')
		}
		p.write(key)
		p.writeRune('=')
		p.startFormat("bright cyan")
		p.write(val)
		p.resetFormat()
	}

	p.writeRune('\n')
	p.w.Write(p.buf.Bytes())
	p.reset()
}

var colorEffects = map[string]string{
	"black":          "30",
	"red":            "31",
	"green":          "32",
	"yellow":         "33",
	"blue":           "34",
	"magenta":        "35",
	"cyan":           "36",
	"white":          "37",
	"gray":           "1;30", // same as bright black
	"grey":           "1;30",
	"bright black":   "90",
	"bright red":     "91",
	"bright green":   "92",
	"bright yellow":  "93",
	"bright blue":    "94",
	"bright magenta": "95",
	"bright cyan":    "96",
	"bright white":   "37;1",
}

// fileDescriptor returns the file descriptor associated with the
// writer, or (0, false) if no file descriptor is available.
func fileDescriptor(w io.Writer) (fd int, ok bool) {
	file, ok := w.(interface{ Fd() uintptr })
	if ok {
		fd = int(file.Fd())
	}
	return fd, ok
}
