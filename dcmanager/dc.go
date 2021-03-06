package dcmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/dcmanager/mtp"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/mtproto"
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

	m *Manager
}

func (m *Manager) dc(dc tg.DCOption) *dcBuilder {
	return &dcBuilder{
		dc: dc,
		m:  m,
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

func (b *dcBuilder) Connect(ctx context.Context) (*mtp.Conn, error) {
	var (
		m         = b.m
		dc        = b.dc
		asPrimary = b.asPrimary
	)

	m.log.Info("Connecting",
		zap.Any("dc_info", dc),
		zap.Bool("transfer", b.transfer),
		zap.Bool("as_primary", asPrimary),
	)

	if dc.CDN && asPrimary {
		return nil, xerrors.Errorf("CDN could not be a primary DC")
	}

	if dc.CDN {
		b.transfer = false
	}

	log := m.log.With(zap.Int("dc_id", dc.ID))
	switch {
	case dc.CDN:
		log = log.With(zap.String("dc_type", "cdn"))
	case asPrimary:
		log = log.With(zap.String("dc_type", "primary"))
	default:
		log = log.With(zap.String("dc_type", "data"))
	}

	var (
		opts = mtproto.Options{
			//Transport: m.transport,
			//Network:   m.network,
			Key:    b.authKey,
			Salt:   b.salt,
			Logger: log,
		}

		gotSession = make(chan struct{})
		once       sync.Once
	)

	if asPrimary {
		opts.MessageHandler = m.onMessage
	}

	if !dc.CDN && b.transfer {
		// Setup session handlers for transfer result check.
		opts.SessionHandler = func(sess mtproto.Session) error {
			once.Do(func() { close(gotSession) })
			return nil
		}

		if asPrimary {
			opts.SessionHandler = func(sess mtproto.Session) error {
				once.Do(func() { close(gotSession) })
				return m.onPrimarySessionUpdate(sess)
			}
		}
	} else if dc.CDN {
		m.mux.RLock()
		primary := m.primary
		m.mux.RUnlock()

		cdnCfg, err := tg.NewClient(primary).HelpGetCDNConfig(ctx)
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

	mtconn, err := mtp.New(fmt.Sprintf("%d|%s:%d", dc.ID, dc.IPAddress, dc.Port), opts)
	if err != nil {
		return nil, err
	}

	cfg, err := m.initConn(ctx, mtconn, !asPrimary)
	if err != nil {
		_ = mtconn.Close()
		return nil, err
	}

	if !dc.CDN && b.transfer {
		if err := m.transfer(ctx, mtconn, dc.ID); err != nil {
			_ = mtconn.Close()
			return nil, xerrors.Errorf("transfer: %w", err)
		}

		select {
		case <-gotSession:
			break
		case <-time.After(time.Second * 20):
			_ = mtconn.Close()
			return nil, xerrors.Errorf("session timeout")
		}
	}

	if asPrimary {
		// TODO(ccln): recheck cfg dc id
		m.mux.Lock()
		defer m.mux.Unlock()
		m.cfg.PrimaryDC = dc.ID
		m.cfg.TGConfig = cfg

		if err := m.saveConfig(m.cfg); err != nil {
			_ = mtconn.Close()
			return nil, err
		}

		_ = m.primary.Close()
		m.primary = mtconn
	}

	return mtconn, nil
}

func (m *Manager) initConn(ctx context.Context, conn *mtp.Conn, noUpdates bool) (tg.Config, error) {
	wrap := func(req bin.Object) bin.Object { return req }
	if noUpdates {
		wrap = func(req bin.Object) bin.Object {
			return &tg.InvokeWithoutUpdatesRequest{
				Query: req,
			}
		}
	}

	q := wrap(&tg.InitConnectionRequest{
		APIID:          m.appID,
		DeviceModel:    m.device.DeviceModel,
		SystemVersion:  m.device.SystemVersion,
		AppVersion:     m.device.AppVersion,
		SystemLangCode: m.device.SystemLangCode,
		LangPack:       m.device.LangPack,
		LangCode:       m.device.LangCode,
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

func (m *Manager) onPrimarySessionUpdate(sess mtproto.Session) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.cfg.AuthKey = sess.Key
	m.cfg.Salt = sess.Salt
	return m.saveConfig(m.cfg)
}
