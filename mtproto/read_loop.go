package mtproto

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/proto/codec"
)

func (c *Conn) handleSessionCreated(b *bin.Buffer) error {
	var ns mt.NewSessionCreated
	if err := ns.Decode(b); err != nil {
		return xerrors.Errorf("failed to decode: %x", err)
	}

	atomic.StoreInt64(&c.salt, ns.ServerSalt)
	c.sessionCreated.Done()

	if err := c.saveSession(c.ctx); err != nil {
		return xerrors.Errorf("failed to save session: %w", err)
	}

	c.log.Info("Session created")

	return nil
}

func (c *Conn) handleUnknown(b *bin.Buffer) error {
	if c.handler != nil {
		return c.handler(b)
	}
	return nil
}

func (c *Conn) handleMessage(b *bin.Buffer) error {
	c.trace.OnMessage(b)

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
	default:
		return c.handleUnknown(b)
	}
}

func (c *Conn) processContainer(b *bin.Buffer) error {
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

func (c *Conn) processContainerMessage(msg proto.Message) error {
	b := &bin.Buffer{Buf: msg.Body}
	return c.handleMessage(b)
}

func (c *Conn) read(ctx context.Context, b *bin.Buffer) (*crypto.EncryptedMessageData, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	b.Reset()
	if err := c.conn.Recv(ctx, b); err != nil {
		return nil, err
	}

	msg, err := c.cipher.DecryptFromBuffer(c.authKey, b)
	if err != nil {
		return nil, xerrors.Errorf("decrypt: %w", err)
	}

	if msg.SessionID != atomic.LoadInt64(&c.session) {
		return nil, xerrors.Errorf("invalid session")
	}

	return msg, nil
}

func (c *Conn) isDone() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func (c *Conn) readLoop(ctx context.Context) {
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

		if shouldReconnect(err) {
			if c.isDone() {
				// Ignoring connection error after client close.
				return
			}
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

		if c.isDone() {
			c.log.Debug("Read loop done (closing)")
			return
		}

		// Notifying about unhandled errors.
		log.With(zap.Error(err)).Error("Read returned error")
	}
}

func (c *Conn) handleEncryptedMessage(msg *crypto.EncryptedMessageData) error {
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
