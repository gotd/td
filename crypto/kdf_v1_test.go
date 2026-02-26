package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func mustHex128(t *testing.T, h string) bin.Int128 {
	t.Helper()
	data, err := hex.DecodeString(h)
	require.NoError(t, err)
	require.Len(t, data, len(bin.Int128{}))

	var v bin.Int128
	copy(v[:], data)
	return v
}

func mustHex256(t *testing.T, h string) bin.Int256 {
	t.Helper()
	data, err := hex.DecodeString(h)
	require.NoError(t, err)
	require.Len(t, data, len(bin.Int256{}))

	var v bin.Int256
	copy(v[:], data)
	return v
}

func TestMessageKeyV1(t *testing.T) {
	a := require.New(t)
	// Guard against regressions in exact SHA1 slice position [4:20].
	a.Equal(
		mustHex128(t, "fc11a3669566de22b87f66550dc3de6d"),
		MessageKeyV1([]byte("bind-message-test-payload")),
	)
}

func TestKeysV1(t *testing.T) {
	a := require.New(t)

	var (
		authKey Key
		msgKey  bin.Int128
	)
	for i := range authKey {
		authKey[i] = byte(i)
	}
	for i := range msgKey {
		msgKey[i] = byte(i)
	}

	// Test vectors pin KDF composition byte-for-byte.
	key, iv := KeysV1(authKey, msgKey)
	a.Equal(mustHex256(t, "17d7295ca9213d1ab656acdb1ad48b2ea7f3a8f7095098d5508b900bbd5fccfc"), key)
	a.Equal(mustHex256(t, "2d7d16a65a84108e9805656caa474501cc580aa2edc33abfd0bfad785464d1c6"), iv)
}
