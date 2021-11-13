package crypto

import (
	"encoding/binary"

	"github.com/go-faster/xor"

	"github.com/gotd/td/bin"
)

// ServerSalt computes server salt.
func ServerSalt(newNonce bin.Int256, serverNonce bin.Int128) (salt int64) {
	var serverSalt [8]byte
	copy(serverSalt[:], newNonce[:8])
	xor.Bytes(serverSalt[:], serverSalt[:], serverNonce[:8])
	return int64(binary.LittleEndian.Uint64(serverSalt[:]))
}
