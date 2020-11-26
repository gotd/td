package telegram

import "crypto/sha1" // #nosec

func sha(v []byte) []byte {
	h := sha1.Sum(v) // #nosec
	return h[:]
}
