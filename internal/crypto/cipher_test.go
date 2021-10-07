package crypto

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
)

var testAuthKey = Key{
	93, 46, 125, 101, 244, 158, 194, 139, 208, 41, 168, 135, 97, 234, 39, 184, 164, 199,
	159, 18, 34, 101, 37, 68, 62, 125, 124, 89, 110, 243, 48, 53, 48, 219, 33, 7, 232, 154,
	169, 151, 199, 160, 22, 74, 182, 148, 24, 122, 222, 255, 21, 107, 214, 239, 113, 24,
	161, 150, 35, 71, 117, 60, 14, 126, 137, 160, 53, 75, 142, 195, 100, 249, 153, 126,
	113, 188, 105, 35, 251, 134, 232, 228, 52, 145, 224, 16, 96, 106, 108, 232, 69, 226,
	250, 1, 148, 9, 119, 239, 10, 163, 42, 223, 90, 151, 219, 246, 212, 40, 236, 4, 52,
	215, 23, 162, 211, 173, 25, 98, 44, 192, 88, 135, 100, 33, 19, 199, 150, 95, 251, 134,
	42, 62, 60, 203, 10, 185, 90, 221, 218, 87, 248, 146, 69, 219, 215, 107, 73, 35, 72,
	248, 233, 75, 213, 167, 192, 224, 184, 72, 8, 82, 60, 253, 30, 168, 11, 50, 254, 154,
	209, 152, 188, 46, 16, 63, 206, 183, 213, 36, 146, 236, 192, 39, 58, 40, 103, 75, 201,
	35, 238, 229, 146, 101, 171, 23, 160, 2, 223, 31, 74, 162, 197, 155, 129, 154, 94, 94,
	29, 16, 94, 193, 23, 51, 111, 92, 118, 198, 177, 135, 3, 125, 75, 66, 112, 206, 233,
	204, 33, 7, 29, 151, 233, 188, 162, 32, 198, 215, 176, 27, 153, 140, 242, 229, 205,
	185, 165, 14, 205, 161, 133, 42, 54, 230, 53, 105, 12, 142,
}.WithID()

func checkSame(t *testing.T, a, b Cipher) {
	asserts := require.New(t)

	sessionID, err := rand.Int(rand.Reader, big.NewInt(2345512351))
	asserts.NoError(err)

	msg := []byte("data")
	data := EncryptedMessageData{
		SessionID:              sessionID.Int64(),
		MessageDataLen:         int32(len(msg)),
		MessageDataWithPadding: msg,
	}

	var buf bin.Buffer
	err = a.Encrypt(testAuthKey, data, &buf)
	asserts.NoError(err)

	decrypt, err := b.DecryptFromBuffer(testAuthKey, &buf)
	asserts.NoError(err)

	asserts.Equal(data.SessionID, decrypt.SessionID)
	asserts.Equal(data.Data(), decrypt.Data())
}

func TestCipher(t *testing.T) {
	client := NewClientCipher(rand.Reader)
	server := NewServerCipher(rand.Reader)

	checkSame(t, client, server)
	checkSame(t, server, client)
}
