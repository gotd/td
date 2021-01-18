package tlstypes

// ReverseBytes is a common slice reverser.
func ReverseBytes(data []byte) []byte {
	dataLen := len(data)
	rv := make([]byte, dataLen)
	rv[dataLen/2] = data[dataLen/2]

	for i := dataLen/2 - 1; i >= 0; i-- {
		opp := dataLen - i - 1
		rv[i], rv[opp] = data[opp], data[i]
	}

	return rv
}

type Uint24 [3]byte

func ToUint24(number uint32) Uint24 {
	return Uint24{byte(number), byte(number >> 8), byte(number >> 16)}
}

func FromUint24(number Uint24) uint32 {
	return uint32(number[0]) + (uint32(number[1]) << 8) + (uint32(number[2]) << 16)
}
