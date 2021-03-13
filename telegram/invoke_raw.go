package telegram

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// InvokeRaw sens input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Client) InvokeRaw(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
	c.pmux.RLock()
	primary := c.primary
	c.pmux.RUnlock()

	if err := primary.InvokeRaw(ctx, in, out); err != nil {
		// Handling datacenter migration request.
		if rpcErr, ok := tgerr.As(err); ok && rpcErr.IsCode(303) {
			// If migration error is FILE_MIGRATE or STATS_MIGRATE, then the method
			// called by authorized client, so we should try to transfer auth to new DC
			// and create new connection.
			if rpcErr.IsOneOf("FILE_MIGRATE", "STATS_MIGRATE") {
				c.log.Info("Got migrate error: Creating sub-connection",
					zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
				)
				return c.invokeDC(ctx, rpcErr.Argument, in, out)
			}

			c.log.Info("Got migrate error",
				zap.String("error", rpcErr.Type), zap.Int("dc", rpcErr.Argument),
			)

			// Prevent parallel migrations.
			cb, perform := c.migrateOp.Try()
			if !perform {
				c.log.Info("Other goroutine has already started migration, waiting for completion")
				cb()
				c.log.Info("Other goroutine has completed the migration, re-invoking request on new DC")
				return c.InvokeRaw(ctx, in, out)
			}

			c.log.Info("Starting migration to another DC", zap.Int("dc", rpcErr.Argument))
			defer cb()
			dcInfo, err := c.lookupDC(rpcErr.Argument)
			if err != nil {
				return err
			}

			// Change primary DC.
			if _, err := c.dc(dcInfo).AsPrimary().Connect(ctx); err != nil {
				return xerrors.Errorf("migrate to dc %d: %w", rpcErr.Argument, err)
			}

			c.log.Info("Migration completed, re-invoking request on new DC")
			return c.InvokeRaw(ctx, in, out)
		}
		return err
	}
	return nil
}

func (c *Client) invokeDC(ctx context.Context, dcID int, in bin.Encoder, out bin.Decoder) (err error) {
	c.omux.Lock()
	conn, found := c.others[dcID]
	if !found {
		dcInfo, err := c.lookupDC(dcID)
		if err != nil {
			c.omux.Unlock()
			return err
		}

		conn, err = c.dc(dcInfo).WithAuthTransfer().Connect(ctx)
		if err != nil {
			c.omux.Unlock()
			return xerrors.Errorf("dial dc %d: %w", dcID, err)
		}

		c.others[dcID] = conn
	}
	c.omux.Unlock()

	return conn.InvokeRaw(ctx, &tg.InvokeWithoutUpdatesRequest{
		Query: nopDecoder{in},
	}, out)
}

type nopDecoder struct {
	bin.Encoder
}

func (n nopDecoder) Decode(b *bin.Buffer) error { panic("unreachable") }
