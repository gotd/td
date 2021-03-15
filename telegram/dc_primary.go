package telegram

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/tg"
)

func (c *Client) connectPrimary(ctx context.Context, dc tg.DCOption, reuseCreds bool) error {
	c.pmux.Lock()
	defer c.pmux.Unlock()

	switch {
	case dc.TCPObfuscatedOnly:
		return xerrors.Errorf("can't use tcpo DC as primary (%d)", dc.ID)
	case dc.MediaOnly:
		return xerrors.Errorf("can't use media-only DC as primary (%d)", dc.ID)
	case dc.CDN:
		return xerrors.Errorf("cdn could not be a primary DC (%d)", dc.ID)
	}

	cfgchan := make(chan tg.Config, 1)
	dcBuilder := c.dc(dc).
		WithMessageHandler(c.onPrimaryMessage).
		WithSessionHandler(c.onPrimarySession).
		OnConfig(func(c tg.Config) { cfgchan <- c })

	if reuseCreds {
		dcBuilder = dcBuilder.WithCreds(c.sess.Key, c.sess.Salt)
	}

	conn, err := dcBuilder.Connect(ctx)
	if err != nil {
		return err
	}

	if c.primary != nil {
		// Cleanup previous connection.
		if err := c.lf.Stop(c.primary); err != nil {
			c.log.Warn("Failed to cleanup connection", zap.Error(err))
		}
	}

	var cfg tg.Config
	select {
	case c := <-cfgchan:
		cfg = c
	case <-c.clock.After(time.Second * 5):
		if err := c.lf.Stop(conn); err != nil {
			c.log.Warn("Failed to cleanup connection", zap.Error(err))
		}

		return xerrors.Errorf("Config timeout")
	}

	c.primary = conn
	c.primaryDC = cfg.ThisDC
	c.cfg = cfg

	return c.storageSave()
}

func (c *Client) onPrimarySession(session mtproto.Session) error {
	c.pmux.Lock()
	defer c.pmux.Unlock()
	c.sess = session
	return c.storageSave()
}

func (c *Client) onPrimaryMessage(b *bin.Buffer) error {
	updates, err := tg.DecodeUpdates(b)
	if err != nil {
		return xerrors.Errorf("decode updates: %w", err)
	}

	return c.processUpdates(updates)
}

func (c *Client) onPrimaryConfig(cfg tg.Config) error {
	c.pmux.Lock()
	defer c.pmux.Unlock()
	c.cfg = cfg
	return c.storageSave()
}
