package kvlog

import (
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jjeffery/kv/internal/parse"
)

var (
	// Levels is the default configuration for interpreting levels at the
	// beginning of the message text. A level is mapped to an effect, which
	// can be a color, "hide" or "none".
	Levels = map[string]string{
		"trace":   "none",
		"debug":   "none",
		"info":    "cyan",
		"warning": "yellow",
		"error":   "red",
		"alert":   "red",
		"fatal":   "red",
	}

	// Std is the 'standard' writer, which can be attached to the
	// 'standard' logger using the Attach() function.
	Std = NewWriter(os.Stderr)
)

// logEntry is a structured representation of the text emitted by a standard library logger.
//
// Explanation for why some fields are strings and some fields are byte slices.
// Byte slices point to substrings inside the input. Strings fields were populated
// without requiring memory allocation.
type logEntry struct {
	Timestamp time.Time // Time that the logger called the output's Write method
	Prefix    string    // Prefix from the logger
	Date      []byte    // Date from the logger, format YYYY/MM/DD
	Time      []byte    // Time from the logger, format HH:MM:SS[.999999]
	File      []byte    // File name and line number from the logger
	Level     string    // Message level (eg "debug")
	Effect    string    // Effect associated with level
	Text      []byte    // Message text
	List      [][]byte  // Key/value pairs
}

// Message is a structured representation of the text emitted by a standard library logger.
type Message struct {
	Timestamp time.Time // Time that the logger called the output's Write method
	Prefix    string    // Prefix from the logger
	File      string    // File name and line number from the logger
	Level     string    // Message level (eg "debug")
	Text      string    // Message text
	List      []string  // Key/value pairs
}

// Handler is the interface to implement in order to handle structured
// messages emitted by the logger.
type Handler interface {
	// Handles reports whether the handler is interested in
	// handling a message with the given prefix and level.
	// If none of the handlers want to handle a message, it is
	// not necessary to allocate the memory for the Message
	// structure.
	Handles(prefix, level string) bool

	// Handle a message written by the logger. The same message
	// is passed to all handlers on the assumption that the handlers
	// will not modify its contents.
	Handle(msg *Message)
}

// levelInfo has information about a level that is to be displayed
type levelInfo struct {
	levelb   []byte
	levelstr string
	effect   string
}

// Writer acts as the output for one or more log.Loggers. It parses each message
// written by the logger, and passes the information to any handlers registered with
// the output. The message is then formatted and printed to the output writer. If the
// output writer is a terminal, it formats the message for improved readability.
type Writer struct {
	mutex        sync.Mutex          // controls exclusive access
	printer      printer             // used for printing to the output writer
	suppress     [][]byte            // levels that should be suppressed
	suppressMap  map[string]struct{} // Levels that should be suppressed
	display      []*levelInfo        // levels that should be displayed
	levels       map[string]string   // copy of original level map
	handlers     []Handler           // list of handlers to process unsuppressed messages
	entryHandler func(*logEntry)     // for testing
}

// NewWriter creates writer that logs messages to out. If the output writer is a terminal
// device, the output will be formatted for improved readability.
func NewWriter(out io.Writer) *Writer {
	w := &Writer{
		printer: newPrinter(out),
	}
	return w
}

// Attach configures the 'standard' logger to log via this package.
// Log output will go to standard error. Use the SetOutput method to override.
func Attach() *Writer {
	Std.Attach(nil)
	return Std
}

// SetOutput sets the output destination for log messages.
func SetOutput(writer io.Writer) {
	Std.SetOutput(writer)
}

// Suppress instructs the Std writer to suppress any message with the specified level.
func Suppress(levels ...string) {
	Std.Suppress(levels...)
}

// Levels returns a list of levels and their associated actions.
func (w *Writer) Levels() map[string]string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.levels == nil {
		w.setLevels(Levels)
	}
	levels := make(map[string]string)
	for level, effect := range w.levels {
		levels[level] = effect
	}

	return levels
}

// SetLevels sets a list of levels and their associated actions.
// It replaces any existing level/effect mapping. If an unknown
// effect is supplied, a message is logged.
func (w *Writer) SetLevels(levels map[string]string) {
	w.mutex.Lock()
	w.setLevels(levels)
	w.mutex.Unlock()
}

// SetLevel sets an individual level and its associated display effect.
func (w *Writer) SetLevel(level string, effect string) {
	levels := w.Levels()
	levels[level] = effect
	w.SetLevels(levels)
}

func (w *Writer) setLevels(levels map[string]string) {
	w.suppress = nil
	w.suppressMap = make(map[string]struct{})
	w.display = nil
	w.levels = make(map[string]string)

	for level, effect := range levels {
		level := strings.TrimSpace(level)
		level = strings.TrimRight(level, ": ")
		levelb := []byte(level)
		w.levels[level] = effect
		if effect == "hide" || effect == "suppress" || effect == "ignore" {
			w.suppress = append(w.suppress, levelb)
			w.suppressMap[level] = struct{}{}
			continue
		}

		// TODO(jpj): check for known effect

		w.display = append(w.display, &levelInfo{
			levelb:   levelb,
			levelstr: level,
			effect:   effect,
		})
	}
}

// Suppress instructs the writer to suppress any message with the specified level.
func (w *Writer) Suppress(levels ...string) {
	p := w.Levels()
	for _, level := range levels {
		p[level] = "hide"
	}
	w.SetLevels(p)
}

// IsSuppressed reports true if level should be suppressed.
func (w *Writer) IsSuppressed(level string) bool {
	_, ok := w.suppressMap[level]
	return ok
}

// Handle registers a handler that will be called for every logging
// event. The function receives a message, which is a structured
// representation of the log message text.
//
// This function is useful for registering handlers that send log
// information to external sources.
func (w *Writer) Handle(h Handler) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if h != nil {
		w.handlers = append(w.handlers, h)
	}
}

// Attach sets this writer as the output destination
// for the specified logger. If the logger is not specified,
// then this writer attaches to the log package 'standard' logger.
//
// This method calls SetOutput for the specified logger
// (or the standard logger) to set its output writer.
func (w *Writer) Attach(logger ...*log.Logger) {
	if len(logger) == 0 {
		logger = []*log.Logger{nil}
	}
	for _, l := range logger {
		lw := newLogWriter(w, l)
		if l == nil {
			log.SetOutput(lw)
		} else {
			l.SetOutput(lw)
		}
	}
}

// SetOutput sets the output destination for log messages.
func (w *Writer) SetOutput(out io.Writer) {
	w.mutex.Lock()
	w.printer = newPrinter(out)
	w.mutex.Unlock()
}

func (w *Writer) shouldSuppress(msg []byte) bool {
	for _, levelb := range w.suppress {
		if bytes.HasPrefix(msg, levelb) {
			if colonRE.Match(msg[len(levelb):]) {
				return true
			}
		}
	}
	return false
}

func (w *Writer) getLevel(msg []byte) (level string, effect string, skip int) {
	for _, levelInfo := range w.display {
		if len(msg) < len(levelInfo.levelb)+1 {
			continue
		}
		if bytes.EqualFold(msg[:len(levelInfo.levelb)], levelInfo.levelb) {
			// match the level but we need a following colon for a match
			suffix := colonRE.Find(msg[len(levelInfo.levelb):])
			if suffix != nil {
				level = levelInfo.levelstr
				effect = levelInfo.effect
				skip = len(levelInfo.levelb) + len(suffix)
				break
			}
		}
	}
	return level, effect, skip
}

func (w *Writer) handler(entry *logEntry) {
	if w.entryHandler != nil {
		w.entryHandler(entry)
	}
	if w.handlers != nil {
		var msg *Message
		for _, h := range w.handlers {
			if h.Handles(entry.Prefix, entry.Level) {
				if msg == nil {
					msg = &Message{
						Timestamp: entry.Timestamp,
						Prefix:    entry.Prefix,
						Level:     entry.Level,
						Text:      string(entry.Text),
					}
					if entry.File != nil {
						msg.File = string(entry.File)
					}
					if len(entry.List) > 0 {
						msg.List = make([]string, len(entry.List))
						for i, v := range entry.List {
							msg.List[i] = string(v)
						}
					}
				}
				h.Handle(msg)
			}
		}
	}
	w.printer.Print(entry)
}

// logWriter is a writer tailored for a specific logger.
type logWriter struct {
	prefixb []byte         // logger prefix bytes
	prefixs string         // logger prefix string
	utc     bool           // is time in UTC
	dateRE  *regexp.Regexp // regexp for extracting date
	timeRE  *regexp.Regexp // regexp for extracting time
	fileRE  *regexp.Regexp // regexp for extracting file (???:0 D:/go/src/github.com/jjeffery/kv/kv.go:123)
	output  *Writer
	logger  *log.Logger
	changed bool
}

var (
	dateRE  = regexp.MustCompile(`^\d{4}/\d\d/\d\d`)
	timeRE  = regexp.MustCompile(`^\d\d:\d\d:\d\d(\.\d+)?`)
	fileRE  = regexp.MustCompile(`^([a-zA-Z]:)?[^:]+:\d+`)
	colonRE = regexp.MustCompile(`^\s*:\s*`)
)

func newLogWriter(output *Writer, logger *log.Logger) *logWriter {
	w := &logWriter{
		output: output,
		logger: logger,
	}
	w.setup()
	return w
}

func isspace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func (w *logWriter) setup() {
	var (
		prefix string
		flags  int
	)
	if w.logger == nil {
		prefix = log.Prefix()
		flags = log.Flags()
	} else {
		prefix = w.logger.Prefix()
		flags = w.logger.Flags()
	}
	if prefix != "" {
		w.prefixs = prefix
		w.prefixb = []byte(prefix)
	}
	if flags&log.Ldate != 0 {
		w.dateRE = dateRE
	}
	if flags&(log.Ltime|log.Lmicroseconds) != 0 {
		w.timeRE = timeRE
	}
	if flags&(log.Lshortfile|log.Llongfile) != 0 {
		w.fileRE = fileRE
	}
	if flags&log.LUTC != 0 {
		w.utc = true
	}
}

// Write implements the io.Writer interface. This method is
// called from the logger. Because the logger's mutex is
// locked, and because we want to read the logger's prefix and
// its flags, we copy the details and let the actual logging
// be done by a goroutine.
func (w *logWriter) Write(p []byte) (n int, err error) {
	var (
		now     = time.Now() // do this early
		prefix  string
		logdate []byte
		logtime []byte
		file    []byte
		changed bool
	)

	if w.utc {
		now = now.UTC()
	}
	if w.prefixb != nil {
		if bytes.HasPrefix(p, w.prefixb) {
			prefix = w.prefixs
			p = p[len(prefix):]
		} else {
			changed = true
		}
	}
	p = bytes.TrimLeftFunc(p, isspace)
	if w.dateRE != nil {
		dateb := w.dateRE.Find(p)
		if dateb != nil {
			logdate = dateb
			p = p[len(dateb):]
			p = bytes.TrimLeftFunc(p, isspace)
		} else {
			changed = true
		}
	}
	if w.timeRE != nil {
		timeb := w.timeRE.Find(p)
		if timeb != nil {
			logtime = timeb
			p = p[len(timeb):]
			p = bytes.TrimLeftFunc(p, isspace)
		} else {
			changed = true
		}
	}
	if w.fileRE != nil {
		fileb := w.fileRE.Find(p)
		if fileb != nil {
			file = fileb
			p = p[len(fileb):]
			p = bytes.TrimLeftFunc(p, isspace)
			skip := colonRE.Find(p)
			p = p[len(skip):]
		} else {
			changed = true
		}
	}

	w.output.mutex.Lock()
	if w.output.levels == nil {
		// apply the default levels as late as possible,
		// which gives the calling program an opportunity
		// to change default levels at program initialization
		w.output.setLevels(Levels)
	}
	if !w.output.shouldSuppress(p) {
		level, effect, skip := w.output.getLevel(p)
		p = p[skip:]
		msg := parse.Bytes(p)
		ent := logEntry{
			Timestamp: now,
			Prefix:    prefix,
			Date:      logdate,
			Time:      logtime,
			File:      file,
			Level:     level,
			Effect:    effect,
			Text:      msg.Text,
			List:      msg.List,
		}
		w.output.handler(&ent)
		msg.Release()
	}
	w.output.mutex.Unlock()

	if changed && !w.changed {
		w.changed = true
		go func() {
			w.output.Attach(w.logger)
			if w.logger == nil {
				log.Println("warning: logger details changed after kvlog.Attach")
			} else {
				w.logger.Println("warning: logger details after kvlog.Writer.Attach")
			}
		}()
	}

	return len(p), nil
}
