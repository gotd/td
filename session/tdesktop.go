package session

import (
	"github.com/ogen-go/errors"

	"github.com/gotd/td/session/tdesktop"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

func findDCAddr(list []tg.DCOption, dcID int) string {
	for _, opt := range list {
		if opt.ID != dcID {
			continue
		}
		if opt.TCPObfuscatedOnly ||
			opt.CDN ||
			opt.MediaOnly {
			continue
		}

		return opt.IPAddress
	}

	return ""
}

// TDesktopSession converts TDesktop's Account to Data.
func TDesktopSession(account tdesktop.Account) (*Data, error) {
	auth := account.Authorization
	cfg := account.Config
	test := cfg.Environment.Test()
	dc := auth.MainDC

	key, ok := auth.Keys[dc]
	if !ok {
		return nil, errors.Errorf("key for main DC (%d) not found", dc)
	}
	keyID := key.ID()

	var list dcs.List
	if !test {
		list = dcs.Prod()
	} else {
		list = dcs.Test()
	}

	addr := findDCAddr(list.Options, dc)
	if addr == "" {
		return nil, errors.Errorf("can't find address for DC %d", dc)
	}

	return &Data{
		DC:        dc,
		Addr:      addr,
		AuthKey:   key[:],
		AuthKeyID: keyID[:],
	}, nil
}
