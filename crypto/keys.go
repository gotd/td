package crypto

import (
	"crypto/sha256"

	"github.com/gotd/td/bin"
)

// Side on which encryption is performed.
type Side byte

const (
	// Client side of encryption (e.g. messages from client).
	Client Side = 0
	// Server side of encryption (e.g. RPC responses).
	Server Side = 1
)

// DecryptSide returns Side for decryption.
func (s Side) DecryptSide() Side {
	return s ^ 1 // flips bit, so 0 becomes 1, 1 becomes 0
}

func getX(mode Side) int {
	switch mode {
	case Client:
		return 0
	case Server:
		return 8
	default:
		return 0
	}
}

// Message keys are defined here:
// * https://core.telegram.org/mtproto/description#defining-aes-key-and-initialization-vector

// msgKeyLarge returns msg_key_large value.
func msgKeyLarge(r []byte, authKey Key, plaintextPadded []byte, mode Side) []byte {
	h := sha256.New()

	x := getX(mode)
	_, _ = h.Write(authKey[88+x : 32+88+x])
	_, _ = h.Write(plaintextPadded)
	return h.Sum(r)
}

// messageKey returns msg_key = substr (msg_key_large, 8, 16).
func messageKey(messageKeyLarge []byte) (v bin.Int128) {
	b := messageKeyLarge[8 : 16+8]
	copy(v[:len(b)], b)
	return v
}

// sha256a returns sha256_a value.
//
// sha256_a = SHA256 (msg_key + substr (auth_key, x, 36));
func sha256a(r []byte, authKey *Key, msgKey *bin.Int128, x int) []byte {
	h := sha256.New()

	_, _ = h.Write(msgKey[:])
	_, _ = h.Write(authKey[x : x+36])

	return h.Sum(r)
}

// sha256b returns sha256_b value.
//
// sha256_b = SHA256 (substr (auth_key, 40+x, 36) + msg_key);
func sha256b(r []byte, authKey *Key, msgKey *bin.Int128, x int) []byte {
	h := sha256.New()

	_, _ = h.Write(authKey[40+x : 40+x+36])
	_, _ = h.Write(msgKey[:])

	return h.Sum(r)
}

// aesKey returns aes_key value.
//
// aes_key = substr (sha256_a, 0, 8) + substr (sha256_b, 8, 16) + substr (sha256_a, 24, 8);
func aesKey(sha256a, sha256b []byte, v *bin.Int256) {
	copy(v[:8], sha256a[:8])
	copy(v[8:], sha256b[8:16+8])
	copy(v[24:], sha256a[24:24+8])
}

// aesIV returns aes_iv value.
//
// aes_iv = substr (sha256_b, 0, 8) + substr (sha256_a, 8, 16) + substr (sha256_b, 24, 8);
func aesIV(sha256a, sha256b []byte, v *bin.Int256) {
	// Same as aes_key, but with swapped params.
	aesKey(sha256b, sha256a, v)
}

// Keys returns (aes_key, aes_iv) pair for AES-IGE.
//
// See https://core.telegram.org/mtproto/description#defining-aes-key-and-initialization-vector
//
// Example:
//
//	key, iv := crypto.Keys(authKey, messageKey, crypto.Client)
//	cipher, err := aes.NewCipher(key[:])
//	if err != nil {
//		return nil, err
//	}
//	encryptor := ige.NewIGEEncrypter(cipher, iv[:])
func Keys(authKey Key, msgKey bin.Int128, mode Side) (key, iv bin.Int256) {
	x := getX(mode)

	r := make([]byte, 512)
	// `sha256_a = SHA256 (msg_key + substr (auth_key, x, 36));`
	a := sha256a(r[0:0], &authKey, &msgKey, x)
	// `sha256_b = SHA256 (substr (auth_key, 40+x, 36) + msg_key);`
	b := sha256b(r[256:256], &authKey, &msgKey, x)

	aesKey(a, b, &key)
	aesIV(a, b, &iv)
	return key, iv
}

// MessageKey computes message key for provided auth_key and padded payload.
func MessageKey(authKey Key, plaintextPadded []byte, mode Side) bin.Int128 {
	r := make([]byte, 0, 256)
	// `msg_key_large = SHA256 (substr (auth_key, 88+x, 32) + plaintext + random_padding);`
	msgKeyLarge := msgKeyLarge(r, authKey, plaintextPadded, mode)
	// `msg_key = substr (msg_key_large, 8, 16);`
	return messageKey(msgKeyLarge)
}
