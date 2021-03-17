package telegram

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/pool"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
)

// CloseInvoker is a closeable tg.Invoker.
type CloseInvoker interface {
	tg.Invoker
	Close(ctx context.Context) error
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

// Pool creates new multi-connection invoker to current DC.
func (c *Client) Pool(max int64) (CloseInvoker, error) {
	if max < 0 {
		return nil, xerrors.Errorf("invalid max value %d", max)
	}

	s := c.session.Load()
	return c.createPool(s.DC, max, func() pool.Conn {
		id := c.connsCounter.Inc()
		return c.createConn(id, manager.ConnModeData, nil)
	})
}

func (c *Client) dc(ctx context.Context, dcID int, max int64) (*pool.DC, error) {
	if max < 0 {
		return nil, xerrors.Errorf("invalid max value %d", max)
	}

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
			c.dialerDC(dcID), manager.ConnModeData, c.appID,
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

	_, err = c.transfer(ctx, tg.NewClient(p), dcID)
	if err != nil {
		// Ignore case then we are not authorized.
		if unauthorized(err) {
			return p, nil
		}

		return nil, xerrors.Errorf("transfer: %w", err)
	}

	return p, nil
}

// DC creates new multi-connection invoker to given DC.
func (c *Client) DC(ctx context.Context, id int, max int64) (CloseInvoker, error) {
	return c.dc(ctx, id, max)
}
