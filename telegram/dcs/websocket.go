package dcs

import (
	"context"
	"io"

	"github.com/coder/websocket"
	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproxy"
	"github.com/gotd/td/mtproxy/obfuscator"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/transport"
	"github.com/gotd/td/wsutil"
)

var _ Resolver = ws{}

type ws struct {
	dialOptions *websocket.DialOptions
	protocol    protocol

	tag  [4]byte
	rand io.Reader
}

func (w ws) connect(ctx context.Context, dc int, domains map[int]string) (transport.Conn, error) {
	addr, ok := domains[dc]
	if !ok {
		return nil, errors.Errorf("domain for %d not found", dc)
	}

	conn, resp, err := websocket.Dial(ctx, addr, w.dialOptions)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "dial ws")
	}
	obsConn := obfuscator.Obfuscated2(w.rand, wsutil.NetConn(conn))

	if err := obsConn.Handshake(w.tag, dc, mtproxy.Secret{
		Secret: nil,
		Type:   mtproxy.Simple,
	}); err != nil {
		return nil, errors.Wrap(err, "handshake")
	}

	transportConn, err := w.protocol.Handshake(obsConn)
	if err != nil {
		return nil, errors.Wrap(err, "transport handshake")
	}

	return transportConn, nil
}

func (w ws) Primary(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return w.connect(ctx, dc, list.Domains)
}

func (w ws) MediaOnly(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return nil, errors.Errorf("can't resolve %d: MediaOnly is unsupported", dc)
}

func (w ws) CDN(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return nil, errors.Errorf("can't resolve %d: CDN is unsupported", dc)
}

// WebsocketOptions is Websocket resolver creation options.
type WebsocketOptions struct {
	// Dialer specifies the websocket dialer.
	// If Dialer is nil, then the resolver dials using websocket.DefaultDialer.
	DialOptions *websocket.DialOptions
	// Random source for MTProxy obfuscator.
	Rand io.Reader
}

func (m *WebsocketOptions) setDefaults() {
	if m.DialOptions == nil {
		m.DialOptions = &websocket.DialOptions{Subprotocols: []string{
			"binary",
		}}
	}
	if m.Rand == nil {
		m.Rand = crypto.DefaultRand()
	}
}

// Websocket creates Websocket DC resolver.
//
// See https://core.telegram.org/mtproto/transports#websocket.
func Websocket(opts WebsocketOptions) Resolver {
	cdc := codec.Intermediate{}
	opts.setDefaults()

	return ws{
		dialOptions: opts.DialOptions,
		protocol:    transport.NewProtocol(func() transport.Codec { return codec.NoHeader{Codec: cdc} }),
		tag:         cdc.ObfuscatedTag(),
		rand:        opts.Rand,
	}
}
