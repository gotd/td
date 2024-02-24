//go:build go1.18

package crypto

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/gotd/td/testutil"
)

func FuzzRSA(f *testing.F) {
	f.Add([]byte{1, 2, 3})
	f.Add([]byte{3, 2, 3, 1, 10})

	f.Fuzz(func(t *testing.T, data []byte) {
		source := rand.NewSource(0)
		if len(data) > rsaDataLen {
			data = data[:rsaDataLen]
		}
		reader := rand.New(source)
		k := testutil.RSAPrivateKey()

		encrypted, err := RSAEncryptHashed(data, &k.PublicKey, reader)
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := RSADecryptHashed(encrypted, k)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(data, decrypted) {
			t.Fatal("mismatch")
		}
	})
}
