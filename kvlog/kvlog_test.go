package kvlog

import (
	"bytes"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"testing"

	"github.com/jjeffery/kv"
)

func TestWriter(t *testing.T) {
	tests := []struct {
		input     string
		output    string
		flags     int
		prefix    string
		width     int
		showColor bool
		verbose   bool
	}{
		{
			input: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			output: "2099/12/31 12:34:56 this is the message key1=value1\n" +
				"                    key2=value2 key3=value3\n",
			width: 60,
			flags: log.LstdFlags,
		},
		{
			input: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			output: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2\n" +
				"                    key3=value3\n",
			width: 70,
			flags: log.LstdFlags,
		},
		{
			input: "prog [400] 2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			output: "prog [400] 2099/12/31 12:34:56 this is the message key1=value1\n" +
				"                               key2=value2 key3=value3\n",
			width:  70,
			prefix: "prog [400] ",
			flags:  log.LstdFlags,
		},
		{
			input:  "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			output: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			width:  80,
			flags:  log.LstdFlags,
		},
		{
			input: "this is the message key1=value1 key2=value2 key3=value3\n",
			output: "this is the message key1=value1\n" +
				"    key2=value2 key3=value3\n",
			width: 40,
		},
		{
			input:  "12:34:56 error: this is the message key1=value1 key2=value2: file not found\n",
			output: "12:34:56 error: this is the message key1=value1 key2=value2: file not found\n",
			flags:  log.Ltime,
		},
		{
			input: "12:34:56 error: this is the message key1=value1 key2=value2: file not found\n",
			output: "12:34:56 error: this is the message key1=value1 key2=value2: file not\n" +
				"         found\n",
			width: 70,
			flags: log.Ltime,
		},
		{
			input: "12:34:56 error: this is a very long message that will wrap over the line key1=value1 key2=value2: file not found\n",
			output: "12:34:56 error: this is a very long message that will wrap\n" +
				"         over the line key1=value1 key2=value2: file not\n" +
				"         found\n",
			width: 60,
			flags: log.Ltime,
		},
		{
			input: `11:17:29 select "id","version","created_at","updated_at","status","message","nick_name","user_id",` +
				`"customer_type","customer_company_name","customer_trading_name","customer_abn",` +
				`"customer_acn","customer_phone_number","customer_fax_number","customer_first_name",` +
				`"customer_last_name","customer_middle_names","customer_mobile_number",` +
				`"customer_date_of_birth" from customers where id > $1 order by id limit $2 [14 10]`,
			output: `11:17:29 select "id","version","created_at","updated_at","status","message","nick_name","user_id",` + "\n" +
				`         "customer_type","customer_company_name","customer_trading_name","customer_abn",` + "\n" +
				`         "customer_acn","customer_phone_number","customer_fax_number","customer_first_name",` + "\n" +
				`         "customer_last_name","customer_middle_names","customer_mobile_number",` + "\n" +
				`         "customer_date_of_birth" from customers where id > $1 order by id limit $2 [14 10]` + "\n",
			width: 100,
			flags: log.Ltime,
		},
		{ // colors
			input:     "prefix: 11:17:39 error: this is an error message\n",
			output:    "prefix: 11:17:39 \x1b[0;31merror: \x1b[0mthis is an error message\n",
			width:     120,
			showColor: true,
			prefix:    "prefix: ",
			flags:     log.Ltime,
		},
		{ // colors
			input:     "11:17:39  Error: this is an error message\n",
			output:    "11:17:39 \x1b[0;31merror: \x1b[0mthis is an error message\n",
			width:     120,
			showColor: true,
			flags:     log.Ltime,
		},
		{ // colors
			input:     "11:17:39  custom: this is a custom level\n",
			output:    "11:17:39 \x1b[0;32;1mcustom: \x1b[0mthis is a custom level\n",
			width:     120,
			showColor: true,
			flags:     log.Ltime,
		},
		{ // suppress debug
			input:  "12:34:56 debug: should be suppressed",
			output: "",
			flags:  log.Ltime,
		},
		{ // suppress trace
			input:  "12:34:56 trace: should be suppressed",
			output: "",
			flags:  log.Ltime,
		},
		{ // verbose shows debug
			input:   "12:34:56 debug: should be displayed",
			output:  "12:34:56 debug: should be displayed\n",
			verbose: true,
			flags:   log.Ltime | log.LUTC,
		},
		{ // verbose shows trace
			input:   "12:34:56 trace: should be displayed",
			output:  "12:34:56 trace: should be displayed\n",
			verbose: true,
			flags:   log.Ltime,
		},
		{ // trailing white space
			input:   "12:34:56 trailing white space   ",
			output:  "12:34:56 trailing white space\n",
			verbose: true,
			flags:   log.Ltime,
		},
		{ // file format
			input:     "12:34:56 file.go:123 message",
			output:    "12:34:56 \x1b[0;90mfile.go:123: \x1b[0mmessage\n",
			verbose:   true,
			flags:     log.Ltime | log.Lshortfile,
			showColor: true,
		},
	}

	for tn, tt := range tests {
		var buf bytes.Buffer
		output := NewWriter(&buf)
		if !tt.verbose {
			output.Suppress("trace", "debug")
		}
		output.SetLevel("custom", "32;1")
		printer := &terminalPrinter{
			w:       &buf,
			nocolor: !tt.showColor,
		}
		if tt.width > 0 {
			printer.width = func() int { return tt.width }
		} else {
			printer.width = func() int { return 999999 }
		}
		output.printer = printer
		logger := log.New(ioutil.Discard, tt.prefix, tt.flags)
		writer := newLogWriter(output, logger)
		writer.Write([]byte(tt.input))
		str := buf.String()
		if got, want := str, tt.output; got != want {
			t.Errorf("%d:\n got=%q\nwant=%q", tn, got, want)
		}
	}
}

type testHandler struct {
	handles func(prefix, level string) bool
	handle  func(*Message)
}

func (th *testHandler) Handles(prefix, level string) bool {
	if th.handles != nil {
		return th.handles(prefix, level)
	}
	return true
}

func (th *testHandler) Handle(msg *Message) {
	if th.handle != nil {
		th.handle(msg)
	}
}

func TestOutput(t *testing.T) {
	tests := []struct {
		text   string
		logger *log.Logger
		entry  *logEntry
	}{
		{
			text:   "message text a=1 b=2",
			logger: log.New(ioutil.Discard, "", log.LstdFlags),
			entry: &logEntry{
				Date: []byte("0000/00/00"),
				Time: []byte("00:00:00"),
				Text: []byte("message text"),
				List: [][]byte{b("a"), b("1"), b("b"), b("2")},
			},
		},
		{
			text:   "debug: message text a=1 b=2",
			logger: log.New(ioutil.Discard, "", log.LstdFlags),
			entry:  nil, // message suppressed
		},
		{
			text:   "warning: message text a=1 b=2",
			logger: log.New(ioutil.Discard, "", log.LstdFlags),
			entry: &logEntry{
				Date:   []byte("0000/00/00"),
				Time:   []byte("00:00:00"),
				Level:  "warning",
				Effect: "yellow",
				Text:   b("message text"),
				List:   [][]byte{b("a"), b("1"), b("b"), b("2")},
			},
		},
		{
			text:   "message text a=1 b=2",
			logger: log.New(ioutil.Discard, "a really long prefix", log.LstdFlags),
			entry: &logEntry{
				Prefix: "a really long prefix",
				Date:   []byte("0000/00/00"),
				Time:   []byte("00:00:00"),
				Text:   b("message text"),
				List:   [][]byte{b("a"), b("1"), b("b"), b("2")},
			},
		},
		{
			text:   "message text a=1 b=2",
			logger: log.New(ioutil.Discard, "a really long prefix", log.LstdFlags|log.Lmicroseconds),
			entry: &logEntry{
				Prefix: "a really long prefix",
				Date:   []byte("0000/00/00"),
				Time:   []byte("00:00:00.000000"),
				Text:   b("message text"),
				List:   [][]byte{b("a"), b("1"), b("b"), b("2")},
			},
		},
		{
			text:   "message text a=1 b=2",
			logger: log.New(ioutil.Discard, "", log.Llongfile),
			entry: &logEntry{
				File: []byte("present"),
				Text: b("message text"),
				List: [][]byte{b("a"), b("1"), b("b"), b("2")},
			},
		},
		{
			text:   "error: message text a=1 b=2",
			logger: log.New(ioutil.Discard, "", log.Lshortfile),
			entry: &logEntry{
				File:   []byte("present"),
				Level:  "error",
				Effect: "red",
				Text:   b("message text"),
				List:   [][]byte{b("a"), b("1"), b("b"), b("2")},
			},
		},
		{
			text:   "message text a=1 b=2",
			logger: log.New(ioutil.Discard, "2049-07-08", 0),
			entry: &logEntry{
				Prefix: "2049-07-08",
				Text:   b("message text"),
				List:   [][]byte{b("a"), b("1"), b("b"), b("2")},
			},
		},
	}

	output := NewWriter(ioutil.Discard)
	output.Suppress("debug")
	var entry *logEntry
	output.entryHandler = func(e *logEntry) {
		// need to clone the entry, because it gets zeroed out
		// after this handler is called
		entry = cloneEntry(e)
	}

	for tn, tt := range tests {
		t.Run(strconv.Itoa(tn), func(t *testing.T) {
			output.Attach(tt.logger)
			entry = nil
			tt.logger.Println(tt.text)
			if got, want := entry, tt.entry; !entriesEqual(got, want) {
				t.Errorf("\n got=%+v\nwant=%+v", got, want)
			}
		})
	}
}

func cloneByteSlice(slice []byte) []byte {
	if slice == nil {
		return nil
	}
	clone := make([]byte, len(slice))
	copy(clone, slice)
	return clone
}

func cloneByteSliceSlice(slice [][]byte) [][]byte {
	if slice == nil {
		return nil
	}
	clone := make([][]byte, len(slice))
	for i, s := range slice {
		clone[i] = cloneByteSlice(s)
	}
	return clone
}

func cloneEntry(e *logEntry) *logEntry {
	return &logEntry{
		Timestamp: e.Timestamp,
		Prefix: e.Prefix,
		Date: cloneByteSlice(e.Date),
		Time: cloneByteSlice(e.Time),
		File: cloneByteSlice(e.File),
		Level: e.Level,
		Effect: e.Effect,
		Text: cloneByteSlice(e.Text),
		List: cloneByteSliceSlice(e.List),
	}
}

func entriesEqual(e1, e2 *logEntry) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if string(e1.Prefix) != string(e2.Prefix) ||
		len(e1.Date) != len(e2.Date) ||
		len(e1.Time) != len(e2.Time) ||
		e1.Level != e2.Level ||
		e1.Effect != e2.Effect ||
		string(e1.Text) != string(e2.Text) {
		return false
	}

	if len(e1.File) > 0 && len(e2.File) == 0 {
		return false
	}
	if len(e1.File) == 0 && len(e2.File) > 0 {
		return false
	}

	if len(e1.List) > 0 || len(e2.List) > 0 {
		if !reflect.DeepEqual(e1.List, e2.List) {
			return false
		}
	}
	return true
}

func b(s string) []byte {
	return []byte(s)
}

// String implements the fmt.Stringer interface, and is useful for printing
// entries in unit tests.
func (e *logEntry) String() string {
	var buf bytes.Buffer
	p := simplePrinter{w: &buf}
	p.Print(e)
	return buf.String()
}

func BenchmarkStdLog(b *testing.B) {
	logger := log.New(ioutil.Discard, "testing", log.LstdFlags)
	benchmarkLog(b, logger)
}

func BenchmarkKVLog(b *testing.B) {
	logger := log.New(ioutil.Discard, "testing", log.LstdFlags)
	w := NewWriter(ioutil.Discard)
	w.Attach(logger)
	benchmarkLog(b, logger)
}

func BenchmarkSuppress(b *testing.B) {
	logger := log.New(ioutil.Discard, "testing", log.LstdFlags)
	w := NewWriter(ioutil.Discard)
	w.Attach(logger)
	w.Suppress("info")
	benchmarkLog(b, logger)
}

func benchmarkLog(b *testing.B, logger *log.Logger) {
	b.ReportAllocs()
	kv := kv.With("n", 0)
	for n := 0; n < b.N; n++ {
		kv[1] = n
		logger.Println("info: message", kv)
	}
}
