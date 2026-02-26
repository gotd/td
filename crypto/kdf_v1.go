package crypto

import (
	"crypto/sha1" // #nosec G505

	"github.com/gotd/td/bin"
)

// MessageKeyV1 computes msg_key for MTProto v1.
//
// msg_key = SHA1(plaintext)[4:20]
func MessageKeyV1(plaintext []byte) (v bin.Int128) {
	sum := sha1.Sum(plaintext) // #nosec G401
	copy(v[:], sum[4:20])
	return v
}

// KeysV1 returns (aes_key, aes_iv) pair for MTProto v1.
//
// The KDF is used by auth.bindTempAuthKey encrypted_message.
// It intentionally mirrors the old "x=0 client side" key schedule, because
// Telegram explicitly requires this exact derivation for binding payload.
func KeysV1(authKey Key, msgKey bin.Int128) (key, iv bin.Int256) {
	r := make([]byte, sha1.Size*4)

	// x = 0 (client side) for binding message.
	a := sha1a(r[0:0], authKey, msgKey, 0)
	b := sha1b(r[sha1.Size:sha1.Size], authKey, msgKey, 0)
	c := sha1c(r[2*sha1.Size:2*sha1.Size], authKey, msgKey, 0)
	d := sha1d(r[3*sha1.Size:3*sha1.Size], authKey, msgKey, 0)

	// aes_key = sha1_a[0:8] + sha1_b[8:20] + sha1_c[4:16]
	n := copy(key[:], a[:8])
	n += copy(key[n:], b[8:20])
	copy(key[n:], c[4:16])

	// aes_iv = sha1_a[8:20] + sha1_b[0:8] + sha1_c[16:20] + sha1_d[0:8]
	n = copy(iv[:], a[8:20])
	n += copy(iv[n:], b[:8])
	n += copy(iv[n:], c[16:20])
	copy(iv[n:], d[:8])

	return key, iv
}
