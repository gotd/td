package crypto

import (
	"crypto/sha1" // #nosec G505

	"github.com/gotd/td/bin"
)

// sha1a returns sha1_a value.
//
// sha1_a = SHA1 (msg_key + substr (auth_key, x, 32));
func sha1a(r []byte, authKey Key, msgKey bin.Int128, x int) []byte {
	h := sha1.New() // #nosec G401

	_, _ = h.Write(msgKey[:])
	_, _ = h.Write(authKey[x : x+32])

	return h.Sum(r)
}

// sha1b returns sha1_b value.
//
// sha1_b = SHA1 (substr (auth_key, 32+x, 16) + msg_key + substr (auth_key, 48+x, 16));
func sha1b(r []byte, authKey Key, msgKey bin.Int128, x int) []byte {
	h := sha1.New() // #nosec G401

	_, _ = h.Write(authKey[32+x : 32+x+16])
	_, _ = h.Write(msgKey[:])
	_, _ = h.Write(authKey[48+x : 48+x+16])

	return h.Sum(r)
}

// sha1c returns sha1_c value.
//
// sha1_c = SHA1 (substr (auth_key, 64+x, 32) + msg_key);
func sha1c(r []byte, authKey Key, msgKey bin.Int128, x int) []byte {
	h := sha1.New() // #nosec G401

	_, _ = h.Write(authKey[64+x : 64+x+32])
	_, _ = h.Write(msgKey[:])

	return h.Sum(r)
}

// sha1d returns sha1_d value.
//
// sha1_d = SHA1 (msg_key + substr (auth_key, 96+x, 32));
func sha1d(r []byte, authKey Key, msgKey bin.Int128, x int) []byte {
	h := sha1.New() // #nosec G401

	_, _ = h.Write(msgKey[:])
	_, _ = h.Write(authKey[96+x : 96+x+32])

	return h.Sum(r)
}

// OldKeys returns (aes_key, aes_iv) pair for AES-IGE.
//
// See https://core.telegram.org/mtproto/description_v1#defining-aes-key-and-initialization-vector
//
// Example:
//
//	key, iv := crypto.OldKeys(authKey, messageKey, crypto.Client)
//	cipher, err := aes.NewCipher(key[:])
//	if err != nil {
//		return nil, err
//	}
//	encryptor := ige.NewIGEEncrypter(cipher, iv[:])
//
// Warning: MTProto 1.0 is deprecated.
func OldKeys(authKey Key, msgKey bin.Int128, mode Side) (key, iv bin.Int256) {
	x := getX(mode)

	aesKey := func(sha1a, sha1b, sha1c []byte) (v bin.Int256) {
		// aes_key = substr (sha1_a, 0, 8) + substr (sha1_b, 8, 12) + substr (sha1_c, 4, 12);
		n := copy(v[:], sha1a[:8])
		n += copy(v[n:], sha1b[8:8+12])
		copy(v[n:], sha1c[4:4+12])
		return v
	}
	aesIV := func(sha1a, sha1b, sha1c, sha1d []byte) (v bin.Int256) {
		// aes_iv = substr(sha1_a, 8, 12) + substr(sha1_b, 0, 8) + substr(sha1_c, 16, 4) + substr(sha1_d, 0, 8);
		n := copy(v[:], sha1a[8:8+12])
		n += copy(v[n:], sha1b[:8])
		n += copy(v[n:], sha1c[16:16+4])
		copy(v[n:], sha1d[:8])
		return v
	}

	r := make([]byte, sha1.Size*4)

	a := sha1a(r[0:0], authKey, msgKey, x)
	b := sha1b(r[sha1.Size:sha1.Size], authKey, msgKey, x)
	c := sha1c(r[2*sha1.Size:2*sha1.Size], authKey, msgKey, x)
	d := sha1d(r[3*sha1.Size:3*sha1.Size], authKey, msgKey, x)

	return aesKey(a, b, c), aesIV(a, b, c, d)
}
