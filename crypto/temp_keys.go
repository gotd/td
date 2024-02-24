package crypto

import (
	"crypto/sha1" // #nosec
	"math/big"
)

// sha1BigInt returns SHA1(a + b).
func sha1BigInt(a, b *big.Int) []byte {
	var buf []byte
	buf = append(buf, a.Bytes()...)
	buf = append(buf, b.Bytes()...)

	h := sha1.Sum(buf) // #nosec
	return h[:]
}

// TempAESKeys returns tmp_aes_key and tmp_aes_iv based on new_nonce and
// server_nonce as defined in "Creating an Authorization Key".
func TempAESKeys(newNonce, serverNonce *big.Int) (key, iv []byte) {
	// See https://core.telegram.org/mtproto/auth_key#presenting-proof-of-work-server-authentication
	// 5. Server responds in one of two ways: [...]

	// tmp_aes_key := SHA1(new_nonce + server_nonce) + substr (SHA1(server_nonce + new_nonce), 0, 12);
	// SHA1(new_nonce + server_nonce)
	key = append(key, sha1BigInt(newNonce, serverNonce)...)
	// substr (SHA1(server_nonce + new_nonce), 0, 12);
	key = append(key, sha1BigInt(serverNonce, newNonce)[:12]...)

	// tmp_aes_iv := substr (SHA1(server_nonce + new_nonce), 12, 8) + SHA1(new_nonce + new_nonce) + substr (new_nonce, 0, 4);
	iv = append(iv, sha1BigInt(serverNonce, newNonce)[12:12+8]...)
	iv = append(iv, sha1BigInt(newNonce, newNonce)...)
	iv = append(iv, newNonce.Bytes()[:4]...)

	return key, iv
}
