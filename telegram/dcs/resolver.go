package dcs

import (
	"context"
	"net"
	"strconv"
	"sync"

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
	return p.connectFastest(ctx, FindPrimaryDCs(dcOptions, dc, p.preferIPv6))
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
	return p.connectFastest(ctx, candidates[:n])
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
	return p.connectFastest(ctx, candidates[:n])
}

func (p plain) connect(ctx context.Context, server tg.DCOption) (transport.Conn, error) {
	addr := net.JoinHostPort(server.IPAddress, strconv.Itoa(server.Port))
	conn, err := p.dial(ctx, p.network, addr)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}

	transportConn, err := p.protocol.Handshake(conn)
	if err != nil {
		err = xerrors.Errorf("transport handshake: %w", err)
		return nil, multierr.Combine(err, conn.Close())
	}

	return transportConn, nil
}

// connectFastest concurrently dials all candidates from servers list,
// selecting first valid candidate.
func (p plain) connectFastest(ctx context.Context, servers []tg.DCOption) (transport.Conn, error) {
	if len(servers) == 0 {
		return nil, xerrors.New("no candidates")
	}

	var (
		connSelected transport.Conn
		connErrors   []error
		connMux      sync.Mutex
	)

	var g sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := range servers {
		opt := servers[i]
		g.Add(1)

		// Not limiting concurrent dials for now.
		go func() {
			defer g.Done()

			conn, err := p.connect(ctx, opt)
			connMux.Lock()
			defer connMux.Unlock()

			if err != nil {
				// Can be dial error or just context cancellation.
				connErrors = append(connErrors, err)
				return
			}

			if connSelected == nil {
				// Selecting connection and stopping candidate selection.
				connSelected = conn
				cancel()
			} else {
				// Dropping connection as other connection ls already selected.
				// Probably we can do it without mutex.
				_ = conn.Close()
			}
		}()
	}

	// Waiting for candidate selection.
	// Should stop on first valid candidate.
	g.Wait()

	connMux.Lock()
	defer connMux.Unlock()

	if connSelected != nil {
		return connSelected, nil
	}

	for _, err := range connErrors {
		// Return first error.
		return nil, xerrors.Errorf("dial: %w", err)
	}

	// Should be unreachable.
	return nil, xerrors.New("unable to select")
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
