package crypto

import "github.com/gotd/td/bin"

// EncryptedMessage of protocol.
type EncryptedMessage struct {
	AuthKeyID [8]byte
	MsgKey    bin.Int128

	EncryptedData []byte
}

// Decode implements bin.Decoder.
func (e *EncryptedMessage) Decode(b *bin.Buffer) error {
	if err := b.ConsumeN(e.AuthKeyID[:], 8); err != nil {
		return err
	}
	{
		v, err := b.Int128()
		if err != nil {
			return err
		}
		e.MsgKey = v
	}
	// Consuming the rest of the buffer.
	e.EncryptedData = append(e.EncryptedData[:0], make([]byte, b.Len())...)
	if err := b.ConsumeN(e.EncryptedData, b.Len()); err != nil {
		return err
	}
	return nil
}

// DecodeWithoutCopy is like Decode, but EncryptedData references to given buffer instead of
// copying.
func (e *EncryptedMessage) DecodeWithoutCopy(b *bin.Buffer) error {
	if err := b.ConsumeN(e.AuthKeyID[:], 8); err != nil {
		return err
	}
	{
		v, err := b.Int128()
		if err != nil {
			return err
		}
		e.MsgKey = v
	}
	// Consuming the rest of the buffer.
	e.EncryptedData = b.Buf
	return nil
}

// Encode implements bin.Encoder.
func (e EncryptedMessage) Encode(b *bin.Buffer) error {
	b.Put(e.AuthKeyID[:])
	b.PutInt128(e.MsgKey)
	b.Put(e.EncryptedData)
	return nil
}
