package crypto

import (
	"crypto/rsa"
	"crypto/sha1" // #nosec
	"encoding/binary"
	"math/big"

	"github.com/gotd/td/bin"
)

// RSAFingerprint returns fingerprint of RSA public key as defined in MTProto.
func RSAFingerprint(key *rsa.PublicKey) int64 {
	e := big.NewInt(int64(key.E))

	// See "Creating an Authorization Key" for reference:
	// * https://core.telegram.org/mtproto/auth_key#dh-exchange-initiation
	// rsa_public_key n:string e:string = RSAPublicKey
	buf := new(bin.Buffer)
	buf.PutBytes(key.N.Bytes())
	buf.PutBytes(e.Bytes())

	h := sha1.New() // #nosec
	_, _ = h.Write(buf.Buf)
	result := h.Sum(nil)[12:sha1.Size]
	return int64(binary.LittleEndian.Uint64(result))
}
