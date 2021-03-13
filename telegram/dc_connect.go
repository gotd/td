package telegram

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mtproto"
	"github.com/gotd/td/internal/mtproto/reliable"
	"github.com/gotd/td/tg"
)

type dcBuilder struct {
	dc       tg.DCOption
	transfer bool // Export authorization from primary DC.

	// For primary connection.
	onMessage func(b *bin.Buffer) error
	onSession func(sess mtproto.Session) error
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

func (b *dcBuilder) WithMessageHandler(h func(b *bin.Buffer) error) *dcBuilder {
	b.onMessage = h
	return b
}

func (b *dcBuilder) WithSessionHandler(h func(sess mtproto.Session) error) *dcBuilder {
	b.onSession = h
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
		noUpdates = b.onMessage == nil
		opts      = c.opts
		log       = c.log.With(zap.Int("dc_id", dc.ID), zap.Bool("with_updates", !noUpdates))

		gotSession = make(chan struct{}) // Session transfer check.
		once       sync.Once             // Session transfer check.
	)

	if attrs := dcAttrs(dc); len(attrs) > 0 {
		log = log.With(zap.Strings("dc_attributes", attrs))
	}

	// Overwrite options.
	opts.MessageHandler = b.onMessage
	opts.SessionHandler = b.onSession
	opts.Key = b.authKey
	opts.Salt = b.salt
	opts.Logger = log

	if dc.CDN {
		b.transfer = false
		cdnCfg, err := c.tg.HelpGetCDNConfig(ctx)
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

	if b.transfer {
		opts.SessionHandler = func(sess mtproto.Session) error {
			once.Do(func() { close(gotSession) })
			return b.onSession(sess)
		}
	}

	conn := reliable.New(reliable.Config{
		Addr:   fmt.Sprintf("%d|%s:%d", dc.ID, dc.IPAddress, dc.Port),
		MTOpts: opts,
		OnConnected: func(conn reliable.MTConn) error {
			_, err := c.initConn(c.ctx, conn, noUpdates)
			return err
		},
	})

	if err := c.lf.Start(conn); err != nil {
		return nil, err
	}

	if b.transfer {
		if err := func() error {
			if err := transfer(ctx, c.tg, tg.NewClient(conn), dc.ID); err != nil {
				return err
			}

			select {
			case <-gotSession:
				return nil
			case <-time.After(time.Second * 10):
				return xerrors.Errorf("session timeout")
			}
		}(); err != nil {
			c.pmux.RLock()
			primaryDC := c.primaryDC
			c.pmux.RUnlock()

			c.log.Warn("Failed to transfer auth",
				zap.Int("from", primaryDC),
				zap.Int("to", dc.ID),
				zap.Error(err),
			)

			if err := c.lf.Stop(conn); err != nil {
				c.log.Warn("Failed to cleanup connection", zap.Error(err))
			}

			return nil, err
		}
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

func dcAttrs(dc tg.DCOption) (attrs []string) {
	switch {
	case dc.CDN:
		attrs = append(attrs, "cdn")
		fallthrough
	case dc.MediaOnly:
		attrs = append(attrs, "media_only")
		fallthrough
	case dc.Static:
		attrs = append(attrs, "static")
		fallthrough
	case dc.TCPObfuscatedOnly:
		attrs = append(attrs, "tcpo")
	}
	return
}
