package mtproto

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/mt"
)

func (c *Conn) storeSalt(salt int64) {
	c.sessionMux.Lock()
	// Copy to log.
	oldSalt := c.salt
	c.salt = salt
	c.sessionMux.Unlock()

	if salt != oldSalt {
		c.log.Info(context.Background(), "Salt updated", log.Int64("old", oldSalt), log.Int64("new", salt))
		c.traceEvent(WireEvent{Type: WireSaltChange, Old: oldSalt, New: salt})
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
		return errors.Wrap(err, "request salts")
	}

	return nil
}

// saltLoop fetches future salts every saltFetchInterval. It runs only if
// the interval is set.
//
// Nothing is fetched at start: Telegram Desktop never sends
// get_future_salts, it uses the salt of new_session_created and the one
// bad_server_salt brings when the current salt expires.
func (c *Conn) saltLoop(ctx context.Context) error {
	select {
	case <-c.gotSession.Ready():
	case <-ctx.Done():
		return ctx.Err()
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
