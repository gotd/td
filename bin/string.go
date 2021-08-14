package bin

import (
	"errors"
)

const (
	// If L <= 253, the serialization contains one byte with the value of L,
	// then L bytes of the string followed by 0 to 3 characters containing 0,
	// such that the overall length of the value be divisible by 4,
	// whereupon all of this is interpreted as a sequence of int(L/4)+1 32-bit numbers.
	maxSmallStringLength = 253
	// If L >= 254, the serialization contains byte 254, followed by 3 bytes with
	// the string length L, followed by L bytes of the string, further followed
	// by 0 to 3 null padding bytes.
	firstLongStringByte = 254
)

func encodeString(b []byte, v string) []byte {
	l := len(v)
	if l <= maxSmallStringLength {
		b = append(b, byte(l))
		b = append(b, v...)
		currentLen := l + 1
		b = append(b, make([]byte, nearestPaddedValueLength(currentLen)-currentLen)...)
		return b
	}

	b = append(b, firstLongStringByte, byte(l), byte(l>>8), byte(l>>16))
	b = append(b, v...)
	currentLen := l + 4
	b = append(b, make([]byte, nearestPaddedValueLength(currentLen)-currentLen)...)

	return b
}

var errInvalidLength = errors.New("invalid length")

func decodeString(b []byte) (padding int, v string, err error) {
	n, v1, err := decodeBytes(b)
	return n, string(v1), err
}
