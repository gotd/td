package mtproto

import (
	"sync/atomic"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
)

func (c *Conn) handleSessionCreated(b *bin.Buffer) error {
	var s mt.NewSessionCreated
	if err := s.Decode(b); err != nil {
		return xerrors.Errorf("decode: %w", err)
	}
	c.gotSession.Signal()
	c.log.Debug("Session created",
		zap.Int64("unique_id", s.UniqueID),
		zap.Int64("first_msg_id", s.FirstMsgID),
	)

	atomic.StoreInt64(&c.salt, s.ServerSalt)
	if err := c.handler.OnSession(c.session()); err != nil {
		return xerrors.Errorf("handler.OnSession: %w", err)
	}
	return nil
}
