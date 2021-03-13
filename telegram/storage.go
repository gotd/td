package telegram

import (
	"context"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/session"
)

func (c *Client) storageLoad(ctx context.Context) error {
	if c.storage == nil {
		return session.ErrNotFound
	}

	data, err := c.storage.Load(ctx)
	if err != nil {
		return err
	}

	c.sess = mtproto.Session{
		// TODO(ccln): Save session ID.
		ID: 0,
		Key: crypto.AuthKey{
			Value: data.AuthKey,
			ID:    data.AuthKeyID,
		},
		Salt: data.Salt,
	}

	c.primaryDC = data.DC
	c.addr = data.Addr
	c.cfg = data.Config
	return nil
}

func (c *Client) storageSave() error {
	if c.storage == nil {
		return nil
	}

	cfg := c.cfg
	primaryDC := c.primaryDC
	addr := c.addr
	sess := c.sess

	return c.storage.Save(c.ctx, &session.Data{
		Config:    cfg,
		DC:        primaryDC,
		Addr:      addr,
		AuthKey:   sess.Key.Value,
		AuthKeyID: sess.Key.ID,
		Salt:      sess.Salt,
	})
}
