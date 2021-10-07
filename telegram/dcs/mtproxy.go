package dcs

import (
	"context"
	"io"
	"net"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/mtproxy"
	"github.com/nnqq/td/internal/mtproxy/obfuscator"
	"github.com/nnqq/td/internal/proto/codec"
	"github.com/nnqq/td/transport"
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
	return m.resolve(ctx, dc+10000)
}

func (m mtProxy) CDN(ctx context.Context, dc int, _ List) (transport.Conn, error) {
	return m.resolve(ctx, dc)
}

func (m mtProxy) resolve(ctx context.Context, dc int) (transport.Conn, error) {
	c, err := m.dial(ctx, m.network, m.addr)
	if err != nil {
		return nil, xerrors.Errorf("connect to the MTProxy %q: %w", m.addr, err)
	}

	conn, err := m.handshakeConn(c, dc)
	if err != nil {
		err = xerrors.Errorf("handshake: %w", err)
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
		return nil, xerrors.Errorf("unknown MTProxy secret type: %d", m.secret.Type)
	}

	secret := m.secret
	secret.DC = dc
	if err := obsConn.Handshake(m.tag, secret); err != nil {
		return nil, xerrors.Errorf("MTProxy handshake: %w", err)
	}

	transportConn, err := m.protocol.Handshake(obsConn)
	if err != nil {
		return nil, xerrors.Errorf("transport handshake: %w", err)
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
	s, err := mtproxy.ParseSecret(2, secret)
	if err != nil {
		return nil, err
	}

	cdc := codec.PaddedIntermediate{}
	tag := codec.PaddedIntermediateClientStart

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
