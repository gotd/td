//go:build fuzz
// +build fuzz

package crypto

import (
	"bytes"
	"math/rand"

	"github.com/nnqq/td/internal/testutil"
)

func FuzzRSA(data []byte) int {
	source := rand.NewSource(0)
	if len(data) > rsaDataLen {
		data = data[:rsaDataLen]
	}
	reader := rand.New(source)
	k := testutil.RSAPrivateKey()

	encrypted, err := RSAEncryptHashed(data, &k.PublicKey, reader)
	if err != nil {
		panic(err)
	}

	decrypted, err := RSADecryptHashed(encrypted, k)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(data, decrypted) {
		panic("mismatch")
	}

	return 1
}
