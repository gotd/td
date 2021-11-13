package telegram

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

func (c *Client) invokeSub(ctx context.Context, dc int, input bin.Encoder, output bin.Decoder) error {
	c.subConnsMux.Lock()

	conn, ok := c.subConns[dc]
	if ok {
		c.subConnsMux.Unlock()
		return conn.Invoke(ctx, input, output)
	}

	conn, err := c.dc(ctx, dc, 1, c.primaryDC(dc))
	if err != nil {
		c.subConnsMux.Unlock()
		return errors.Wrapf(err, "create connection to DC %d", dc)
	}
	c.subConns[dc] = conn
	c.subConnsMux.Unlock()

	return conn.Invoke(ctx, input, output)
}
