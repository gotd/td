package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/proto"
)

type pingMessage struct {
	id int64
}

func (p pingMessage) Encode(b *bin.Buffer) error {
	b.PutID(0x7abe77ec)
	b.PutLong(p.id)
	return nil
}

func (c Client) Ping(ctx context.Context) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(&pingMessage{id: 0xbad}, b); err != nil {
		return err
	}
	if err := proto.WriteIntermediate(c.conn, b); err != nil {
		return err
	}

	b.Reset()
	if err := proto.ReadIntermediate(c.conn, b); err != nil {
		return err
	}

	if b.Len() == 4 {
		// Protocol error?
		code, err := b.Int32()
		if err != nil {
			return err
		}
		code *= -1
		switch code {
		case proto.CodeAuthKeyNotFound:
			return xerrors.New("protocol error: auth key not found")
		case proto.CodeTransportFlood:
			return xerrors.New("protocol error: transport flood")
		default:
			return xerrors.Errorf("protocol erorr: code %d", code)
		}
	}

	encMessage := proto.EncryptedMessage{}
	if err := encMessage.Decode(b); err != nil {
		return err
	}

	return nil
}
