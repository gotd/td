package dcs

import (
	"context"
	"io"

	"golang.org/x/xerrors"
	"nhooyr.io/websocket"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/mtproxy"
	"github.com/nnqq/td/internal/mtproxy/obfuscator"
	"github.com/nnqq/td/internal/proto/codec"
	"github.com/nnqq/td/internal/wsutil"
	"github.com/nnqq/td/transport"
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
		return nil, xerrors.Errorf("domain for %d not found", dc)
	}

	conn, resp, err := websocket.Dial(ctx, addr, w.dialOptions)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, xerrors.Errorf("dial ws: %w", err)
	}
	obsConn := obfuscator.Obfuscated2(w.rand, wsutil.NetConn(conn))

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

func (w ws) Primary(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return w.connect(ctx, dc, list.Domains)
}

func (w ws) MediaOnly(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return nil, xerrors.Errorf("can't resolve %d: MediaOnly is unsupported", dc)
}

func (w ws) CDN(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return nil, xerrors.Errorf("can't resolve %d: CDN is unsupported", dc)
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

	var tag [4]byte
	copy(tag[:], cdc.ObfuscatedTag())
	return ws{
		dialOptions: opts.DialOptions,
		protocol:    transport.NewProtocol(func() transport.Codec { return codec.NoHeader{Codec: cdc} }),
		tag:         tag,
		rand:        opts.Rand,
	}
}
