package kv

import (
	"reflect"
	"testing"
)

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
