package dcs

import (
	"context"
	"net"
	"strconv"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// Resolver resolves DC and creates transport MTProto connection.
type Resolver interface {
	Primary(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error)
	MediaOnly(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error)
	CDN(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error)
}

// DialFunc connects to the address on the named network.
type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// Protocol is MTProto transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports
type Protocol interface {
	Codec() transport.Codec
	Handshake(conn net.Conn) (transport.Conn, error)
}

type plain struct {
	dial       DialFunc
	protocol   Protocol
	network    string
	preferIPv6 bool
}

func (p plain) Primary(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	return p.connect(ctx, dc, FindPrimaryDCs(dcOptions, dc, p.preferIPv6))
}

func (p plain) MediaOnly(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	candidates := FindDCs(dcOptions, dc, p.preferIPv6)
	// Filter (in place) from SliceTricks.
	n := 0
	for _, x := range candidates {
		if x.MediaOnly {
			candidates[n] = x
			n++
		}
	}
	return p.connect(ctx, dc, candidates[:n])
}

func (p plain) CDN(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	candidates := FindDCs(dcOptions, dc, p.preferIPv6)
	// Filter (in place) from SliceTricks.
	n := 0
	for _, x := range candidates {
		if x.CDN {
			candidates[n] = x
			n++
		}
	}
	return p.connect(ctx, dc, candidates[:n])
}

func (p plain) dialTransport(ctx context.Context, dc tg.DCOption) (_ transport.Conn, rerr error) {
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

	transportConn, err := p.protocol.Handshake(conn)
	if err != nil {
		return nil, xerrors.Errorf("transport handshake: %w", err)
	}

	return transportConn, nil
}

func (p plain) connect(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	switch len(dcOptions) {
	case 0:
		return nil, xerrors.Errorf("no addresses for DC %d", dc)
	case 1:
		return p.dialTransport(ctx, dcOptions[0])
	}

	type dialResult struct {
		conn transport.Conn
		err  error
	}

	// We use unbuffered channel to ensure that only one connection will be returned
	// and all other will be closed.
	results := make(chan dialResult)
	tryDial := func(ctx context.Context, option tg.DCOption) {
		conn, err := p.dialTransport(ctx, option)
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
	// Network to use. Defaults to "tcp".
	Network string
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
	if m.Network == "" {
		m.Network = "tcp"
	}
}

// PlainResolver creates plain DC resolver.
func PlainResolver(opts PlainOptions) Resolver {
	opts.setDefaults()
	return plain{
		protocol:   opts.Protocol,
		dial:       opts.Dial,
		network:    opts.Network,
		preferIPv6: opts.PreferIPv6,
	}
}
