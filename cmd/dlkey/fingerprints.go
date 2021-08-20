package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/exchange"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/transport"
)

func dial(ctx context.Context, addr string) (_ transport.Conn, rErr error) {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rErr != nil {
			multierr.AppendInto(&rErr, conn.Close())
		}
	}()

	transportConn, err := transport.Intermediate.Handshake(conn)
	if err != nil {
		return nil, xerrors.Errorf("transport handshake: %w", err)
	}

	return transportConn, nil
}

func getAvailable(ctx context.Context, keys Keys) (Keys, error) {
	available := Keys{}
	dedup := map[int64]struct{}{}

	for listName, list := range map[string]dcs.List{
		"production": dcs.Prod(),
		"staging":    dcs.Staging(),
	} {
		for _, dc := range list.Options {
			if dc.TCPObfuscatedOnly || dc.Ipv6 {
				continue
			}

			addr := net.JoinHostPort(dc.IPAddress, strconv.Itoa(dc.Port))
			conn, err := dial(ctx, addr)
			if err != nil {
				return nil, xerrors.Errorf("dial: %w", err)
			}

			fingerprints, err := exchange.FetchFingerprints(ctx, conn)
			if err != nil {
				return nil, xerrors.Errorf("fetch fingerprints: %w", err)
			}
			fmt.Printf("%s DC %d (%s), fingerprints: %v\n", listName, dc.ID, addr, fingerprints)

			for _, fingerprint := range fingerprints {
				if _, ok := dedup[fingerprint]; ok {
					continue
				}
				dedup[fingerprint] = struct{}{}

				v, ok := keys.Find(fingerprint)
				if ok {
					available = append(available, v)
				}
			}
		}
	}
	fmt.Printf("Parsed: %d, fetched: %d, available: %d\n", keys.Len(), len(dedup), available.Len())

	return available, nil
}
