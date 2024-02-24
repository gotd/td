package mtproto

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/proto/codec"
)

// https://core.telegram.org/mtproto/description#message-identifier-msg-id
// A message is rejected over 300 seconds after it is created or 30 seconds
// before it is created (this is needed to protect from replay attacks).
const (
	maxPast   = time.Second * 300
	maxFuture = time.Second * 30
)

// errRejected is returned on invalid message that should not be processed.
var errRejected = errors.New("message rejected")

func checkMessageID(now time.Time, rawID int64) error {
	id := proto.MessageID(rawID)

	// Check that message is from server.
	switch id.Type() {
	case proto.MessageFromServer, proto.MessageServerResponse:
		// Valid.
	default:
		return errors.Wrapf(errRejected, "unexpected type %s", id.Type())
	}

	created := id.Time()
	if created.Before(now) && now.Sub(created) > maxPast {
		return errors.Wrap(errRejected, "created too far in past")
	}
	if created.Sub(now) > maxFuture {
		return errors.Wrap(errRejected, "created too far in future")
	}

	return nil
}

func (c *Conn) decryptMessage(b *bin.Buffer) (*crypto.EncryptedMessageData, error) {
	session := c.session()
	msg, err := c.cipher.DecryptFromBuffer(session.Key, b)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt")
	}

	// Validating message. This protects from replay attacks.
	if msg.SessionID != session.ID {
		return nil, errors.Wrapf(errRejected, "invalid session (got %d, expected %d)", msg.SessionID, session.ID)
	}
	if err := checkMessageID(c.clock.Now(), msg.MessageID); err != nil {
		return nil, errors.Wrapf(err, "bad message id %d", msg.MessageID)
	}
	if !c.messageIDBuf.Consume(msg.MessageID) {
		return nil, errors.Wrapf(errRejected, "duplicate or too low message id %d", msg.MessageID)
	}

	return msg, nil
}

func (c *Conn) consumeMessage(ctx context.Context, buf *bin.Buffer) error {
	msg, err := c.decryptMessage(buf)
	if errors.Is(err, errRejected) {
		c.log.Warn("Ignoring rejected message", zap.Error(err))
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "consume message")
	}

	if err := c.handleMessage(msg.MessageID, &bin.Buffer{Buf: msg.Data()}); err != nil {
		// Probably we can return here, but this will shutdown whole
		// connection which can be unexpected.
		c.log.Warn("Error while handling message", zap.Error(err))
		// Sending acknowledge even on error. Client should restore
		// from missing updates via explicit pts check and getDiff call.
	}

	needAck := (msg.SeqNo & 0x01) != 0
	if needAck {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c.ackSendChan <- msg.MessageID:
		}
	}

	return nil
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
		return errors.Wrap(err, "unable to create auth key")
	}
	c.log.Info("Re-created auth keys")
	// Request will be retried by ack loop.
	// Probably we can speed-up this.
	return nil
}

func (c *Conn) readLoop(ctx context.Context) (err error) {
	log := c.log.Named("read")
	log.Debug("Read loop started")
	defer func() {
		l := log
		if err != nil {
			l = log.With(zap.NamedError("reason", err))
		}
		l.Debug("Read loop done")
	}()

	var (
		// Last error encountered by consumeMessage.
		lastErr atomic.Value
		// To wait all spawned goroutines
		handlers sync.WaitGroup
	)
	defer handlers.Wait()

	for {
		// We've tried multiple ways to reduce allocations via reusing buffer,
		// but naive implementation induces high idle memory waste.
		//
		// Proper optimization will probably require total rework of bin.Buffer
		// with sharded (by payload size?) pool that can be used after message
		// size read (after readLen).
		//
		// Such optimization can introduce additional complexity overhead and
		// is probably not worth it.
		buf := &bin.Buffer{}

		// Halting if consumeMessage encountered error.
		// Should be something critical with crypto.
		if err, ok := lastErr.Load().(error); ok && err != nil {
			return errors.Wrap(err, "halting")
		}

		if err := c.conn.Recv(ctx, buf); err != nil {
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
					return errors.Wrap(err, "auth key not found")
				}

				continue
			}

			select {
			case <-ctx.Done():
				return errors.Wrap(ctx.Err(), "read loop")
			default:
				return errors.Wrap(err, "read")
			}
		}

		handlers.Add(1)
		go func() {
			defer handlers.Done()

			// Spawning goroutine per incoming message to utilize as much
			// resources as possible while keeping idle utilization low.
			//
			// The "worker" model was replaced by this due to idle utilization
			// overhead, especially on multi-CPU systems with multiple running
			// clients.
			if err := c.consumeMessage(ctx, buf); err != nil {
				log.Error("Failed to process message", zap.Error(err))
				lastErr.Store(errors.Wrap(err, "consume"))
			}
		}()
	}
}
