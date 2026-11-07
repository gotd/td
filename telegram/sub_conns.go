package telegram

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/manager"
)

// subConnCall is an in-flight creation of a sub-connection to some DC.
type subConnCall struct {
	done chan struct{}
	conn CloseInvoker
	err  error
}

func (c *Client) invokeSub(ctx context.Context, dc int, input bin.Encoder, output bin.Decoder) error {
	conn, err := c.subConn(ctx, dc)
	if err != nil {
		return errors.Wrapf(err, "create connection to DC %d", dc)
	}
	return conn.Invoke(ctx, input, output)
}

// subConn returns cached sub-connection to given DC or creates new one.
//
// Creation (dial, key exchange and authorization transfer) is done once per DC
// without holding subConnsMux, so redirects to other DCs and calls to already
// created sub-connections are not blocked. Creation is bounded by migration
// timeout and is not canceled when the caller that started it gives up.
// Failed creation is not cached.
func (c *Client) subConn(ctx context.Context, dc int) (CloseInvoker, error) {
	c.subConnsMux.Lock()
	if conn, ok := c.subConns[dc]; ok {
		c.subConnsMux.Unlock()
		return conn, nil
	}
	call, ok := c.subConnsPending[dc]
	if !ok {
		call = &subConnCall{done: make(chan struct{})}
		c.subConnsPending[dc] = call
		go c.createSubConn(ctx, dc, call)
	}
	c.subConnsMux.Unlock()

	select {
	case <-call.done:
		return call.conn, call.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) createSubConn(ctx context.Context, dc int, call *subConnCall) {
	// Keep caller values (tracing, logging), but not its cancellation:
	// creation is shared by all callers waiting for this DC.
	clientCtx := c.ctx
	if clientCtx == nil {
		clientCtx = context.Background()
	}
	cctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	defer cancel()
	stop := context.AfterFunc(clientCtx, cancel)
	defer stop()
	if c.migrationTimeout > 0 {
		var cancelTimeout context.CancelFunc
		cctx, cancelTimeout = context.WithTimeout(cctx, c.migrationTimeout)
		defer cancelTimeout()
	}

	// Sub-invoker is regular data connection to target DC.
	conn, err := c.dc(cctx, dc, 1, c.primaryDC(dc), manager.ConnModeData)
	if err != nil && clientCtx.Err() != nil {
		err = errors.Wrap(clientCtx.Err(), "client closed")
	}

	c.subConnsMux.Lock()
	delete(c.subConnsPending, dc)
	if err == nil {
		if clientCtx.Err() != nil {
			// Client is closing (or closed) and has already drained subConns.
			err = errors.Wrap(clientCtx.Err(), "client closed")
		} else {
			c.subConns[dc] = conn
		}
	}
	call.conn, call.err = conn, err
	if err != nil {
		call.conn = nil
	}
	c.subConnsMux.Unlock()
	close(call.done)

	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		c.log.Debug(ctx, "Sub-connection creation failed",
			log.Int("dc_id", dc),
			log.Error(err),
		)
	}
}
