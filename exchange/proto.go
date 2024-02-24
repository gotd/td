package exchange

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/transport"
)

type unencryptedWriter struct {
	clock   clock.Clock
	conn    transport.Conn
	timeout time.Duration
	input   proto.MessageType
	output  proto.MessageType
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

	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()
	return w.conn.Send(ctx, b)
}

func (w unencryptedWriter) tryRead(ctx context.Context, b *bin.Buffer) error {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	if err := w.conn.Recv(ctx, b); err != nil {
		return err
	}

	return nil
}

func (w unencryptedWriter) isClient() bool {
	return w.output == proto.MessageFromClient
}

func (w unencryptedWriter) readUnencrypted(ctx context.Context, b *bin.Buffer, data bin.Decoder) error {
	b.Reset()

	for {
		if err := w.tryRead(ctx, b); err != nil {
			var protocolErr *codec.ProtocolErr
			if w.isClient() &&
				errors.As(err, &protocolErr) &&
				protocolErr.Code == codec.CodeAuthKeyNotFound {
				continue
			}
			return err
		}

		break
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
		return errors.New("bad msg type")
	}
	return nil
}
