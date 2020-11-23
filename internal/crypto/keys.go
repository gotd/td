package crypto

import (
	"crypto/sha256"
	"hash"
	"sync"
)

// See https://core.telegram.org/mtproto/description#defining-aes-key-and-initialization-vector

type Mode byte

const (
	ModeClient = 0
	ModeServer = 1
)

func getX(mode Mode) int {
	switch mode {
	case ModeClient:
		return 0
	case ModeServer:
		return 8
	default:
		return 0
	}
}

// nolint:gochecknoglobals // optimization for sha256-hash reuse
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

func MsgKeyLarge(b, authKey, plaintextPadded []byte, mode Mode) []byte {
	h := getSHA256()
	defer sha256Pool.Put(h)

	x := getX(mode)
	_, _ = h.Write(authKey[88+x : 32+88+x])
	_, _ = h.Write(plaintextPadded)
	return h.Sum(b)
}

func MsgKey(msgKey []byte) []byte {
	return msgKey[8 : 16+8]
}

// SHA256A appends sha256_a value to b.
//
// sha256_a = SHA256 (msg_key + substr (auth_key, x, 36));
func SHA256A(b, authKey, msgKey []byte, mode Mode) []byte {
	h := getSHA256()
	defer sha256Pool.Put(h)

	_, _ = h.Write(msgKey)
	_, _ = h.Write(authKey[getX(mode) : getX(mode)+36])

	return h.Sum(b)
}

// SHA256B appends sha256_b value to b.
//
// sha256_b = SHA256 (substr (auth_key, 40+x, 36) + msg_key);
func SHA256B(b, authKey, msgKey []byte, mode Mode) []byte {
	h := getSHA256()
	defer sha256Pool.Put(h)

	_, _ = h.Write(authKey[40+getX(mode) : 40+getX(mode)+36])
	_, _ = h.Write(msgKey)

	return h.Sum(b)
}

// AESKey appends aes_key to b.
//
// aes_key = substr (sha256_a, 0, 8) + substr (sha256_b, 8, 16) + substr (sha256_a, 24, 8);
func AESKey(b, sha256a, sha256b []byte) []byte {
	b = append(b, sha256a[:8]...)
	b = append(b, sha256b[8:16+8]...)
	b = append(b, sha256a[24:24+8]...)
	return b
}

// AESIV appends aes_iv to b.
//
// aes_iv = substr (sha256_b, 0, 8) + substr (sha256_a, 8, 16) + substr (sha256_b, 24, 8);
func AESIV(b, sha256a, sha256b []byte) []byte {
	// Same as aes_key, but with swapped params.
	return AESKey(b, sha256b, sha256a)
}
