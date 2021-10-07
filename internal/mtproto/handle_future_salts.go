package mtproto

import (
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
)

func (c *Conn) handleFutureSalts(b *bin.Buffer) error {
	var res mt.FutureSalts

	if err := res.Decode(b); err != nil {
		return xerrors.Errorf("error decode: %w", err)
	}

	c.salts.Store(res.Salts)

	serverTime := time.Unix(int64(res.Now), 0)
	c.log.Debug("Got future salts", zap.Time("server_time", serverTime))
	return nil
}
