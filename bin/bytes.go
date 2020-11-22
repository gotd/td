package bin

import "io"

// encodeBytes is same as encodeString, but for bytes.
func encodeBytes(b, v []byte) []byte {
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

// decodeBytes is same as decodeString, but for bytes.
//
// NB: v is slice of b.
func decodeBytes(b []byte) (n int, v []byte, err error) {
	if b[0] == firstLongStringByte {
		vLen := uint32(b[1]) | uint32(b[2])<<8 | uint32(b[3])<<16
		if len(b) < (int(vLen) + 4) {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return nearestPaddedValueLength(int(vLen) + 4), b[4 : vLen+4], nil
	}
	vLen := b[0]
	if len(b) < (int(vLen) + 1) {
		return 0, nil, io.ErrUnexpectedEOF
	}
	return nearestPaddedValueLength(int(vLen) + 1), b[1 : vLen+1], nil
}
