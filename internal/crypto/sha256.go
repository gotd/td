package crypto

import (
	"crypto/sha256"
	"hash"
	"sync"
)

// SHA256 returns SHA256 hash.
func SHA256(from ...[]byte) []byte {
	h := getSHA256()
	defer sha256Pool.Put(h)
	for _, b := range from {
		_, _ = h.Write(b)
	}
	return h.Sum(nil)
}

var sha256Pool = &sync.Pool{
	New: func() interface{} {
		return sha256.New()
	},
}

func getSHA256() hash.Hash {
	h := sha256Pool.Get().(hash.Hash)
	h.Reset()
	return h
}
