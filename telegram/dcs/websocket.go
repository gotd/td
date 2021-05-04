package dcs

import (
	"context"
	"crypto/rand"
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mtproxy"
	"github.com/gotd/td/internal/mtproxy/obfuscator"
	"github.com/gotd/td/internal/proto/codec"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type wsConn struct {
	conn   *websocket.Conn
	reader io.Reader
}

func (c *wsConn) Write(p []byte) (int, error) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *wsConn) Read(p []byte) (int, error) {
	for {
		if c.reader == nil {
			// Advance to next message.
			var err error
			_, c.reader, err = c.conn.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.reader.Read(p)
		if err == io.EOF {
			// At end of message.
			c.reader = nil
			if n > 0 {
				return n, nil
			}

			// No data read, continue to next message.
			continue
		}
		return n, err
	}
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *wsConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *wsConn) SetDeadline(t time.Time) error {
	return multierr.Append(c.conn.SetReadDeadline(t), c.conn.SetWriteDeadline(t))
}

func (c *wsConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *wsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

type ws struct {
	dialer   *websocket.Dialer
	domain   func(dc int) (string, error)
	protocol protocol

	tag  [4]byte
	rand io.Reader
}

func (w ws) connect(ctx context.Context, dc int) (transport.Conn, error) {
	addr, err := w.domain(dc)
	if err != nil {
		return nil, xerrors.Errorf("resolve domain %d", dc)
	}

	conn, resp, err := w.dialer.DialContext(ctx, addr, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, xerrors.Errorf("dial ws: %w", err)
	}
	obsConn := obfuscator.Obfuscated2(w.rand, &wsConn{conn: conn})

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
	Dialer *websocket.Dialer
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
	if m.Dialer == nil {
		m.Dialer = websocket.DefaultDialer
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
		dialer:   opts.Dialer,
		domain:   opts.Domain,
		protocol: transport.NewProtocol(func() transport.Codec { return codec.NoHeader{Codec: cdc} }),
		tag:      cdc.ObfuscatedTag(),
		rand:     opts.Rand,
	}
}
