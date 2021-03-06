package dcmanager

import (
	"crypto/rsa"
	"encoding/pem"
	"fmt"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/tg"
)

func (m *Manager) lookupDC(id int) (tg.DCOption, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	// TODO(ccln): support IPv6?
	return m.cfg.findDC(id, true)
}

func parseCDNKeys(keys ...tg.CDNPublicKey) ([]*rsa.PublicKey, error) {
	r := make([]*rsa.PublicKey, 0, len(keys))

	for _, key := range keys {
		block, _ := pem.Decode([]byte(key.PublicKey))
		if block == nil {
			continue
		}

		key, err := crypto.ParseRSA(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA from PEM: %w", err)
		}

		r = append(r, key)
	}

	return r, nil
}
