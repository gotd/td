package crypto

import (
	"crypto/aes"
	"fmt"
	"io"

	"github.com/go-faster/errors"
	"github.com/gotd/ige"

	"github.com/gotd/td/bin"
)

// BindAuthKeyInnerTypeID is TL type id for bind_auth_key_inner.
const BindAuthKeyInnerTypeID = 0x75a3f765

// BindAuthKeyInner is a temporary auth key binding payload.
//
// bind_auth_key_inner#75a3f765 nonce:long temp_auth_key_id:long
//
//	perm_auth_key_id:long temp_session_id:long expires_at:int = BindAuthKeyInner
type BindAuthKeyInner struct {
	Nonce         int64
	TempAuthKeyID int64
	PermAuthKeyID int64
	TempSessionID int64
	ExpiresAt     int
}

// TypeID returns TL type id.
func (*BindAuthKeyInner) TypeID() uint32 {
	return BindAuthKeyInnerTypeID
}

// Encode implements bin.Encoder.
func (m *BindAuthKeyInner) Encode(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't encode bind_auth_key_inner#75a3f765 as nil")
	}
	b.PutID(BindAuthKeyInnerTypeID)
	b.PutLong(m.Nonce)
	b.PutLong(m.TempAuthKeyID)
	b.PutLong(m.PermAuthKeyID)
	b.PutLong(m.TempSessionID)
	b.PutInt(m.ExpiresAt)
	return nil
}

// Decode implements bin.Decoder.
func (m *BindAuthKeyInner) Decode(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't decode bind_auth_key_inner#75a3f765 to nil")
	}
	if err := b.ConsumeID(BindAuthKeyInnerTypeID); err != nil {
		return fmt.Errorf("unable to decode bind_auth_key_inner#75a3f765: %w", err)
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode bind_auth_key_inner#75a3f765: field nonce: %w", err)
		}
		m.Nonce = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode bind_auth_key_inner#75a3f765: field temp_auth_key_id: %w", err)
		}
		m.TempAuthKeyID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode bind_auth_key_inner#75a3f765: field perm_auth_key_id: %w", err)
		}
		m.PermAuthKeyID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode bind_auth_key_inner#75a3f765: field temp_session_id: %w", err)
		}
		m.TempSessionID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode bind_auth_key_inner#75a3f765: field expires_at: %w", err)
		}
		m.ExpiresAt = value
	}
	return nil
}

// EncryptBindMessage encrypts binding message using permanent key and MTProto v1 KDF.
//
// This helper exists specifically for auth.bindTempAuthKey flow where Telegram
// requires a legacy MTProto v1 envelope/KDF even if the transport session uses
// modern MTProto 2.0 packets for regular API traffic.
//
// Result format:
//
//	perm_auth_key_id(8) + msg_key(16) + encrypted_data
func EncryptBindMessage(rand io.Reader, permKey AuthKey, msgID int64, inner *BindAuthKeyInner) ([]byte, error) {
	if permKey.Zero() {
		return nil, errors.New("permanent key is zero")
	}
	if rand == nil {
		rand = DefaultRand()
	}

	payload := &bin.Buffer{}
	if err := payload.Encode(inner); err != nil {
		return nil, errors.Wrap(err, "encode bind_auth_key_inner")
	}

	// Binding encrypted message envelope:
	// random:int128 + msg_id:long + seq_no:int + msg_len:int + message + padding.
	//
	// Quote (Special binding message): "msg_key = substr(sha1(message_data), 4, 16)."
	// Link: https://core.telegram.org/api/pfs#special-binding-message
	//
	// Here message_data is the plaintext envelope before random padding.
	plaintext := &bin.Buffer{}
	random := make([]byte, 16)
	if _, err := io.ReadFull(rand, random); err != nil {
		return nil, errors.Wrap(err, "generate random prefix")
	}
	plaintext.Put(random)
	plaintext.PutLong(msgID)
	plaintext.PutInt32(0)
	plaintext.PutInt32(int32(payload.Len()))
	plaintext.Put(payload.Buf)

	// Important: for binding message msg_key is calculated from the envelope
	// without random padding, so we do it before alignment bytes are appended.
	msgKey := MessageKeyV1(plaintext.Buf)

	if rem := len(plaintext.Buf) % aes.BlockSize; rem != 0 {
		paddingLen := aes.BlockSize - rem
		offset := len(plaintext.Buf)
		plaintext.Buf = append(plaintext.Buf, make([]byte, paddingLen)...)
		if _, err := io.ReadFull(rand, plaintext.Buf[offset:]); err != nil {
			return nil, errors.Wrap(err, "generate random padding")
		}
	}

	// Binding message uses KDF v1 with permanent key material.
	key, iv := KeysV1(permKey.Value, msgKey)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, errors.Wrap(err, "create aes cipher")
	}

	// Quote (Method docs): "prepend perm_auth_key_id and msg_key as usual".
	// Link: https://core.telegram.org/method/auth.bindTempAuthKey
	//
	// For this method, encrypted_message includes auth_key_id + msg_key + encrypted_data.
	msg := EncryptedMessage{
		AuthKeyID:     permKey.ID,
		MsgKey:        msgKey,
		EncryptedData: make([]byte, len(plaintext.Buf)),
	}
	ige.EncryptBlocks(block, iv[:], msg.EncryptedData, plaintext.Buf)

	encrypted := &bin.Buffer{}
	if err := msg.Encode(encrypted); err != nil {
		return nil, errors.Wrap(err, "encode encrypted message")
	}
	return append([]byte(nil), encrypted.Buf...), nil
}
