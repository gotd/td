package crypto

import (
	"crypto/sha1" // #nosec
	"encoding/binary"
)

// AuthKeyID returns auth_key_id (64 lower-order bits of the SHA1).
func AuthKeyID(k AuthKey) int64 {
	raw := sha1.Sum(k[:]) // #nosec
	return int64(binary.BigEndian.Uint64(raw[12:]))
}
