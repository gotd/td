package mtproto

import (
	"context"
	"sort"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mt"
)

func (c *Conn) getValidSalt() (int64, bool) {
	c.saltsMux.Lock()
	defer c.saltsMux.Unlock()

	// Sort slice by valid until.
	sort.SliceStable(c.salts, func(i, j int) bool {
		return c.salts[i].ValidUntil < c.salts[j].ValidUntil
	})

	// Filter (in place) from SliceTricks.
	n := 0
	dedup := map[int64]struct{}{}
	// Check that the salt will be valid next 5 minute.
	date := int(time.Now().Add(5 * time.Minute).Unix())
	for _, salt := range c.salts {
		// Filter expired salts.
		if _, ok := dedup[salt.Salt]; !ok && salt.ValidUntil > date {
			dedup[salt.Salt] = struct{}{}
			c.salts[n] = salt
			n++
		}
	}
	c.salts = c.salts[:n]

	if len(c.salts) < 1 {
		return 0, false
	}
	return c.salts[0].Salt, true
}

func (c *Conn) updateSalt() {
	salt, ok := c.getValidSalt()
	if !ok {
		return
	}

	c.sessionMux.Lock()
	c.salt = salt
	c.sessionMux.Unlock()
}

const defaultSaltsNum = 64

func (c *Conn) getSalts(ctx context.Context) error {
	request := &mt.GetFutureSaltsRequest{
		Num: defaultSaltsNum,
	}
	ctx, cancel := context.WithTimeout(context.Background(),
		c.getTimeout(request.TypeID()),
	)
	defer cancel()

	if err := c.write(ctx, c.newMessageID(), c.seqNo(true), request); err != nil {
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
