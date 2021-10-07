package mtproto

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/mt"
)

func (c *Conn) storeSalt(salt int64) {
	c.sessionMux.Lock()
	// Copy to log.
	oldSalt := c.salt
	c.salt = salt
	c.sessionMux.Unlock()

	if salt != oldSalt {
		c.log.Info("Salt updated", zap.Int64("old", oldSalt), zap.Int64("new", salt))
	}
}

func (c *Conn) updateSalt() {
	salt, ok := c.salts.Get(c.clock.Now().Add(time.Minute * 5))
	if !ok {
		return
	}

	c.storeSalt(salt)
}

const defaultSaltsNum = 4

func (c *Conn) getSalts(ctx context.Context) error {
	request := &mt.GetFutureSaltsRequest{
		Num: defaultSaltsNum,
	}
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout(request.TypeID()))
	defer cancel()

	if err := c.writeServiceMessage(ctx, request); err != nil {
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

	ticker := c.clock.Ticker(c.saltFetchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C():
			if err := c.getSalts(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
