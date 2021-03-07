package dcmanager

import (
	"context"

	"github.com/gotd/td/mtproto"
	"go.uber.org/zap"
)

func (m *Manager) initWithoutConfig(ctx context.Context, addr string) error {
	conn := m.createConn(addr, mtproto.Options{
		//Transport: m.transport,
		//Network:   m.network,

		Logger: m.log.With(zap.String("dc_addr", addr), zap.String("dc_type", "primary")),
	})

	m.runConn(ctx, conn)

	cfg, err := m.initConn(ctx, conn, false)
	if err != nil {
		return err
	}

	m.cfg.TGConfig = cfg
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

func (m *Manager) runConn(ctx context.Context, conn Conn) {
	m.g.Go(func() error {
		return conn.Run(ctx, func(ctx context.Context) error {
			// Dumb function.
			<-ctx.Done()
			return ctx.Err()
		})
	})
}
