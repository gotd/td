package calls

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// IncomingCall is a pending call request from a peer. Answer it with Accept or
// turn it down with Reject.
type IncomingCall struct {
	client *Client
	req    *tg.PhoneCallRequested
}

// UserID returns the caller's user ID.
func (ic *IncomingCall) UserID() int64 { return ic.req.AdminID }

// Video reports whether the caller requested a video call.
func (ic *IncomingCall) Video() bool { return ic.req.Video }

// Reject declines the call as busy.
func (ic *IncomingCall) Reject(ctx context.Context) error {
	c := ic.client
	cl := newCallState(false)
	cl.input = tg.InputPhoneCall{ID: ic.req.ID, AccessHash: ic.req.AccessHash}
	c.setActive(cl)
	return c.Discard(ctx, DiscardBusy)
}

// Accept answers the call, performing the DH exchange and returning the media
// connection once the signaling handshake completes.
func (ic *IncomingCall) Accept(ctx context.Context) (*Conn, error) {
	c := ic.client

	dh, serverRandom, err := getDHConfig(ctx, c.api)
	if err != nil {
		return nil, err
	}
	b, gBInt, err := dh.randomExp(c.rand, serverRandom)
	if err != nil {
		return nil, err
	}
	gB := pad(gBInt)

	cl := newCallState(false)
	cl.dh = dh
	cl.secret = b
	cl.gAHash = ic.req.GAHash
	cl.input = tg.InputPhoneCall{ID: ic.req.ID, AccessHash: ic.req.AccessHash}
	c.setActive(cl)

	if _, err := c.api.PhoneAcceptCall(ctx, &tg.PhoneAcceptCallRequest{
		Peer:     cl.input,
		GB:       gB,
		Protocol: acceptProtocol(ic.req.Protocol),
	}); err != nil {
		c.clearActive(cl)
		return nil, errors.Wrap(err, "accept call")
	}

	var obj *tg.PhoneCall
	select {
	case obj = <-cl.confirmed:
	case d := <-cl.discarded:
		c.clearActive(cl)
		return nil, discardedError(d)
	case <-ctx.Done():
		c.clearActive(cl)
		return nil, ctx.Err()
	}

	// Verify the caller's g_a matches the commitment hash sent in the request.
	if !bytes.Equal(gAHash(obj.GAOrB), ic.req.GAHash) {
		c.clearActive(cl)
		return nil, errors.New("g_a hash commitment mismatch")
	}
	key, _, err := dh.computeKey(obj.GAOrB, b)
	if err != nil {
		c.clearActive(cl)
		return nil, err
	}
	c.setInput(cl, tg.InputPhoneCall{ID: obj.ID, AccessHash: obj.AccessHash})

	return c.startCall(cl, key, obj.Connections)
}

func randInt32() int32 {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	return int32(binary.LittleEndian.Uint32(buf[:]) & 0x7fffffff)
}
