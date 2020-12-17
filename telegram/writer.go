package telegram

import (
	"context"

	"github.com/gotd/td/internal/proto"
)

func (c *Client) writeLoop(ctx context.Context) {
	for {
		select {
		case b := <-c.pchan:
			if err := proto.WriteIntermediate(c.conn, b); err != nil {
				c.pchan <- b

				go c.reconnect()
				return
			}
		default:
		}

		select {
		case <-ctx.Done():
			return
		case b := <-c.wchan:
			// we have priority message, skip this
			if len(c.pchan) > 0 {
				c.wchan <- b
				continue
			}

			if err := proto.WriteIntermediate(c.conn, b); err != nil {
				c.wchan <- b

				go c.reconnect()
				return
			}
		}
	}
}
