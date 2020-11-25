package proto

import "github.com/ernado/td/bin"

// EncryptedMessage of protocol.
type EncryptedMessage struct {
	AuthKeyID int64
	MsgKey    bin.Int128

	EncryptedData []byte
}

func (e *EncryptedMessage) Decode(b *bin.Buffer) error {
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		e.AuthKeyID = v
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

func (e EncryptedMessage) Encode(b *bin.Buffer) error {
	b.PutLong(e.AuthKeyID)
	b.PutInt128(e.MsgKey)
	b.Put(e.EncryptedData)
	return nil
}
