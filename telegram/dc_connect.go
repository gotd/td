package telegram

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/mtproto/reliable"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

type dcBuilder struct {
	dc       tg.DCOption
	transfer bool // Export authorization from primary DC.

	// For primary connection migration.
	asPrimary bool
	authKey   crypto.AuthKey
	salt      int64

	c *Client
}

func (c *Client) dc(dc tg.DCOption) *dcBuilder {
	return &dcBuilder{
		dc: dc,
		c:  c,
	}
}

func (b *dcBuilder) AsPrimary() *dcBuilder {
	b.asPrimary = true
	return b
}

func (b *dcBuilder) WithAuthTransfer() *dcBuilder {
	b.transfer = true
	return b
}

func (b *dcBuilder) WithCreds(key crypto.AuthKey, salt int64) *dcBuilder {
	b.authKey = key
	b.salt = salt
	return b
}

func (b *dcBuilder) Connect(ctx context.Context) (conn, error) {
	var (
		c         = b.c
		dc        = b.dc
		asPrimary = b.asPrimary
	)

	c.log.Info("Connecting",
		zap.Any("dc_info", dc),
		zap.Bool("transfer", b.transfer),
		zap.Bool("as_primary", asPrimary),
	)

	if asPrimary {
		if dc.TCPObfuscatedOnly {
			return nil, xerrors.Errorf("can't migrate to obfuscated transport only DC %d", dc.ID)
		}

		if dc.MediaOnly {
			return nil, xerrors.Errorf("can't migrate to Media-only DC %d", dc.ID)
		}

		if dc.CDN {
			return nil, xerrors.Errorf("CDN could not be a primary DC %d", dc.ID)
		}
	}

	if dc.CDN {
		b.transfer = false
	}

	log := c.log.With(zap.Int("dc_id", dc.ID))
	switch {
	case dc.CDN:
		log = log.With(zap.String("dc_type", "cdn"))
	case asPrimary:
		log = log.With(zap.String("dc_type", "primary"))
	default:
		log = log.With(zap.String("dc_type", "data"))
	}

	var (
		opts       = b.c.opts
		gotSession = make(chan struct{})
		once       sync.Once
	)

	opts.Key = b.authKey
	opts.Salt = b.salt
	opts.Logger = log

	if asPrimary {
		opts.MessageHandler = c.onPrimaryMessage
		opts.SessionHandler = c.onPrimarySession

		c.pmux.Lock()
		defer c.pmux.Unlock()
	}

	if !dc.CDN && b.transfer {
		opts.SessionHandler = func(sess mtproto.Session) error {
			once.Do(func() { close(gotSession) })
			return nil
		}

		if asPrimary {
			opts.SessionHandler = func(sess mtproto.Session) error {
				once.Do(func() { close(gotSession) })
				return c.onPrimarySession(sess)
			}
		}
	} else if dc.CDN {
		cdnCfg, err := tg.NewClient(c.primary).HelpGetCDNConfig(ctx)
		if err != nil {
			return nil, xerrors.Errorf("get CDN config: %w", err)
		}

		keys, err := parseCDNKeys(cdnCfg.PublicKeys...)
		if err != nil {
			return nil, xerrors.Errorf("parse CDN keys: %w", err)
		}

		opts.PublicKeys = keys
		// Zero key for CDN.
		opts.Key = crypto.AuthKey{}
		opts.Salt = 0
	}

	conn := reliable.New(reliable.Config{
		Addr:   fmt.Sprintf("%d|%s:%d", dc.ID, dc.IPAddress, dc.Port),
		MTOpts: opts,
		OnConnected: func(conn reliable.MTConn) error {
			_, err := c.initConn(c.ctx, conn, !asPrimary)
			return err
		},
	})

	if err := c.lf.Start(conn); err != nil {
		return nil, err
	}

	if err := func() error {
		cfg, err := tg.NewClient(conn).HelpGetConfig(ctx)
		if err != nil {
			return err
		}

		if !dc.CDN && b.transfer {
			if err := transfer(ctx, c.primary, conn, dc.ID); err != nil {
				return xerrors.Errorf("transfer: %w", err)
			}

			select {
			case <-gotSession:
				break
			case <-time.After(time.Second * 10):
				return xerrors.Errorf("session timeout")
			}
		}

		if asPrimary {
			c.primaryDC = dc.ID

			// TODO(ccln): recheck cfg dc id
			if err := c.onPrimaryConfig(*cfg); err != nil {
				return err
			}

			if err := c.lf.Stop(c.primary); err != nil {
				c.log.Warn("Failed to cleanup connection", zap.Error(err))
			}
			c.primary = conn
		}

		return nil
	}(); err != nil {
		c.log.Warn("Failed to initialize connection", zap.Error(err))
		if err := c.lf.Stop(conn); err != nil {
			c.log.Warn("Failed to cleanup connection", zap.Error(err))
		}

		return nil, err
	}

	return conn, nil
}

func (c *Client) initConn(ctx context.Context, conn conn, noUpdates bool) (tg.Config, error) {
	wrap := func(req bin.Object) bin.Object { return req }
	if noUpdates {
		wrap = func(req bin.Object) bin.Object {
			return &tg.InvokeWithoutUpdatesRequest{
				Query: req,
			}
		}
	}

	q := wrap(&tg.InitConnectionRequest{
		APIID:          c.appID,
		DeviceModel:    c.device.DeviceModel,
		SystemVersion:  c.device.SystemVersion,
		AppVersion:     c.device.AppVersion,
		SystemLangCode: c.device.SystemLangCode,
		LangPack:       c.device.LangPack,
		LangCode:       c.device.LangCode,
		Query:          wrap(&tg.HelpGetConfigRequest{}),
	})

	var cfg tg.Config
	if err := conn.InvokeRaw(ctx, wrap(&tg.InvokeWithLayerRequest{
		Layer: tg.Layer,
		Query: q,
	}), &cfg); err != nil {
		return cfg, xerrors.Errorf("invoke: %w", err)
	}

	return cfg, nil
}
