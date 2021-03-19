package ascii

import (
	"testing"
)

var testDigit = []rune{
	0x0030, // 0
	0x0031, // 1
	0x0032, // 2
	0x0033, // 3
	0x0034, // 4
	0x0035, // 5
	0x0036, // 6
	0x0037, // 7
	0x0038, // 8
	0x0039, // 9
}

var testLetter = []rune{
	0x0041,
	0x0061,
	0x00AA,
	0x00BA,
	0x00C8,
	0x00DB,
	0x00F9,
	0x0DC0,
	0x0EDD,
	0x1000,
	0x1200,
	0x1312,
	0x10000,
	0x10300,
	0x10400,
	0x20000,
	0x2F800,
	0x2FA1D,
}

func TestDigit(t *testing.T) {
	for _, r := range testDigit {
		if !IsDigit(r) {
			t.Errorf("IsDigit(U+%04X) = false, want true", r)
		}
	}
	for _, r := range testLetter {
		if IsDigit(r) {
			t.Errorf("IsDigit(U+%04X) = true, want false", r)
		}
	}
}
