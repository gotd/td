package ascii

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testDigit    = []rune("0123456789")
	acsiiLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
)

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
		require.Truef(t, IsDigit(r), "IsDigit(U+%04X)", r)
	}
	for _, r := range testLetter {
		require.Falsef(t, IsDigit(r), "IsDigit(U+%04X)", r)
	}
}

func TestIsLatinLetter(t *testing.T) {
	for _, r := range acsiiLetters {
		require.True(t, IsLatinLetter(r), "IsLatinLetter(U+%04X)", r)
	}
	for _, r := range testDigit {
		require.Falsef(t, IsLatinLetter(r), "IsLatinLetter(U+%04X)", r)
	}
}
