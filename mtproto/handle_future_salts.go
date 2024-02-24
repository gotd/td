package mtproto

import (
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
)

func (c *Conn) handleFutureSalts(b *bin.Buffer) error {
	var res mt.FutureSalts

	if err := res.Decode(b); err != nil {
		return errors.Wrap(err, "error decode")
	}

	c.salts.Store(res.Salts)

	serverTime := time.Unix(int64(res.Now), 0)
	c.log.Debug("Got future salts", zap.Time("server_time", serverTime))
	return nil
}
