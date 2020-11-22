package bin

import "io"

func encodeString(b []byte, v string) []byte {
	l := len(v)
	if l <= 253 {
		// If L <= 253, the serialization contains one byte with the value of L,
		// then L bytes of the string followed by 0 to 3 characters containing 0,
		// such that the overall length of the value be divisible by 4,
		// whereupon all of this is interpreted as a sequence of int(L/4)+1 32-bit numbers.
		b = append(b, byte(l))
		b = append(b, v...)
		currentLen := l + 1
		b = append(b, make([]byte, nearestPaddedValueLength(currentLen)-currentLen)...)
		return b
	}

	// If L >= 254, the serialization contains byte 254, followed by 3 bytes with
	// the string length L, followed by L bytes of the string, further followed
	// by 0 to 3 null padding bytes.
	b = append(b, 254, byte(l), byte(l>>8), byte(l>>16))
	b = append(b, v...)
	currentLen := l + 4
	b = append(b, make([]byte, nearestPaddedValueLength(currentLen)-currentLen)...)

	return b
}

func decodeString(b []byte) (n int, v string, err error) {
	if b[0] == 254 {
		strLen := uint32(b[1]) | uint32(b[2])<<8 | uint32(b[3])<<16
		if len(b) < (int(strLen) + 4) {
			return 0, "", io.ErrUnexpectedEOF
		}
		return nearestPaddedValueLength(int(strLen) + 4), string(b[4 : strLen+4]), nil
	}
	strLen := b[0]
	if len(b) < (int(strLen) + 1) {
		return 0, "", io.ErrUnexpectedEOF
	}
	return nearestPaddedValueLength(int(strLen) + 1), string(b[1 : strLen+1]), nil
}
