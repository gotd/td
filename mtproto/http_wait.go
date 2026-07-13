package mtproto

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
)

// httpWaiter is an optional capability of a transport.Conn that needs the
// mtproto layer to drive HTTP long-polling.
//
// The MTProto HTTP transport is request/response: to receive server-pushed
// messages while idle, the client must keep an http_wait service message in
// flight so the server has a response to hold open. http_wait is encrypted with
// the session key, so only the mtproto layer can produce it — the transport
// cannot. A transport implementing this interface (see the telegram/dcs HTTP
// resolver) is detected by Conn after the handshake and handed a factory that
// yields freshly-encrypted http_wait frames.
//
// See https://core.telegram.org/mtproto/transports#http-transport.
type httpWaiter interface {
	// HTTPWaitParams reports the desired http_wait fields (milliseconds).
	HTTPWaitParams() (maxDelay, waitAfter, maxWait int)
	// StartHTTPWait starts the transport long-poll loop, using frame to build
	// each encrypted http_wait message. It must be non-blocking.
	StartHTTPWait(frame func(ctx context.Context) (*bin.Buffer, error))
}

// startHTTPWait enables http_wait long-polling if the transport requires it.
// It is a no-op for transports that do not implement httpWaiter (TCP, WebSocket,
// MTProxy), which receive server pushes over a full-duplex socket instead.
func (c *Conn) startHTTPWait() {
	w, ok := c.conn.(httpWaiter)
	if !ok {
		return
	}
	maxDelay, waitAfter, maxWait := w.HTTPWaitParams()
	w.StartHTTPWait(func(ctx context.Context) (*bin.Buffer, error) {
		return c.encodeHTTPWait(ctx, maxDelay, waitAfter, maxWait)
	})
}

// encodeHTTPWait builds an encrypted http_wait service message into a fresh
// buffer. It mirrors write, but does not send: the transport sends the frame as
// a long-poll POST.
func (c *Conn) encodeHTTPWait(ctx context.Context, maxDelay, waitAfter, maxWait int) (*bin.Buffer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	msgID, seqNo := c.nextMsgSeq(false)

	// Guard the auth key against concurrent key regeneration, like write does.
	// During createAuthKey (exclusive Lock) this blocks, so no http_wait is sent
	// with a stale key; any long-poll started earlier delivers at most a -404,
	// which the exchange layer swallows (continue-on-ProtocolErr).
	c.exchangeLock.RLock()
	defer c.exchangeLock.RUnlock()

	b := &bin.Buffer{}
	if err := c.newEncryptedMessage(msgID, seqNo, &mt.HTTPWaitRequest{
		MaxDelay:  maxDelay,
		WaitAfter: waitAfter,
		MaxWait:   maxWait,
	}, b); err != nil {
		return nil, errors.Wrap(err, "encode http_wait")
	}
	return b, nil
}
