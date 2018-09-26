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
