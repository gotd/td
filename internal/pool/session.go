package pool

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
)

// SessionStorage represents MTProto session storage.
type SessionStorage interface {
	Load(ctx context.Context) (*session.Data, error)
	Save(ctx context.Context, data *session.Data) error
}

func (c *Pool) saveSession(addr string, cfg tg.Config, s mtproto.Session) error {
	if c.storage == nil {
		return nil
	}

	data, err := c.storage.Load(c.ctx)
	if errors.Is(err, session.ErrNotFound) {
		// Initializing new state.
		err = nil
		data = &session.Data{}
	}
	if err != nil {
		return xerrors.Errorf("load: %w", err)
	}

	// Updating previous data.
	data.Config = cfg
	data.AuthKey = s.Key.Value[:]
	data.AuthKeyID = s.Key.ID[:]
	data.DC = cfg.ThisDC
	data.Addr = addr
	data.Salt = s.Salt

	if err := c.storage.Save(c.ctx, data); err != nil {
		return xerrors.Errorf("save: %w", err)
	}

	c.log.Debug("Data saved",
		zap.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)
	return nil
}
