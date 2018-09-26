package kvlog

import (
	"bytes"
	"testing"
)

func TestWriter(t *testing.T) {
	tests := []struct {
		input  string
		output string
		width  int
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
			input:  "12:34:56 error: this is the message key1=value1 key2=value2: file not found\n",
			output: "12:34:56 error: this is the message key1=value1 key2=value2\n         file not found\n",
			width:  70,
		},
		{
			input: "12:34:56 error: this is a very long message that will wrap over the line key1=value1 key2=value2: file not found\n",
			output: "12:34:56 error: this is a very long message that will wrap\n" +
				"         over the line key1=value1 key2=value2\n" +
				"         file not found\n",
			width: 60,
		},
	}

	for tn, tt := range tests {
		var buf bytes.Buffer
		writer := NewWriter(&buf)
		if tt.width > 0 {
			writer.Width = func() int { return tt.width }
		}
		writer.Write([]byte(tt.input))
		output := buf.String()
		if got, want := output, tt.output; got != want {
			t.Errorf("%d:\n got=%q\nwant=%q", tn, got, want)
		}
	}

}
