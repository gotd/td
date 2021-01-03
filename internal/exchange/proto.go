package exchange

import (
	"context"

	"github.com/gotd/td/internal/clock"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/transport"
)

type unencryptedWriter struct {
	clock  clock.Clock
	conn   transport.Conn
	input  proto.MessageType
	output proto.MessageType
}

func (w unencryptedWriter) writeUnencrypted(ctx context.Context, b *bin.Buffer, data bin.Encoder) error {
	b.Reset()

	if err := data.Encode(b); err != nil {
		return err
	}
	msg := proto.UnencryptedMessage{
		MessageID:   int64(proto.NewMessageID(w.clock.Now(), w.output)),
		MessageData: b.Copy(),
	}
	b.Reset()
	if err := msg.Encode(b); err != nil {
		return err
	}

	return w.conn.Send(ctx, b)
}

func (w unencryptedWriter) readUnencrypted(ctx context.Context, b *bin.Buffer, data bin.Decoder) error {
	b.Reset()
	if err := w.conn.Recv(ctx, b); err != nil {
		return err
	}

	var msg proto.UnencryptedMessage
	if err := msg.Decode(b); err != nil {
		return err
	}
	if err := w.checkMsgID(msg.MessageID); err != nil {
		return err
	}
	b.ResetTo(msg.MessageData)

	return data.Decode(b)
}

func (w unencryptedWriter) checkMsgID(id int64) error {
	if proto.MessageID(id).Type() != w.input {
		return xerrors.New("bad msg type")
	}
	return nil
}
