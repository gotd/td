package dcs

import (
	"context"
	"errors"
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

// Transport is MTProto connection creator.
type Transport interface {
	Codec() transport.Codec
	Handshake(conn net.Conn) (transport.Conn, error)
}

type plain struct {
	dialer     transport.Dialer
	transport  Transport
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

func (p plain) connect(ctx context.Context, dc int, dcOptions []tg.DCOption) (transport.Conn, error) {
	for _, dc := range dcOptions {
		addr := net.JoinHostPort(dc.IPAddress, strconv.Itoa(dc.Port))
		conn, err := p.dialer.DialContext(ctx, p.network, addr)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
				select {
				case <-ctx.Done():
				default:
					continue
				}
			}
			return nil, err
		}

		transportConn, err := p.transport.Handshake(conn)
		if err != nil {
			err = xerrors.Errorf("transport handshake: %w", err)
			return nil, multierr.Combine(err, conn.Close())
		}

		return transportConn, nil
	}

	return nil, xerrors.Errorf("no addresses for DC %d", dc)
}

// PlainOptions is plain resolver creation options.
type PlainOptions struct {
	// Transport to use.
	Transport Transport
	// Dialer to use. Default net.Dialer will be used by default.
	Dialer transport.Dialer
	// Network to use.
	Network string
	// PreferIPv6 gives IPv6 DCs higher precedence.
	// Default is to prefer IPv4 DCs over IPv6.
	PreferIPv6 bool
}

func (m *PlainOptions) setDefaults() {
	if m.Transport == nil {
		m.Transport = transport.Intermediate()
	}
	if m.Dialer == nil {
		m.Dialer = &net.Dialer{}
	}
	if m.Network == "" {
		m.Network = "tcp"
	}
}

// PlainResolver creates plain DC resolver.
func PlainResolver(opts PlainOptions) Resolver {
	opts.setDefaults()
	return plain{
		transport:  opts.Transport,
		dialer:     opts.Dialer,
		network:    opts.Network,
		preferIPv6: opts.PreferIPv6,
	}
}
