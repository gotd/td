// +build go1.17

package crypto

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
)

func FuzzRSA(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		source := rand.NewSource(0)
		if len(data) > rsaDataLen {
			data = data[:rsaDataLen]
		}
		reader := rand.New(source)
		k := testutil.RSAPrivateKey()

		encrypted, err := RSAEncryptHashed(data, &k.PublicKey, reader)
		require.NoError(t, err)

		decrypted, err := RSADecryptHashed(encrypted, k)
		require.NoError(t, err)
		require.Equal(t, data, decrypted)
	})
}
