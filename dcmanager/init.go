package dcmanager

import (
	"context"

	"github.com/gotd/td/internal/mtproto/reliable"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

func (m *Manager) initWithoutConfig(ctx context.Context, addr string) error {
	m.pmux.Lock()
	defer m.pmux.Unlock()

	opts := m.mtopts
	opts.MessageHandler = m.onMessage
	opts.SessionHandler = m.onPrimarySessionUpdate
	opts.Logger = m.log.With(zap.String("dc_addr", addr), zap.String("dc_type", "primary"))

	conn := reliable.New(reliable.Config{
		Addr:   addr,
		MTOpts: opts,
		OnConnected: func(conn reliable.MTConn) error {
			_, err := m.initConn(context.TODO(), conn, false)
			return err
		},
	})
	if err := m.conns.Start(conn); err != nil {
		return err
	}

	cfg, err := tg.NewClient(conn).HelpGetConfig(ctx)
	if err != nil {
		return err
	}

	m.cfg.TGConfig = *cfg
	m.primary = conn
	return nil
}

func (m *Manager) initWithConfig(ctx context.Context) error {
	dcInfo, err := m.cfg.findDC(m.cfg.PrimaryDC, true)
	if err != nil {
		return err
	}

	conn, err := m.dc(dcInfo).
		AsPrimary().
		WithCreds(m.cfg.AuthKey, m.cfg.Salt).
		Connect(ctx)
	if err != nil {
		return err
	}

	m.primary = conn
	return nil
}
