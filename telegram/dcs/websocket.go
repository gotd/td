package dcs

import (
	"context"
	"crypto/rand"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"

	"github.com/gotd/td/internal/mtproxy"
	"github.com/gotd/td/internal/mtproxy/obfuscator"
	"github.com/gotd/td/internal/proto/codec"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type wsConn struct {
	conn *websocket.Conn

	writeTimer   *time.Timer
	writeContext context.Context

	readTimer   *time.Timer
	readContext context.Context

	readMu sync.Mutex
	eofed  bool
	reader io.Reader
}

func netConn(ctx context.Context, c *websocket.Conn) net.Conn {
	nc := &wsConn{
		conn: c,
	}

	var cancel context.CancelFunc
	nc.writeContext, cancel = context.WithCancel(ctx)
	nc.writeTimer = time.AfterFunc(math.MaxInt64, cancel)
	if !nc.writeTimer.Stop() {
		<-nc.writeTimer.C
	}

	nc.readContext, cancel = context.WithCancel(ctx)
	nc.readTimer = time.AfterFunc(math.MaxInt64, cancel)
	if !nc.readTimer.Stop() {
		<-nc.readTimer.C
	}

	return nc
}

func (w *wsConn) Write(b []byte) (int, error) {
	err := w.conn.Write(w.writeContext, websocket.MessageBinary, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *wsConn) Read(b []byte) (int, error) {
	w.readMu.Lock()
	defer w.readMu.Unlock()

	for {
		if w.reader == nil {
			// Advance to next message.
			var err error
			_, w.reader, err = w.conn.Reader(w.readContext)
			if err != nil {
				return 0, err
			}
		}
		n, err := w.reader.Read(b)
		if err == io.EOF {
			// At end of message.
			w.reader = nil
			if n > 0 {
				return n, nil
			}

			// No data read, continue to next message.
			continue
		}
		return n, err
	}
}

func (w *wsConn) Close() error {
	w.writeTimer.Stop()
	w.readTimer.Stop()
	return w.conn.Close(websocket.StatusNormalClosure, "")
}

type websocketAddr struct {
}

func (a websocketAddr) Network() string {
	return "websocket"
}

func (a websocketAddr) String() string {
	return "websocket/unknown-addr"
}

func (w *wsConn) LocalAddr() net.Addr {
	return websocketAddr{}
}

func (w *wsConn) RemoteAddr() net.Addr {
	return websocketAddr{}
}

func (w *wsConn) SetDeadline(t time.Time) error {
	return multierr.Append(w.SetWriteDeadline(t), w.SetReadDeadline(t))
}

func (w *wsConn) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		w.writeTimer.Stop()
	} else {
		w.writeTimer.Reset(t.Sub(time.Now()))
	}
	return nil
}

func (w *wsConn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		w.readTimer.Stop()
	} else {
		w.readTimer.Reset(t.Sub(time.Now()))
	}
	return nil
}

type ws struct {
	dialOptions *websocket.DialOptions
	domain      func(dc int) (string, error)
	protocol    protocol

	tag  [4]byte
	rand io.Reader
}

func (w ws) connect(ctx context.Context, dc int) (transport.Conn, error) {
	addr, err := w.domain(dc)
	if err != nil {
		return nil, xerrors.Errorf("resolve domain %d", dc)
	}

	conn, resp, err := websocket.Dial(ctx, addr, w.dialOptions)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, xerrors.Errorf("dial ws: %w", err)
	}
	obsConn := obfuscator.Obfuscated2(w.rand, netConn(context.Background(), conn))

	if err := obsConn.Handshake(w.tag, mtproxy.Secret{
		DC:     dc,
		Secret: nil,
		Type:   mtproxy.Simple,
	}); err != nil {
		return nil, xerrors.Errorf("handshake: %w", err)
	}

	transportConn, err := w.protocol.Handshake(obsConn)
	if err != nil {
		return nil, xerrors.Errorf("transport handshake: %w", err)
	}

	return transportConn, nil
}

func (w ws) Primary(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	return w.connect(ctx, dc)
}

func (w ws) MediaOnly(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	return nil, xerrors.Errorf("can't resolve %d: MediaOnly is unsupported", dc)
}

func (w ws) CDN(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	return nil, xerrors.Errorf("can't resolve %d: CDN is unsupported", dc)
}

// WebsocketOptions is Websocket resolver creation options.
type WebsocketOptions struct {
	// Dialer specifies the websocket dialer.
	// If Dialer is nil, then the resolver dials using websocket.DefaultDialer.
	DialOptions *websocket.DialOptions
	// Random source for MTProxy obfuscator.
	Rand io.Reader
	// Domain resolves connection URL by DC ID.
	Domain func(dc int) (string, error)
}

// nolint:gochecknoglobals
var websocketDomains = map[int]string{
	1: "wss://pluto.web.telegram.org/apiws",
	2: "wss://venus.web.telegram.org/apiws",
	3: "wss://aurora.web.telegram.org/apiws",
	4: "wss://vesta.web.telegram.org/apiws",
	5: "wss://flora.web.telegram.org/apiws",
}

func (m *WebsocketOptions) setDefaults() {
	if m.DialOptions == nil {
		m.DialOptions = &websocket.DialOptions{Subprotocols: []string{
			"binary",
		}}
	}
	if m.Rand == nil {
		m.Rand = rand.Reader
	}
	if m.Domain == nil {
		m.Domain = func(dc int) (string, error) {
			v, ok := websocketDomains[dc]
			if !ok {
				return "", xerrors.Errorf("domain for %d not found", dc)
			}
			return v, nil
		}
	}
}

// WebsocketResolver creates Websocket DC resolver.
//
// See https://core.telegram.org/mtproto/transports#websocket.
func WebsocketResolver(opts WebsocketOptions) Resolver {
	cdc := codec.Intermediate{}
	opts.setDefaults()
	return ws{
		dialOptions: opts.DialOptions,
		domain:      opts.Domain,
		protocol:    transport.NewProtocol(func() transport.Codec { return codec.NoHeader{Codec: cdc} }),
		tag:         cdc.ObfuscatedTag(),
		rand:        opts.Rand,
	}
}
