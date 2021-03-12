package dcmanager

import (
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/tg"
	"golang.org/x/xerrors"
)

type Config struct {
	TGConfig  tg.Config      `json:"tg_config"`
	PrimaryDC int            `json:"primary_dc"`
	AuthKey   crypto.AuthKey `json:"auth_key"`
	Salt      int64          `json:"salt"`
}

func (cfg Config) findDC(id int, noIPv6 bool) (tg.DCOption, error) {
	for _, dc := range cfg.TGConfig.DCOptions {
		if noIPv6 && dc.Ipv6 {
			continue
		}

		if dc.ID == id {
			return dc, nil
		}
	}

	return tg.DCOption{}, xerrors.Errorf("dc not found in config: %d", id)
}

// func (cfg Config) findDCIndex(addr string) (int, error) {
// 	for _, dc := range cfg.TGConfig.DCOptions {
// 		if dc.IPAddress == addr {
// 			return dc.ID, nil
// 		}
// 	}

// 	return 0, xerrors.Errorf("dc not found in config: %s", addr)
// }

func (m *Manager) UpdateConfig(cfg Config) {
	m.cfgMux.Lock()
	defer m.cfgMux.Unlock()
	m.cfg = cfg
}
