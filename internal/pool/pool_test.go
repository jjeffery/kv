// +build !race

package pool

import "testing"

func TestPool(t *testing.T) {
	buf1 := AllocBuffer()
	buf1.WriteString("some data")
	ReleaseBuffer(buf1)
	buf2 := AllocBuffer()
	if got, want := buf2.Len(), 0; got != want {
		t.Fatalf("got=%v want=%v", got, want)
	}
	if got, want := buf2, buf1; got != want {
		t.Fatalf("got=%v want=%v", got, want)
	}

	// just test that we don't panic
	ReleaseBuffer(nil)
}
