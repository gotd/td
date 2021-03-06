package dcmanager

import (
	"context"

	"github.com/gotd/td/dcmanager/mtp"
	"github.com/gotd/td/mtproto"
	"go.uber.org/zap"
)

func (m *Manager) initWithoutConfig(addr string) error {
	mtconn, err := mtp.New(addr, mtproto.Options{
		//Transport: m.transport,
		//Network:   m.network,

		Logger: m.log.With(zap.String("dc_addr", addr), zap.String("dc_type", "primary")),
	})
	if err != nil {
		return err
	}

	cfg, err := m.initConn(context.TODO(), mtconn, false)
	if err != nil {
		_ = mtconn.Close()
		return err
	}

	m.cfg.TGConfig = cfg
	m.primary = mtconn
	return nil
}

func (m *Manager) initWithConfig(cfg Config) error {
	m.cfg = cfg

	dcInfo, err := cfg.findDC(m.cfg.PrimaryDC, true)
	if err != nil {
		return err
	}

	mtconn, err := m.dc(dcInfo).
		AsPrimary().
		WithCreds(cfg.AuthKey, cfg.Salt).
		Connect(context.TODO())
	if err != nil {
		return err
	}

	m.primary = mtconn
	return nil
}
