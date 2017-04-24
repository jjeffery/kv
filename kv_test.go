package kv

import (
	"encoding"
	"fmt"
	"reflect"
	"testing"
)

type testKeyvalPairer struct {
	key   string
	value interface{}
}

func (p testKeyvalPairer) keyvalPair() (string, interface{}) {
	return p.key, p.value
}

func TestKeyvals(t *testing.T) {
	tests := []struct {
		keyvalser keyvalser
		want      []interface{}
	}{
		{
			keyvalser: List{"k1", 1, "k2", 2},
			want:      []interface{}{"k1", 1, "k2", 2},
		},
		{
			keyvalser: Map{"k1": 1},
			want:      []interface{}{"k1", 1},
		},
		{
			keyvalser: Map{"k1": 1, "k2": 2, "k3": 3, "a1": "1", "a3": "3", "a2": "2"},
			want:      []interface{}{"a1", "1", "a2", "2", "a3", "3", "k1", 1, "k2", 2, "k3", 3},
		},
		{
			keyvalser: Pair{"k1", 1},
			want:      []interface{}{"k1", 1},
		},
	}

	for i, tt := range tests {
		if got := tt.keyvalser.Keyvals(); !reflect.DeepEqual(tt.want, got) {
			t.Errorf("%d: want=%v, got=%v", i, tt.want, got)
		}
	}
}

func TestKeyvalPair(t *testing.T) {
	tests := []struct {
		keyvalPairer keyvalPairer
		wantKey      string
		wantValue    interface{}
	}{
		{
			keyvalPairer: testKeyvalPairer{"k1", 1},
			wantKey:      "k1",
			wantValue:    1,
		},
		{
			keyvalPairer: Pair{"k1", 1},
			wantKey:      "k1",
			wantValue:    1,
		},
	}

	for i, tt := range tests {
		gotKey, gotValue := tt.keyvalPairer.keyvalPair()
		if gotKey != tt.wantKey || !reflect.DeepEqual(tt.wantValue, gotValue) {
			t.Errorf("%d: want=[%s, %v], got=[%s, %v]", i, tt.wantKey, tt.wantValue, gotKey, gotValue)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{
			input: Pair{"key", "value"},
			want:  "key=value",
		},
		{
			input: Map{"key": "value"},
			want:  "key=value",
		},
		{
			input: List{"key1", "value1", "key2", "value2"},
			want:  "key1=value1 key2=value2",
		},
	}
	for i, tt := range tests {
		stringer, ok := tt.input.(fmt.Stringer)
		if !ok {
			t.Errorf("expected fmt.Stringer")
		} else {
			if got, want := stringer.String(), tt.want; got != want {
				t.Errorf("%d: got=%s want=%s", i, got, want)
			}
		}
		marshaler, ok := tt.input.(encoding.TextMarshaler)
		if !ok {
			t.Errorf("expected encoding.TextMarshaler")
		} else {
			b, err := marshaler.MarshalText()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := string(b), tt.want; got != want {
				t.Errorf("%d: got=%s want=%s", i, got, want)
			}
		}
	}
}
