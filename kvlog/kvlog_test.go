package kvlog

import (
	"bytes"
	"testing"
)

func TestWriter(t *testing.T) {
	tests := []struct {
		input     string
		output    string
		width     int
		showColor bool
		verbose   bool
	}{
		{
			input: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			output: "2099/12/31 12:34:56 this is the message key1=value1\n" +
				"                    key2=value2 key3=value3\n",
			width: 60,
		},
		{
			input: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			output: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2\n" +
				"                    key3=value3\n",
			width: 70,
		},
		{
			input:  "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			output: "2099/12/31 12:34:56 this is the message key1=value1 key2=value2 key3=value3\n",
			width:  80,
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
		},
		{
			input: "12:34:56 error: this is the message key1=value1 key2=value2: file not found\n",
			output: "12:34:56 error: this is the message key1=value1 key2=value2: file not\n" +
				"         found\n",
			width: 70,
		},
		{
			input: "12:34:56 error: this is a very long message that will wrap over the line key1=value1 key2=value2: file not found\n",
			output: "12:34:56 error: this is a very long message that will wrap\n" +
				"         over the line key1=value1 key2=value2: file not\n" +
				"         found\n",
			width: 60,
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
		},
		{ // colors
			input:     "11:17:39 prefix: error: this is an error message\n",
			output:    "11:17:39 prefix: \x1b[0;31merror:\x1b[0m this is an error message\n",
			width:     120,
			showColor: true,
		},
		{ // colors
			input:     "11:17:39 prefix: Error: this is an error message\n",
			output:    "11:17:39 prefix: \x1b[0;31mError:\x1b[0m this is an error message\n",
			width:     120,
			showColor: true,
		},
		{ // suppress debug
			input:  "12:34:56 debug: should be suppressed",
			output: "",
		},
		{ // suppress trace
			input:  "12:34:56 trace: should be suppressed",
			output: "",
		},
		{ // verbose shows debug
			input:   "12:34:56 debug: should be displayed",
			output:  "12:34:56 debug: should be displayed\n",
			verbose: true,
		},
		{ // verbose shows trace
			input:   "12:34:56 trace: should be displayed",
			output:  "12:34:56 trace: should be displayed\n",
			verbose: true,
		},
	}

	for tn, tt := range tests {
		var buf bytes.Buffer
		writer := NewWriter(&buf)
		if tt.showColor {
			writer.colorOutput = true
		}
		if tt.width > 0 {
			writer.Width = func() int { return tt.width }
		}
		writer.Verbose = tt.verbose
		writer.Write([]byte(tt.input))
		output := buf.String()
		if got, want := output, tt.output; got != want {
			t.Errorf("%d:\n got=%q\nwant=%q", tn, got, want)
		}
	}

}

func TestNoColor(t *testing.T) {
	w1 := &dummyWriter{n: 1}
	w2 := &dummyWriter{n: 2}
	w := &Writer{
		colorOutput: true,
		origOut:     w1,
		Out:         w2,
	}
	w = w.NoColor()
	if got, want := w.colorOutput, false; got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
	if got, want := w.Out, w1; got != want {
		t.Errorf("got=%v, want=%v", got, want)
	}
}

type dummyWriter struct {
	// don't make a struct{}, otherwise all instances
	// point to the same address
	n int // does nothing
}

func (dw *dummyWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
