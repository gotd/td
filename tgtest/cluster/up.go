package cluster

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/transport"
)

// Up runs all servers in a cluster.
func (c *Cluster) Up(ctx context.Context) error {
	g := tdsync.NewCancellableGroup(ctx)

	listen := func(ctx context.Context, _ int) (net.Listener, error) {
		return newLocalListener(ctx)
	}
	if c.web {
		// Create local random listener
		l, err := newLocalListener(ctx)
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		srv := http.Server{
			Handler: mux,
			BaseContext: func(net.Listener) context.Context {
				return ctx
			},
		}
		g.Go(func(ctx context.Context) error {
			if err := srv.Serve(l); err != nil && !xerrors.Is(err, http.ErrServerClosed) {
				return xerrors.Errorf("serve: %w", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()

			return srv.Close()
		})

		baseURL := url.URL{
			Scheme: "http",
			Host:   l.Addr().String(),
		}
		listen = func(ctx context.Context, dc int) (net.Listener, error) {
			listener, handler := transport.WebsocketListener(l.Addr())

			path := fmt.Sprintf("/dc/%d", dc)
			mux.Handle(path, handler)

			dcURL := baseURL
			dcURL.Path = path
			c.domains[dc] = dcURL.String()
			return listener, nil
		}
	}

	for dcID, s := range c.setups {
		l, err := listen(ctx, dcID)
		if err != nil {
			return xerrors.Errorf("DC %d: listen port: %w", dcID, err)
		}

		if !c.web {
			// Add TCP listeners to config.
			if addr, ok := l.Addr().(*net.TCPAddr); ok {
				c.cfg.DCOptions = append(c.cfg.DCOptions, tg.DCOption{
					Ipv6:      addr.IP.To16() != nil,
					Static:    true,
					ID:        dcID,
					IPAddress: addr.IP.String(),
					Port:      addr.Port,
				})
			}
		}

		// Copy iteration value.
		srv := s.srv
		g.Go(func(ctx context.Context) error {
			return srv.Serve(ctx, transport.ListenCodec(nil, l))
		})
	}
	c.ready.Signal()

	return g.Wait()
}
