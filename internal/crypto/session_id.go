package crypto

import (
	"encoding/binary"
	"io"
)

// NewSessionID generates new random int64 from reader.
//
// Use crypto/rand.Reader if session id should be cryptographically safe.
func NewSessionID(reader io.Reader) (int64, error) {
	bytes := make([]byte, 8)
	if _, err := io.ReadFull(reader, bytes); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(bytes)), nil
}
