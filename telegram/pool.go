package telegram

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/pool"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// CloseInvoker is a closeable tg.Invoker.
type CloseInvoker interface {
	tg.Invoker
	Close() error
}

func (c *Client) createPool(dc int, max int64, creator func() pool.Conn) (*pool.DC, error) {
	select {
	case <-c.ctx.Done():
		return nil, xerrors.Errorf("client already closed: %w", c.ctx.Err())
	default:
	}

	p := pool.NewDC(c.ctx, dc, creator, pool.DCOptions{
		Logger:             c.log.Named("pool").With(zap.Int("dc_id", dc)),
		MaxOpenConnections: max,
	})

	return p, nil
}

func (c *Client) dc(ctx context.Context, dcID int, max int64, transfer bool, dialer mtproto.Dialer) (*pool.DC, error) {
	if max < 0 {
		return nil, xerrors.Errorf("invalid max value %d", max)
	}

	dcList := dcs.FindDCs(c.cfg.Load().DCOptions, dcID, false)
	if len(dcList) < 1 {
		return nil, xerrors.Errorf("unknown DC %d", dcID)
	}
	c.log.Debug("Creating pool",
		zap.Int("dc_id", dcID),
		zap.Int64("max", max),
		zap.Int("candidates", len(dcList)),
	)

	opts := c.opts
	p, err := c.createPool(dcID, max, func() pool.Conn {
		id := c.connsCounter.Inc()

		c.sessionsMux.Lock()
		session, ok := c.sessions[dcID]
		if !ok {
			session = pool.NewSyncSession(pool.Session{})
			c.sessions[dcID] = session
		}
		c.sessionsMux.Unlock()

		options, _ := session.Options(opts)
		options.Logger = c.log.Named("conn").With(
			zap.Int64("conn_id", id),
			zap.Int("dc_id", dcID),
		)
		return c.create(
			dialer, manager.ConnModeData, c.appID,
			options, manager.ConnOptions{
				DC:      dcID,
				Device:  c.device,
				Handler: c.asHandler(),
			},
		)
	})
	if err != nil {
		return nil, xerrors.Errorf("create pool: %w", err)
	}

	if transfer {
		_, err = c.transfer(ctx, tg.NewClient(p), dcID)
		if err != nil {
			// Ignore case then we are not authorized.
			if unauthorized(err) {
				return p, nil
			}

			// Kill pool if we got error.
			_ = p.Close()
			return nil, xerrors.Errorf("transfer: %w", err)
		}
	}

	return p, nil
}

func (c *Client) createDC(ctx context.Context, dc int, max int64, dialer mtproto.Dialer) (CloseInvoker, error) {
	// Lock migrations during creation of pool to prevent current DC changing.

	// Acquire or cancel.
	select {
	case c.migration <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	// Release.
	defer func() {
		<-c.migration
	}()

	currentDC := c.session.Load().DC
	needTransfer := currentDC != 0 && currentDC == dc

	return c.dc(ctx, dc, max, needTransfer, dialer)
}

// DC creates new multi-connection invoker to given DC.
func (c *Client) DC(ctx context.Context, dc int, max int64) (CloseInvoker, error) {
	p, err := c.createDC(ctx, dc, max, c.primaryDC(dc))
	if err != nil {
		return nil, err
	}

	return p, nil
}

// MediaOnly creates new multi-connection invoker to given DC ID.
// It connects to MediaOnly DCs.
func (c *Client) MediaOnly(ctx context.Context, dc int, max int64) (CloseInvoker, error) {
	p, err := c.createDC(ctx, dc, max, func(ctx context.Context) (transport.Conn, error) {
		return c.resolver.MediaOnly(ctx, dc, c.dcList())
	})
	if err != nil {
		return nil, err
	}

	return p, nil
}
