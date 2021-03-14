package mtproto

import (
	"context"
	"sync/atomic"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mt"
)

func (c *Conn) updateSalt() {
	salt, ok := c.salts.Get(5 * time.Minute)
	if !ok {
		return
	}
	atomic.StoreInt64(&c.salt, salt)
}

const defaultSaltsNum = 64

func (c *Conn) getSalts(ctx context.Context) error {
	request := &mt.GetFutureSaltsRequest{
		Num: defaultSaltsNum,
	}
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout(request.TypeID()))
	defer cancel()

	if err := c.write(ctx, c.newMessageID(), c.seqNo(false), request); err != nil {
		return xerrors.Errorf("request salts: %w", err)
	}

	return nil
}

func (c *Conn) saltLoop(ctx context.Context) error {
	select {
	case <-c.gotSession.Ready():
	case <-ctx.Done():
		return ctx.Err()
	}

	// Get salts first time.
	if err := c.getSalts(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-c.clock.After(c.saltFetchInterval):
			if err := c.getSalts(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
