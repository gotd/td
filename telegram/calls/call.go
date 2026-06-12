package calls

import (
	"context"
	"io"
	"math/big"
	"sync"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/tg"
)

// Client places and answers Telegram 1:1 phone calls. It manages a single
// active call at a time and routes the relevant updates into it.
//
// Wire it to the update dispatcher with Register (or forward updates manually
// via Handle and HandleSignalingData).
type Client struct {
	api  *tg.Client
	log  log.Helper
	rand io.Reader

	onIncoming func(*IncomingCall)

	mu   sync.Mutex
	call *call // current call, if any
}

// call holds the mutable state of a single in-progress call.
type call struct {
	dh       *dhConfig
	secret   *big.Int
	gAHash   []byte
	input    tg.InputPhoneCall
	sig      *signalingEncryption
	conn     *Conn
	isCaller bool

	accepted  chan *tg.PhoneCallAccepted
	confirmed chan *tg.PhoneCall
	discarded chan *tg.PhoneCallDiscarded
}

// NewClient returns a Client bound to the given invoker.
func NewClient(api *tg.Client, opts Options) *Client {
	opts.setDefaults()
	return &Client{
		api:  api,
		log:  log.For(opts.Logger),
		rand: opts.Random,
	}
}

// OnIncoming registers a callback invoked when a peer requests a call.
func (c *Client) OnIncoming(fn func(*IncomingCall)) { c.onIncoming = fn }

// Register installs the call update handlers on the dispatcher.
func (c *Client) Register(d tg.UpdateDispatcher) {
	d.OnPhoneCall(func(ctx context.Context, _ tg.Entities, u *tg.UpdatePhoneCall) error {
		return c.Handle(ctx, u)
	})
	d.OnPhoneCallSignalingData(func(ctx context.Context, _ tg.Entities, u *tg.UpdatePhoneCallSignalingData) error {
		return c.HandleSignalingData(ctx, u)
	})
}

// Handle routes an updatePhoneCall into the active call.
func (c *Client) Handle(ctx context.Context, u *tg.UpdatePhoneCall) error {
	switch p := u.PhoneCall.(type) {
	case *tg.PhoneCallRequested:
		if c.onIncoming != nil {
			c.onIncoming(&IncomingCall{client: c, req: p})
		}
	case *tg.PhoneCallAccepted:
		c.deliver(func(cl *call) { trySend(cl.accepted, p) })
	case *tg.PhoneCall:
		c.deliver(func(cl *call) { trySend(cl.confirmed, p) })
	case *tg.PhoneCallDiscarded:
		c.onDiscarded(p)
	case *tg.PhoneCallWaiting:
		c.log.Debug(ctx, "Call waiting", log.Int64("id", p.ID))
	default:
		c.log.Debug(ctx, "Unhandled phone call update", log.String("type", u.PhoneCall.TypeName()))
	}
	return nil
}

// HandleSignalingData decrypts an incoming signaling packet and feeds it to the
// active call's media connection, then acknowledges it.
func (c *Client) HandleSignalingData(ctx context.Context, u *tg.UpdatePhoneCallSignalingData) error {
	c.mu.Lock()
	cl := c.call
	c.mu.Unlock()
	if cl == nil || cl.sig == nil || cl.conn == nil || cl.input.ID != u.PhoneCallID {
		c.log.Debug(ctx, "Dropping signaling data: no matching call")
		return nil
	}
	msgs, err := cl.sig.decryptMessages(u.Data)
	if err != nil {
		c.log.Debug(ctx, "Drop signaling", log.Error(err))
		return nil
	}
	for _, plain := range msgs {
		if err := cl.conn.onSignal(plain); err != nil {
			c.log.Warn(ctx, "Handle signaling", log.Error(err))
		}
	}
	c.flushAcks(ctx, cl)
	return nil
}

func (c *Client) onDiscarded(d *tg.PhoneCallDiscarded) {
	c.mu.Lock()
	cl := c.call
	c.mu.Unlock()
	if cl == nil {
		return
	}
	trySend(cl.discarded, d)
	if cl.conn != nil {
		_ = cl.conn.Close()
	}
}

// deliver runs fn against the active call under lock.
func (c *Client) deliver(fn func(*call)) {
	c.mu.Lock()
	cl := c.call
	c.mu.Unlock()
	if cl != nil {
		fn(cl)
	}
}

func (c *Client) flushAcks(ctx context.Context, cl *call) {
	seqs := cl.sig.drainAcks()
	if len(seqs) == 0 {
		return
	}
	ct, err := cl.sig.encryptAcks(seqs)
	if err != nil || ct == nil {
		return
	}
	if _, err := c.api.PhoneSendSignalingData(ctx, &tg.PhoneSendSignalingDataRequest{
		Peer: c.inputOf(cl),
		Data: ct,
	}); err != nil {
		c.log.Debug(ctx, "Send acks", log.Error(err))
	}
}

// inputOf returns the call's input peer under lock.
func (c *Client) inputOf(cl *call) tg.InputPhoneCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	return cl.input
}

// Request places an outgoing call to user. It performs the DH exchange and
// returns the media connection once the signaling handshake completes; the
// connection becomes usable when its OnConnected callback fires.
func (c *Client) Request(ctx context.Context, user tg.InputUserClass) (*Conn, error) {
	dh, serverRandom, err := getDHConfig(ctx, c.api)
	if err != nil {
		return nil, err
	}
	a, gAInt, err := dh.randomExp(c.rand, serverRandom)
	if err != nil {
		return nil, err
	}
	gA := pad(gAInt)
	hash := gAHash(gA)

	cl := newCallState(true)
	cl.dh = dh
	cl.secret = a
	cl.gAHash = hash
	c.setActive(cl)

	res, err := c.api.PhoneRequestCall(ctx, &tg.PhoneRequestCallRequest{
		UserID:   user,
		RandomID: int(randInt32()),
		GAHash:   hash,
		Protocol: callProtocol(),
	})
	if err != nil {
		c.clearActive(cl)
		return nil, errors.Wrap(err, "request call")
	}
	waiting, ok := res.PhoneCall.(*tg.PhoneCallWaiting)
	if !ok {
		c.clearActive(cl)
		return nil, errors.Errorf("unexpected request-call response %T", res.PhoneCall)
	}
	c.setInput(cl, tg.InputPhoneCall{ID: waiting.ID, AccessHash: waiting.AccessHash})

	var accepted *tg.PhoneCallAccepted
	select {
	case accepted = <-cl.accepted:
	case d := <-cl.discarded:
		c.clearActive(cl)
		return nil, discardedError(d)
	case <-ctx.Done():
		c.clearActive(cl)
		return nil, ctx.Err()
	}

	key, fingerprint, err := dh.computeKey(accepted.GB, a)
	if err != nil {
		c.clearActive(cl)
		return nil, err
	}
	c.setInput(cl, tg.InputPhoneCall{ID: accepted.ID, AccessHash: accepted.AccessHash})

	confirm, err := c.api.PhoneConfirmCall(ctx, &tg.PhoneConfirmCallRequest{
		Peer:           cl.input,
		GA:             gA,
		KeyFingerprint: fingerprint,
		Protocol:       callProtocol(),
	})
	if err != nil {
		c.clearActive(cl)
		return nil, errors.Wrap(err, "confirm call")
	}
	obj, ok := confirm.PhoneCall.(*tg.PhoneCall)
	if !ok {
		c.clearActive(cl)
		return nil, errors.Errorf("unexpected confirm-call response %T", confirm.PhoneCall)
	}

	return c.startCall(cl, key, obj.Connections)
}

// startCall sets up the media connection and signaling, opens the transport and
// starts the handshake.
func (c *Client) startCall(cl *call, key []byte, conns []tg.PhoneConnectionClass) (*Conn, error) {
	ctx := context.Background()
	conn := newConn(cl.isCaller, c.log)
	sig := newSignalingEncryption(key, cl.isCaller)

	conn.emit = func(payload []byte) {
		ct, err := sig.encryptMessage(payload)
		if err != nil {
			c.log.Warn(ctx, "Encrypt signaling", log.Error(err))
			return
		}
		c.mu.Lock()
		input := cl.input
		c.mu.Unlock()
		if _, err := c.api.PhoneSendSignalingData(context.Background(), &tg.PhoneSendSignalingDataRequest{
			Peer: input,
			Data: ct,
		}); err != nil {
			c.log.Warn(ctx, "Send signaling", log.Error(err))
		}
	}

	c.mu.Lock()
	cl.conn = conn
	cl.sig = sig
	c.mu.Unlock()

	if err := conn.open(conns); err != nil {
		c.clearActive(cl)
		return nil, errors.Wrap(err, "open transport")
	}
	if err := conn.start(); err != nil {
		c.clearActive(cl)
		return nil, errors.Wrap(err, "start transport")
	}
	return conn, nil
}

// Discard ends the active call with the given reason.
func (c *Client) Discard(ctx context.Context, reason DiscardReason) error {
	c.mu.Lock()
	cl := c.call
	c.call = nil
	c.mu.Unlock()
	if cl == nil {
		return nil
	}

	var firstErr error
	if cl.input.ID != 0 {
		if _, err := c.api.PhoneDiscardCall(ctx, &tg.PhoneDiscardCallRequest{
			Peer:   cl.input,
			Reason: reason.tl(),
		}); err != nil {
			firstErr = errors.Wrap(err, "discard call")
		}
	}
	if cl.conn != nil {
		if err := cl.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (c *Client) setActive(cl *call) {
	c.mu.Lock()
	c.call = cl
	c.mu.Unlock()
}

func (c *Client) clearActive(cl *call) {
	c.mu.Lock()
	if c.call == cl {
		c.call = nil
	}
	conn := cl.conn
	c.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func (c *Client) setInput(cl *call, in tg.InputPhoneCall) {
	c.mu.Lock()
	cl.input = in
	c.mu.Unlock()
}

func newCallState(isCaller bool) *call {
	return &call{
		isCaller:  isCaller,
		accepted:  make(chan *tg.PhoneCallAccepted, 1),
		confirmed: make(chan *tg.PhoneCall, 1),
		discarded: make(chan *tg.PhoneCallDiscarded, 1),
	}
}

func trySend[T any](ch chan T, v T) {
	select {
	case ch <- v:
	default:
	}
}
