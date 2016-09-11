package kv

import (
	"reflect"
	"testing"
)

type testKeyvalPairer struct {
	key   string
	value interface{}
}

func (p testKeyvalPairer) KeyvalPair() (string, interface{}) {
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
			keyvalser: Map{"k1": 1, "k2": 2},
			want:      []interface{}{"k1", 1, "k2", 2},
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
	}

	for i, tt := range tests {
		gotKey, gotValue := tt.keyvalPairer.KeyvalPair()
		if gotKey != tt.wantKey || !reflect.DeepEqual(tt.wantValue, gotValue) {
			t.Errorf("%d: want=[%s, %v], got=[%s, %v]", i, tt.wantKey, tt.wantValue, gotKey, gotValue)
		}
	}
}
