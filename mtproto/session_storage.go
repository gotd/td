package mtproto

import (
	"context"
	"errors"
	"sync/atomic"

	"go.uber.org/zap"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/session"
)

// SessionStorage persists session data.
type SessionStorage interface {
	Load(ctx context.Context) (*session.Data, error)
	Save(ctx context.Context, data *session.Data) error
}

func (c *Conn) saveSession(ctx context.Context) error {
	if c.session == nil {
		return nil
	}

	// TODO: fix race condition here
	data, err := c.session.Load(ctx)
	if errors.Is(err, session.ErrNotFound) {
		data = &session.Data{}
	} else if err != nil {
		return xerrors.Errorf("load: %w", err)
	}

	data.Salt = atomic.LoadInt64(&c.salt)
	data.AuthKeyID = c.authKey.AuthKeyID[:]
	data.AuthKey = c.authKey.AuthKey[:]
	data.Config = c.cfg
	data.Addr = c.addr

	if err := c.session.Save(ctx, data); err != nil {
		return xerrors.Errorf("save: %w", err)
	}

	return nil
}

func (c *Conn) loadSession(ctx context.Context) error {
	if c.session == nil {
		return nil
	}

	data, err := c.session.Load(ctx)
	if errors.Is(err, session.ErrNotFound) {
		// Will create session after key exchange.
		c.log.Debug("Session not found", zap.Error(err))
		return nil
	}
	if err != nil {
		return xerrors.Errorf("failed to load session: %w", err)
	}

	// Validating auth key.
	var k crypto.AuthKeyWithID
	copy(k.AuthKey[:], data.AuthKey)
	copy(k.AuthKeyID[:], data.AuthKeyID)

	if k.AuthKey.ID() != k.AuthKeyID {
		return xerrors.New("auth key id does not match (corrupted session data)")
	}

	c.cfg = data.Config
	c.authKey = k
	atomic.StoreInt64(&c.salt, data.Salt)
	c.log.Info("Session loaded from storage")

	// Generating new session id.
	sessID, err := crypto.NewSessionID(c.rand)
	if err != nil {
		return err
	}
	atomic.StoreInt64(&c.sessionID, sessID)

	return nil
}
