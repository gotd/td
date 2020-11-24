package proto

import "testing"

func TestPadding(t *testing.T) {
	for i := 0; i < 1024*2; i++ {
		padded := paddedLen(i)
		if padded%16 != 0 {
			// "the resulting message length be divisible by 16 bytes"
			t.Error(i)
		}
		if padded%padding != 0 {
			t.Error(i)
		}
		if padded < i {
			t.Error("padded < i")
		}
	}
}
