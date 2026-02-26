package crypto

import (
	"bytes"
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/ige"

	"github.com/gotd/td/bin"
)

func TestEncryptBindMessage(t *testing.T) {
	a := require.New(t)

	// Use deterministic key material to validate exact field placement and KDF.
	var raw Key
	for i := range raw {
		raw[i] = byte(255 - i)
	}
	permKey := raw.WithID()

	msgID := int64(0x011223344556677)
	sessionID := int64(0x010203040506070)
	inner := &BindAuthKeyInner{
		Nonce:         int64(0x0001020304050607),
		TempAuthKeyID: int64(0x0011223344556677),
		PermAuthKeyID: permKey.IntID(),
		TempSessionID: sessionID,
		ExpiresAt:     1735689600,
	}

	encrypted, err := EncryptBindMessage(
		// Fixed random stream makes output reproducible for assertions.
		bytes.NewReader(bytes.Repeat([]byte{0xCD}, 64)),
		permKey,
		msgID,
		inner,
	)
	a.NoError(err)

	var msg EncryptedMessage
	a.NoError(msg.Decode(&bin.Buffer{Buf: encrypted}))
	// Quote (Method docs): "prepend perm_auth_key_id and msg_key as usual".
	// Link: https://core.telegram.org/method/auth.bindTempAuthKey
	a.Equal(permKey.ID, msg.AuthKeyID)
	a.Equal(0, len(msg.EncryptedData)%aes.BlockSize)

	key, iv := KeysV1(permKey.Value, msg.MsgKey)
	block, err := aes.NewCipher(key[:])
	a.NoError(err)

	plaintext := make([]byte, len(msg.EncryptedData))
	ige.DecryptBlocks(block, iv[:], plaintext, msg.EncryptedData)
	// 16(random) + 8(msg_id) + 4(seq_no) + 4(len) + 40(bind_auth_key_inner)
	prefixAndBodyLen := 16 + 8 + 4 + 4 + 40
	a.GreaterOrEqual(len(plaintext), prefixAndBodyLen)
	// Quote (Method docs): "Compute msg_key as SHA1 of the resulting string, substring(4, 16)."
	// Link: https://core.telegram.org/method/auth.bindTempAuthKey
	a.Equal(msg.MsgKey, MessageKeyV1(plaintext[:prefixAndBodyLen]))

	b := &bin.Buffer{Buf: plaintext}
	randomPrefix := make([]byte, 16)
	a.NoError(b.ConsumeN(randomPrefix, 16))
	a.NotEqual(make([]byte, 16), randomPrefix)
	gotMsgID, err := b.Long()
	a.NoError(err)
	a.Equal(msgID, gotMsgID)
	seqNo, err := b.Int32()
	a.NoError(err)
	a.Equal(int32(0), seqNo)
	msgLen, err := b.Int32()
	a.NoError(err)
	a.Equal(int32(40), msgLen)

	var got BindAuthKeyInner
	body := make([]byte, int(msgLen))
	a.NoError(b.ConsumeN(body, int(msgLen)))
	a.NoError(got.Decode(&bin.Buffer{Buf: body}))
	a.Equal(*inner, got)
}

func TestEncryptBindMessageZeroPermKey(t *testing.T) {
	a := require.New(t)

	_, err := EncryptBindMessage(bytes.NewReader(nil), AuthKey{}, 2, &BindAuthKeyInner{})
	a.Error(err)
}
