// +build fuzz

package crypto

import (
	"bytes"
	"crypto/rsa"
	"math/rand"
)

func FuzzRSA(data []byte) int {
	source := rand.NewSource(0)
	if len(data) > rsaDataLen {
		data = data[:rsaDataLen]
	}
	reader := rand.New(source)
	k, err := rsa.GenerateKey(reader, 2048)
	if err != nil {
		panic(err)
	}

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
