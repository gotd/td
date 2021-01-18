package transport

import (
	"context"
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mtproxy"
	"github.com/gotd/td/internal/mtproxy/obfuscator"
	"github.com/gotd/td/internal/proto/codec"
)

type mtProxyDialer struct {
	original Dialer
	rand     io.Reader

	secret mtproxy.Secret
	tag    [4]byte
}

func (m mtProxyDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	c, err := m.original.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	var conn *obfuscator.Conn
	switch m.secret.Type {
	case mtproxy.Simple, mtproxy.Secured:
		conn = obfuscator.Obfuscated2(m.rand, c)
	case mtproxy.TLS:
		conn = obfuscator.FakeTLS(m.rand, c)
	default:
		return nil, xerrors.Errorf("unknown MTProxy secret type: %d", m.secret.Type)
	}

	if err := conn.Handshake(m.tag, m.secret); err != nil {
		return nil, xerrors.Errorf("MTProxy handshake: %w", err)
	}

	return conn, nil
}

// MTProxy creates MTProxy obfuscated transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation
func MTProxy(d Dialer, dc int, secret []byte) (*Transport, error) {
	s, err := mtproxy.ParseSecret(dc, secret)
	if err != nil {
		return nil, err
	}

	cdc := codec.PaddedIntermediate{}
	dialer := mtProxyDialer{
		original: orDefaultDialer(d),
		rand:     rand.Reader,
		secret:   s,
		tag:      cdc.ObfuscatedTag(),
	}

	return NewTransport(dialer, func() Codec {
		return codec.NoHeader{Codec: cdc}
	}), nil
}
