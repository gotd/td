package testutil

import (
	"encoding/binary"
	"math/rand"
)

// ZeroRand is zero random source.
type ZeroRand struct{}

// Read implements io.Reader.
func (ZeroRand) Read(p []byte) (n int, err error) { return len(p), nil }

func randSeed(data []byte) int64 {
	if len(data) == 0 {
		return 0
	}

	seedBuf := make([]byte, 64/8)
	copy(seedBuf, data)

	return int64(binary.BigEndian.Uint64(seedBuf))
}

// Rand returns a new rand.Rand with source deterministically initialized
// from seed byte slice.
//
// Zero length seed (or nil) is valid input.
func Rand(seed []byte) *rand.Rand {
	return rand.New(rand.NewSource(randSeed(seed))) // #nosec
}
