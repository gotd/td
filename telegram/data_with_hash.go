package telegram

import (
	"crypto/rand"
	// #nosec
	//
	// Allowing sha1 because it is used in MTProto itself.
	"crypto/sha1"
	"io"
)

// data_with_hash := SHA1(data) + data + (any random bytes); such that the length equals 255 bytes;
func newDataWithHash(data []byte) ([255]byte, error) {
	// data_with_hash := SHA1(data) + data + (any random bytes); such that the length equals 255 bytes;
	var dataWithHash = [255]byte{}
	if _, err := io.ReadFull(rand.Reader, dataWithHash[:]); err != nil {
		return dataWithHash, err
	}
	h := sha1.New() // #nosec
	if _, err := h.Write(data); err != nil {
		return dataWithHash, err
	}
	copy(dataWithHash[:sha1.Size], h.Sum(nil))
	copy(dataWithHash[sha1.Size:], data)
	return dataWithHash, nil
}
