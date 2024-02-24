package dcs

import (
	"context"
	"io"
	"net"
	"strconv"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproxy"
	"github.com/gotd/td/mtproxy/obfuscator"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

var _ Resolver = plain{}

type plain struct {
	dial         DialFunc
	protocol     Protocol
	rand         io.Reader
	network      string
	noObfuscated bool
	preferIPv6   bool
}

func (p plain) Primary(ctx context.Context, dc int, list List) (transport.Conn, error) {
	candidates := FindPrimaryDCs(list.Options, dc, p.preferIPv6)
	if p.noObfuscated {
		n := 0
		for _, x := range candidates {
			if !x.TCPObfuscatedOnly {
				candidates[n] = x
				n++
			}
		}
		candidates = candidates[:n]
	}
	return p.connect(ctx, dc, list.Test, candidates)
}

func (p plain) MediaOnly(ctx context.Context, dc int, list List) (transport.Conn, error) {
	candidates := FindDCs(list.Options, dc, p.preferIPv6)
	// Filter (in place) from SliceTricks.
	n := 0
	for _, x := range candidates {
		if x.MediaOnly {
			candidates[n] = x
			n++
		}
	}
	return p.connect(ctx, dc, list.Test, candidates[:n])
}

func (p plain) CDN(ctx context.Context, dc int, list List) (transport.Conn, error) {
	return nil, errors.Errorf("can't resolve %d: CDN is unsupported", dc)
}

func (p plain) dialTransport(ctx context.Context, test bool, dc tg.DCOption) (_ transport.Conn, rerr error) {
	addr := net.JoinHostPort(dc.IPAddress, strconv.Itoa(dc.Port))

	conn, err := p.dial(ctx, p.network, addr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr != nil {
			multierr.AppendInto(&rerr, conn.Close())
		}
	}()

	proto := p.protocol
	if dc.TCPObfuscatedOnly {
		dcID := dc.ID
		if test {
			if dcID < 0 {
				dcID -= 10000
			} else {
				dcID += 10000
			}
		}

		var (
			cdc  codec.Codec = codec.Intermediate{}
			tag              = codec.IntermediateClientStart
			obfs             = obfuscator.Obfuscated2
		)

		secret, err := mtproxy.ParseSecret(dc.Secret)
		if err != nil {
			return nil, errors.Wrap(err, "check DC secret")
		}

		if secret.Type == mtproxy.TLS {
			obfs = obfuscator.FakeTLS
		} else if c, ok := secret.ExpectedCodec(); ok {
			tag = [4]byte{secret.Tag, secret.Tag, secret.Tag, secret.Tag}
			cdc = c
		}

		obfsConn := obfs(p.rand, conn)
		if err := obfsConn.Handshake(tag, dcID, secret); err != nil {
			return nil, err
		}
		conn = obfsConn

		proto = transport.NewProtocol(func() transport.Codec {
			return codec.NoHeader{Codec: cdc}
		})
	}

	transportConn, err := proto.Handshake(conn)
	if err != nil {
		return nil, errors.Wrap(err, "transport handshake")
	}

	return transportConn, nil
}

func (p plain) connect(ctx context.Context, dc int, test bool, dcOptions []tg.DCOption) (transport.Conn, error) {
	switch len(dcOptions) {
	case 0:
		return nil, errors.Errorf("no addresses for DC %d", dc)
	case 1:
		return p.dialTransport(ctx, test, dcOptions[0])
	}

	type dialResult struct {
		conn transport.Conn
		err  error
	}

	// We use unbuffered channel to ensure that only one connection will be returned
	// and all other will be closed.
	results := make(chan dialResult)
	tryDial := func(ctx context.Context, option tg.DCOption) {
		conn, err := p.dialTransport(ctx, test, option)
		select {
		case results <- dialResult{
			conn: conn,
			err:  err,
		}:
		case <-ctx.Done():
			if conn != nil {
				_ = conn.Close()
			}
		}
	}

	dialCtx, dialCancel := context.WithCancel(ctx)
	defer dialCancel()

	for _, dcOption := range dcOptions {
		go tryDial(dialCtx, dcOption)
	}

	remain := len(dcOptions)
	var rErr error
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-results:
			remain--
			if result.err != nil {
				rErr = multierr.Append(rErr, result.err)
				if remain == 0 {
					return nil, rErr
				}
				continue
			}
			return result.conn, nil
		}
	}
}

// PlainOptions is plain resolver creation options.
type PlainOptions struct {
	// Protocol is the transport protocol to use. Defaults to intermediate.
	Protocol Protocol
	// Dial specifies the dial function for creating unencrypted TCP connections.
	// If Dial is nil, then the resolver dials using package net.
	Dial DialFunc
	// Random source for TCPObfuscated DCs.
	Rand io.Reader
	// Network to use. Defaults to "tcp".
	Network string
	// NoObfuscated denotes to filter out TCP Obfuscated Only DCs.
	NoObfuscated bool
	// PreferIPv6 gives IPv6 DCs higher precedence.
	// Default is to prefer IPv4 DCs over IPv6.
	PreferIPv6 bool
}

func (m *PlainOptions) setDefaults() {
	if m.Protocol == nil {
		m.Protocol = transport.Intermediate
	}
	if m.Dial == nil {
		var d net.Dialer
		m.Dial = d.DialContext
	}
	if m.Rand == nil {
		m.Rand = crypto.DefaultRand()
	}
	if m.Network == "" {
		m.Network = "tcp"
	}
}

// Plain creates plain DC resolver.
func Plain(opts PlainOptions) Resolver {
	opts.setDefaults()
	return plain{
		dial:         opts.Dial,
		protocol:     opts.Protocol,
		rand:         opts.Rand,
		network:      opts.Network,
		noObfuscated: opts.NoObfuscated,
		preferIPv6:   opts.PreferIPv6,
	}
}
