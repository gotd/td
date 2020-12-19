package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/proto/codec"
	"github.com/gotd/td/tg"
)

func (c *Client) handleSessionCreated(b *bin.Buffer) error {
	var ns mt.NewSessionCreated
	if err := ns.Decode(b); err != nil {
		return xerrors.Errorf("failed to decode: %x", err)
	}

	atomic.StoreInt64(&c.salt, ns.ServerSalt)

	if err := c.saveSession(c.ctx); err != nil {
		return xerrors.Errorf("failed to save session: %w", err)
	}

	c.log.Info("Session created")

	return nil
}

func (c *Client) handleUnknown(b *bin.Buffer) error {
	// Can't process unknown type.
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	c.log.With(
		zap.String("type_id", fmt.Sprintf("0x%x", id)),
	).Warn("Unknown type id")

	return nil
}

func (c *Client) handleMessage(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		// Empty body.
		return xerrors.Errorf("failed to determine message type: %w", err)
	}
	typeStr := "unknown"
	if s := c.types.Get(id); s != "" {
		typeStr = s
	}
	c.log.With(
		zap.String("type_id", fmt.Sprintf("0x%x", id)),
		zap.String("type_str", typeStr),
	).Debug("HandleMessage")
	switch id {
	case mt.BadMsgNotificationTypeID, mt.BadServerSaltTypeID:
		return c.handleBadMsg(b)
	case proto.MessageContainerTypeID:
		return c.processContainer(b)
	case mt.NewSessionCreatedTypeID:
		return c.handleSessionCreated(b)
	case proto.ResultTypeID:
		return c.handleResult(b)
	case mt.PongTypeID:
		return c.handlePong(b)
	case mt.MsgsAckTypeID:
		return c.handleAck(b)
	case proto.GZIPTypeID:
		return c.handleGZIP(b)
	// Handling all types of tg.UpdatesClass.
	case tg.UpdatesTooLongTypeID,
		tg.UpdateShortMessageTypeID,
		tg.UpdateShortChatMessageTypeID,
		tg.UpdateShortTypeID,
		tg.UpdatesCombinedTypeID,
		tg.UpdatesTypeID,
		tg.UpdateShortSentMessageTypeID:
		return c.handleUpdates(b)
	default:
		return c.handleUnknown(b)
	}
}

func (c *Client) processContainer(b *bin.Buffer) error {
	var container proto.MessageContainer
	if err := container.Decode(b); err != nil {
		return xerrors.Errorf("container: %w", err)
	}
	for _, msg := range container.Messages {
		if err := c.processContainerMessage(msg); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) processContainerMessage(msg proto.Message) error {
	b := &bin.Buffer{Buf: msg.Body}
	return c.handleMessage(b)
}

func (c *Client) read(ctx context.Context, b *bin.Buffer) (*crypto.EncryptedMessageData, error) {
	b.Reset()
	if err := c.conn.Recv(ctx, b); err != nil {
		return nil, err
	}

	msg, err := c.cipher.DecryptDataFrom(c.authKey, atomic.LoadInt64(&c.session), b)
	if err != nil {
		return nil, xerrors.Errorf("decrypt: %w", err)
	}

	return msg, nil
}

func (c *Client) readLoop(ctx context.Context) {
	b := new(bin.Buffer)
	log := c.log.Named("read")
	log.Debug("Read loop started")

	c.wg.Add(1)
	defer c.wg.Done()

	for {
		msg, err := c.read(ctx, b)
		if err == nil {
			// Reading ok.
			go func() { _ = c.handleEncryptedMessage(msg) }()

			continue
		}

		if errors.Is(err, io.EOF) {
			// Nothing was received.
			// TODO(ernado): also check ctx done
			if err := c.reconnect(); err != nil {
				c.log.With(zap.Error(err)).Error("Failed to reconnect")
			}
			continue
		}

		// Checking for read timeout.
		var syscall *net.OpError
		if errors.As(err, &syscall) && syscall.Timeout() {
			// We call SetReadDeadline so such error is expected.
			c.log.Debug("No updates")
			continue
		}

		var protoErr *codec.ProtocolErr
		if errors.As(err, &protoErr) && protoErr.Code == codec.CodeAuthKeyNotFound {
			c.log.Warn("Re-generating keys (server not found key that we provided)")
			if err := c.createAuthKey(ctx); err != nil {
				// Probably fatal error.
				c.log.With(zap.Error(err)).Error("Unable to create auth key")
			}

			c.log.Info("Created auth keys")
			continue
		}

		select {
		case <-ctx.Done():
			c.log.Debug("Read loop done (closing)")
			return
		default:
			// Notifying about possible errors.
			log.With(zap.Error(err)).Error("Read returned error")
		}
	}
}

func (c *Client) handleEncryptedMessage(msg *crypto.EncryptedMessageData) error {
	// TODO(ccln): we can avoid this buffer allocation
	// by re-using readLoop buffer for decoding message
	// and handle decoded data in separeted goroutine
	b := new(bin.Buffer)
	b.ResetTo(msg.Data())

	if err := c.handleMessage(b); err != nil {
		c.log.Error("handle", zap.Error(err))
		return err
	}

	needAck := (msg.SeqNo & 0x01) != 0
	if needAck {
		c.ackSendChan <- msg.MessageID
	}

	return nil
}
