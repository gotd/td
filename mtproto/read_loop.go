package mtproto

import (
	"context"
	"errors"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/proto/codec"
)

func (c *Conn) handleMessage(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		// Empty body.
		return xerrors.Errorf("failed to determine message type: %w", err)
	}

	c.logWithType(b).Debug("Handle message")
	switch id {
	case mt.NewSessionCreatedTypeID:
		return c.handleSessionCreated(b)
	case mt.BadMsgNotificationTypeID, mt.BadServerSaltTypeID:
		return c.handleBadMsg(b)
	case proto.MessageContainerTypeID:
		return c.processContainer(b)
	case proto.ResultTypeID:
		return c.handleResult(b)
	case mt.PongTypeID:
		return c.handlePong(b)
	case mt.MsgsAckTypeID:
		return c.handleAck(b)
	case proto.GZIPTypeID:
		return c.handleGZIP(b)
	default:
		return c.handler.OnMessage(b)
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

// errRejected is returned on invalid message that should not be processed.
var errRejected = errors.New("message rejected")

func checkMessageID(now time.Time, rawID int64) error {
	id := proto.MessageID(rawID)

	// Check that message is from server.
	switch id.Type() {
	case proto.MessageFromServer, proto.MessageServerResponse:
		// Valid.
	default:
		return xerrors.Errorf("unexpected type %s: %w", id.Type(), errRejected)
	}

	// https://core.telegram.org/mtproto/description#message-identifier-msg-id
	// A message is rejected over 300 seconds after it is created or 30 seconds
	// before it is created (this is needed to protect from replay attacks).
	const (
		maxPast   = time.Second * 300
		maxFuture = time.Second * 30
	)
	created := id.Time()
	if created.Before(now) && now.Sub(created) > maxPast {
		return xerrors.Errorf("created too far in past: %w", errRejected)
	}
	if created.Sub(now) > maxFuture {
		return xerrors.Errorf("created too far in future: %w", errRejected)
	}

	return nil
}

func (c *Conn) read(ctx context.Context, b *bin.Buffer) (*crypto.EncryptedMessageData, error) {
	b.Reset()
	if err := c.conn.Recv(ctx, b); err != nil {
		return nil, err
	}

	session := c.session()
	msg, err := c.cipher.DecryptFromBuffer(session.Key, b)
	if err != nil {
		return nil, xerrors.Errorf("decrypt: %w", err)
	}

	// Validating message. This protects from replay attacks.
	if msg.SessionID != session.ID {
		return nil, xerrors.Errorf("invalid session: %w", errRejected)
	}
	if err := checkMessageID(c.clock.Now(), msg.MessageID); err != nil {
		return nil, xerrors.Errorf("bad message id: %w", err)
	}

	return msg, nil
}

func (c *Conn) noUpdates(err error) bool {
	// Checking for read timeout.
	var syscall *net.OpError
	if errors.As(err, &syscall) && syscall.Timeout() {
		// We call SetReadDeadline so such error is expected.
		c.log.Debug("No updates")
		return true
	}
	return false
}

func (c *Conn) handleAuthKeyNotFound(ctx context.Context) error {
	if c.session().ID == 0 {
		// The 404 error can also be caused by zero session id.
		// See https://github.com/gotd/td/issues/107
		//
		// We should recover from this in createAuthKey, but in general
		// this code branch should be unreachable.
		c.log.Warn("BUG: zero session id found")
	}
	c.log.Warn("Re-generating keys (server not found key that we provided)")
	if err := c.createAuthKey(ctx); err != nil {
		return xerrors.Errorf("unable to create auth key: %w", err)
	}
	c.log.Info("Re-created auth keys")
	// Request will be retried by ack loop.
	// Probably we can speed-up this.
	return nil
}

func (c *Conn) readLoop(ctx context.Context) error {
	b := new(bin.Buffer)
	log := c.log.Named("read")
	log.Debug("Read loop started")
	defer log.Debug("Read loop done")
	defer close(c.messages)

	for {
		msg, err := c.read(ctx, b)
		if errors.Is(err, errRejected) {
			c.log.Warn("Ignoring rejected message", zap.Error(err))
			continue
		}
		if err == nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case c.messages <- msg:
				continue
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if c.noUpdates(err) {
				continue
			}
		}

		var protoErr *codec.ProtocolErr
		if errors.As(err, &protoErr) && protoErr.Code == codec.CodeAuthKeyNotFound {
			if err := c.handleAuthKeyNotFound(ctx); err != nil {
				return xerrors.Errorf("auth key not found: %w", err)
			}
			continue
		}

		select {
		case <-ctx.Done():
			return xerrors.Errorf("read loop: %w", ctx.Err())
		default:
			return xerrors.Errorf("read: %w", err)
		}
	}
}

func (c *Conn) readEncryptedMessages() error {
	b := new(bin.Buffer)
	for msg := range c.messages {
		b.ResetTo(msg.Data())

		if err := c.handleMessage(b); err != nil {
			// Probably we can return here, but this will shutdown whole
			// connection which can be unexpected.
			c.log.Warn("Error while handling message", zap.Error(err))
			// Sending acknowledge even on error. Client should restore
			// from missing updates via explicit pts check and getDiff call.
		}

		needAck := (msg.SeqNo & 0x01) != 0
		if needAck {
			c.ackSendChan <- msg.MessageID
		}
	}

	return nil
}
