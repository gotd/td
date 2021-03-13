package telegram

import (
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/tg"
	"golang.org/x/xerrors"
)

// getVersion optimistically gets current client version.
//
// Does not handle replace directives.
func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	// Hard-coded package name. Probably we can generate this via parsing
	// the go.mod file.
	const pkg = "github.com/gotd/td"
	for _, d := range info.Deps {
		if strings.HasPrefix(d.Path, pkg) {
			return d.Version
		}
	}
	return ""
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

func (c *Client) lookupDC(id int) (tg.DCOption, error) {
	c.pmux.RLock()
	defer c.pmux.RUnlock()

	for _, dc := range c.cfg.DCOptions {
		// TODO(ccln): support IPv6?
		if dc.Ipv6 {
			continue
		}

		if dc.ID == id {
			return dc, nil
		}
	}

	return tg.DCOption{}, xerrors.Errorf("dc not found in config: %d", id)
}

func (c *Client) rpc() *tg.Client {
	return tg.NewClient(c)
}
