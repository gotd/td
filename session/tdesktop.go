package session

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/session/tdesktop"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/tg"
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
func TDesktopSession(a tdesktop.Account) (*Data, error) {
	auth := a.Authorization
	dc := auth.MainDC

	key, ok := auth.Keys[dc]
	if !ok {
		return nil, xerrors.Errorf("key for main DC (%d) not found", dc)
	}
	keyID := key.ID()

	// TODO(tdakkota): distinguish test and production accounts.
	addr := findDCAddr(dcs.Prod().Options, dc)
	if addr == "" {
		return nil, xerrors.Errorf("can't find address for DC %d", dc)
	}

	return &Data{
		DC:        dc,
		Addr:      addr,
		AuthKey:   key[:],
		AuthKeyID: keyID[:],
	}, nil
}
