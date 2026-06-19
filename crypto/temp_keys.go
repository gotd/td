package crypto

import (
	"crypto/sha1" // #nosec
	"math/big"
)

// sha1Bytes returns SHA1(a + b).
func sha1Bytes(a, b []byte) []byte {
	buf := make([]byte, 0, len(a)+len(b))
	buf = append(buf, a...)
	buf = append(buf, b...)

	h := sha1.Sum(buf) // #nosec
	return h[:]
}

// TempAESKeys returns tmp_aes_key and tmp_aes_iv based on new_nonce and
// server_nonce as defined in "Creating an Authorization Key".
//
// new_nonce (int256) and server_nonce (int128) are fixed-length byte strings
// of 32 and 16 bytes. They must be hashed at their full length: see
// https://core.telegram.org/mtproto/auth_key#presenting-proof-of-work-server-authentication
//
// They are passed here as *big.Int, so they are first serialized back to their
// canonical fixed length with FillBytes. Using big.Int.Bytes() instead would
// strip leading zero bytes, and whenever new_nonce or server_nonce starts with
// 0x00 (about 1/256 of the time each) the SHA1 input would be a byte short and
// the derived tmp_aes_key/tmp_aes_iv would be wrong, making the server_DH_params
// answer impossible to decrypt for a spec-compliant peer.
func TempAESKeys(newNonce, serverNonce *big.Int) (key, iv []byte) {
	nn := make([]byte, 32)
	newNonce.FillBytes(nn)
	sn := make([]byte, 16)
	serverNonce.FillBytes(sn)

	// tmp_aes_key := SHA1(new_nonce + server_nonce) + substr(SHA1(server_nonce + new_nonce), 0, 12)
	key = append(key, sha1Bytes(nn, sn)...)
	key = append(key, sha1Bytes(sn, nn)[:12]...)

	// tmp_aes_iv := substr(SHA1(server_nonce + new_nonce), 12, 8) + SHA1(new_nonce + new_nonce) + substr(new_nonce, 0, 4)
	iv = append(iv, sha1Bytes(sn, nn)[12:12+8]...)
	iv = append(iv, sha1Bytes(nn, nn)...)
	iv = append(iv, nn[:4]...)

	return key, iv
}
