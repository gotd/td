package crypto

import (
	"crypto/sha256"
	"hash"
	"sync"

	"github.com/ernado/td/bin"
)

// See https://core.telegram.org/mtproto/description#defining-aes-key-and-initialization-vector

// AuthKey represents 2048-bit authorization key.
type AuthKey [256]byte

// ID returns auth_key_id.
func (k AuthKey) ID() int64 {
	return AuthKeyID(k)
}

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

// Message keys are defined here:
// * https://core.telegram.org/mtproto/description#defining-aes-key-and-initialization-vector

// msgKeyLarge appends msg_key_large value to b.
func msgKeyLarge(authKey AuthKey, plaintextPadded []byte, mode Mode) []byte {
	h := getSHA256()
	defer sha256Pool.Put(h)

	x := getX(mode)
	_, _ = h.Write(authKey[88+x : 32+88+x])
	_, _ = h.Write(plaintextPadded)
	return h.Sum(nil)
}

// messageKey returns msg_key = substr (msg_key_large, 8, 16).
func messageKey(messageKeyLarge []byte) bin.Int128 {
	var v bin.Int128
	b := messageKeyLarge[8 : 16+8]
	copy(v[:len(b)], b)
	return v
}

// sha256a appends sha256_a value to b.
//
// sha256_a = SHA256 (msg_key + substr (auth_key, x, 36));
func sha256a(b []byte, authKey AuthKey, msgKey bin.Int128, mode Mode) []byte {
	h := getSHA256()
	defer sha256Pool.Put(h)

	_, _ = h.Write(msgKey[:])
	_, _ = h.Write(authKey[getX(mode) : getX(mode)+36])

	return h.Sum(b)
}

// sha256b appends sha256_b value to b.
//
// sha256_b = SHA256 (substr (auth_key, 40+x, 36) + msg_key);
func sha256b(b []byte, authKey AuthKey, msgKey bin.Int128, mode Mode) []byte {
	h := getSHA256()
	defer sha256Pool.Put(h)

	_, _ = h.Write(authKey[40+getX(mode) : 40+getX(mode)+36])
	_, _ = h.Write(msgKey[:])

	return h.Sum(b)
}

// aesKey appends aes_key to b.
//
// aes_key = substr (sha256_a, 0, 8) + substr (sha256_b, 8, 16) + substr (sha256_a, 24, 8);
func aesKey(sha256a, sha256b []byte) bin.Int256 {
	var v bin.Int256
	copy(v[:8], sha256a[:8])
	copy(v[8:], sha256b[8:16+8])
	copy(v[24:], sha256a[24:24+8])
	return v
}

// aesIV appends aes_iv to b.
//
// aes_iv = substr (sha256_b, 0, 8) + substr (sha256_a, 8, 16) + substr (sha256_b, 24, 8);
func aesIV(sha256a, sha256b []byte) bin.Int256 {
	// Same as aes_key, but with swapped params.
	return aesKey(sha256b, sha256a)
}

// Keys for message.
type Keys struct {
	MessageKey bin.Int128 // msg_key
	Key        bin.Int256 // aes_key
	IV         bin.Int256 // aes_iv
}

// MessageKeys returns aes_key and aes_iv for plaintext message.
// Basically it is "KDF" in diagram.
//
// Reference:
// * https://core.telegram.org/mtproto/description#defining-aes-key-and-initialization-vector
func MessageKeys(authKey AuthKey, plaintextPadded []byte, mode Mode) *Keys {
	// `msg_key_large = SHA256 (substr (auth_key, 88+x, 32) + plaintext + random_padding);`
	msgKeyLarge := msgKeyLarge(authKey, plaintextPadded, mode)
	// `msg_key = substr (msg_key_large, 8, 16);`
	msgKey := messageKey(msgKeyLarge)

	// `sha256_a = SHA256 (msg_key + substr (auth_key, x, 36));`
	a := sha256a(nil, authKey, msgKey, mode)
	// `sha256_b = SHA256 (substr (auth_key, 40+x, 36) + msg_key);`
	b := sha256b(nil, authKey, msgKey, mode)

	return &Keys{
		Key:        aesKey(a, b),
		IV:         aesIV(a, b),
		MessageKey: msgKey,
	}
}
