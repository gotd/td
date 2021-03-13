package telegram

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

func (c *Client) connectPrimary(ctx context.Context, dc tg.DCOption, reuseCreds bool) error {
	c.pmux.Lock()
	defer c.pmux.Unlock()

	switch {
	case dc.TCPObfuscatedOnly:
		return xerrors.Errorf("can't use tcpo DC as primary (%d)", dc.ID)
	case dc.TCPObfuscatedOnly:
		return xerrors.Errorf("can't use media-only DC as primary (%d)", dc.ID)
	case dc.CDN:
		return xerrors.Errorf("cdn could not be a primary DC (%d)", dc.ID)
	}

	dcBuilder := c.dc(dc).
		WithMessageHandler(c.onPrimaryMessage).
		WithSessionHandler(c.onPrimarySession)

	if reuseCreds {
		dcBuilder = dcBuilder.WithCreds(c.primaryCreds())
	}

	conn, err := dcBuilder.Connect(ctx)
	if err != nil {
		return err
	}

	cfg, err := tg.NewClient(conn).HelpGetConfig(ctx)
	if err != nil {
		return err
	}

	if c.primary != nil {
		// Cleanup previous connection.
		if err := c.lf.Stop(c.primary); err != nil {
			c.log.Warn("Failed to cleanup connection", zap.Error(err))
			return err
		}
	}

	// TODO(ccln): Recheck cfg ID.
	c.primaryDC = dc.ID
	c.primary = conn

	return c.onPrimaryConfig(*cfg)
}

func (c *Client) onPrimarySession(session mtproto.Session) error {
	c.dataMux.Lock()
	defer c.dataMux.Unlock()
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
	c.dataMux.Lock()
	defer c.dataMux.Unlock()
	c.cfg = cfg
	return c.storageSave()
}
