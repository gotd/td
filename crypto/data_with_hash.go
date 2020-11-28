package crypto

import (
	"bytes"
	"crypto/sha1" // #nosec
	"io"
)

// GuessDataWithHash guesses data from data_with_hash.
func GuessDataWithHash(dataWithHash []byte) []byte {
	// data_with_hash := SHA1(data) + data + (0-15 random bytes);
	// such that length be divisible by 16;
	if len(dataWithHash) <= sha1.Size {
		// Data length too small.
		return nil
	}
	v := dataWithHash[:sha1.Size]
	h := sha1.New() // #nosec
	var currentHash []byte
	for i := 0; i < 16; i++ {
		h.Reset()
		if len(dataWithHash)-i < sha1.Size {
			// End of slice reached.
			return nil
		}
		data := dataWithHash[sha1.Size : len(dataWithHash)-i]
		_, _ = h.Write(data)
		currentHash = h.Sum(currentHash[:0])
		if bytes.Equal(currentHash, v) {
			// Found.
			return data
		}
	}
	return nil
}

func paddedLen16(l int) int {
	n := 16 * (l / 16)
	if n < l {
		n += 16
	}
	return n
}

// DataWithHash prepends data with SHA1(data) and 0..15 random bytes so result
// length is divisible by 16.
//
// Use GuessDataWithHash(result) to obtain data.
func DataWithHash(data []byte, randomSource io.Reader) ([]byte, error) {
	dataWithHash := make([]byte, paddedLen16(len(data)+sha1.Size))
	h := sha1.Sum(data) // #nosec
	copy(dataWithHash, h[:])
	copy(dataWithHash[sha1.Size:], data)
	if _, err := io.ReadFull(randomSource, dataWithHash[sha1.Size+len(data):]); err != nil {
		return nil, err
	}
	return dataWithHash, nil
}
