package dcs

import (
	"context"
	"io"
	"net"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproxy"
	"github.com/gotd/td/mtproxy/obfuscator"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/transport"
)

var _ Resolver = mtProxy{}

type mtProxy struct {
	dial          DialFunc
	protocol      protocol
	addr, network string

	secret mtproxy.Secret
	tag    [4]byte
	rand   io.Reader
}

func (m mtProxy) Primary(ctx context.Context, dc int, _ List) (transport.Conn, error) {
	return m.resolve(ctx, dc)
}

func (m mtProxy) MediaOnly(ctx context.Context, dc int, _ List) (transport.Conn, error) {
	if dc > 0 {
		dc *= -1
	}
	return m.resolve(ctx, dc)
}

func (m mtProxy) CDN(ctx context.Context, dc int, _ List) (transport.Conn, error) {
	return m.resolve(ctx, dc)
}

func (m mtProxy) resolve(ctx context.Context, dc int) (transport.Conn, error) {
	c, err := m.dial(ctx, m.network, m.addr)
	if err != nil {
		return nil, errors.Wrapf(err, "connect to the MTProxy %q", m.addr)
	}

	conn, err := m.handshakeConn(c, dc)
	if err != nil {
		err = errors.Wrap(err, "handshake")
		return nil, multierr.Combine(err, c.Close())
	}

	return conn, nil
}

// handshakeConn inits given net.Conn as MTProto connection.
func (m mtProxy) handshakeConn(c net.Conn, dc int) (transport.Conn, error) {
	var obsConn *obfuscator.Conn
	switch m.secret.Type {
	case mtproxy.Simple, mtproxy.Secured:
		obsConn = obfuscator.Obfuscated2(m.rand, c)
	case mtproxy.TLS:
		obsConn = obfuscator.FakeTLS(m.rand, c)
	default:
		return nil, errors.Errorf("unknown MTProxy secret type: %d", m.secret.Type)
	}

	secret := m.secret
	if err := obsConn.Handshake(m.tag, dc, secret); err != nil {
		return nil, errors.Wrap(err, "MTProxy handshake")
	}

	transportConn, err := m.protocol.Handshake(obsConn)
	if err != nil {
		return nil, errors.Wrap(err, "transport handshake")
	}

	return transportConn, nil
}

// MTProxyOptions is MTProxy resolver creation options.
type MTProxyOptions struct {
	// Dial specifies the dial function for creating unencrypted TCP connections.
	// If Dial is nil, then the resolver dials using package net.
	Dial DialFunc
	// Network to use. Defaults to "tcp"
	Network string
	// Random source for MTProxy obfuscator.
	Rand io.Reader
}

func (m *MTProxyOptions) setDefaults() {
	if m.Dial == nil {
		var d net.Dialer
		m.Dial = d.DialContext
	}
	if m.Network == "" {
		m.Network = "tcp"
	}
	if m.Rand == nil {
		m.Rand = crypto.DefaultRand()
	}
}

// MTProxy creates MTProxy obfuscated DC resolver.
//
// See https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation.
func MTProxy(addr string, secret []byte, opts MTProxyOptions) (Resolver, error) {
	s, err := mtproxy.ParseSecret(secret)
	if err != nil {
		return nil, err
	}

	var cdc codec.Codec = codec.PaddedIntermediate{}
	tag := codec.PaddedIntermediateClientStart

	// FIXME(tdakkota): some proxies forces to use Padded (Secure) Intermediate
	// 	even if secret denotes to use another transport type.
	if s.Type != mtproxy.TLS {
		if c, ok := s.ExpectedCodec(); ok {
			cdc = c
			tag = [4]byte{s.Tag, s.Tag, s.Tag, s.Tag}
		}
	}

	opts.setDefaults()
	return mtProxy{
		dial:     opts.Dial,
		addr:     addr,
		network:  opts.Network,
		protocol: transport.NewProtocol(func() transport.Codec { return codec.NoHeader{Codec: cdc} }),
		secret:   s,
		tag:      tag,
		rand:     opts.Rand,
	}, nil
}
