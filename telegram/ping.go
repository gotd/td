package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/mt"
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

// nolint:gocyclo // TODO(ernado): refactor
func (c Client) Ping(ctx context.Context) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(&pingMessage{id: 0xafef}, b); err != nil {
		return err
	}
	if err := proto.WriteIntermediate(c.conn, b); err != nil {
		return xerrors.Errorf("failed to write: %w", err)
	}

	b.Reset()
	if err := proto.ReadIntermediate(c.conn, b); err != nil {
		return xerrors.Errorf("failed to read: %w", err)
	}
	if err := c.checkProtocolError(b); err != nil {
		return xerrors.Errorf("protocol error: %w", err)
	}

	// Decrypting response.
	encMessage := &proto.EncryptedMessage{}
	if err := encMessage.Decode(b); err != nil {
		return err
	}
	msg, err := c.decryptData(encMessage)
	if err != nil {
		return err
	}
	b.ResetTo(msg.MessageDataWithPadding[:msg.MessageDataLen])

	// Checking if response is RPC error.
	if err := c.checkRPCError(b); err != nil {
		return err
	}

	id, err := b.PeekID()
	if err != nil {
		return err
	}
	var gotPong bool

	switch id {
	case mt.PongTypeID:
		var pong mt.Pong
		if err := pong.Decode(b); err != nil {
			return err
		}
		gotPong = true
	case proto.MessageContainerTypeID:
		var batch proto.MessageContainer
		if err := batch.Decode(b); err != nil {
			return xerrors.Errorf("failed to read message container: %w", err)
		}
		for _, msg := range batch.Messages {
			b.ResetTo(msg.Body)
			msgID, err := b.PeekID()
			if err != nil {
				return err
			}
			switch msgID {
			case mt.PongTypeID:
				var pong mt.Pong
				if err := pong.Decode(b); err != nil {
					return err
				}
				gotPong = true
			case mt.NewSessionCreatedTypeID:
				var ns mt.NewSessionCreated
				if err := ns.Decode(b); err != nil {
					return err
				}
				c.log.Info("New session created")
				c.salt = ns.ServerSalt
			}
		}
	default:
		return xerrors.Errorf("unexpected ping response id id %x", id)
	}
	if !gotPong {
		return xerrors.New("no pong message received")
	}
	return nil
}
