package telegram

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/manager"
)

func (c *Client) invokeSub(ctx context.Context, dc int, input bin.Encoder, output bin.Decoder) error {
	c.subConnsMux.Lock()

	conn, ok := c.subConns[dc]
	if ok {
		c.subConnsMux.Unlock()
		return conn.Invoke(ctx, input, output)
	}

	// Sub-invoker is regular data connection to target DC.
	conn, err := c.dc(ctx, dc, 1, c.primaryDC(dc), manager.ConnModeData)
	if err != nil {
		c.subConnsMux.Unlock()
		return errors.Wrapf(err, "create connection to DC %d", dc)
	}
	c.subConns[dc] = conn
	c.subConnsMux.Unlock()

	return conn.Invoke(ctx, input, output)
}
