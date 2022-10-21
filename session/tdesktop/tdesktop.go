// Package tdesktop contains Telegram Desktop session decoder.
package tdesktop

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/go-faster/errors"
)

// Account is a Telegram user account representation in Telegram Desktop storage.
type Account struct {
	// IDx is an internal Telegram Desktop account ID.
	IDx uint32
	// Authorization contains Telegram user and MTProto sessions.
	Authorization MTPAuthorization
	// Config contains Telegram config.
	Config MTPConfig
}

// Read reads accounts info from given Telegram Desktop tdata root.
// Shorthand for:
//
//	ReadFS(os.DirFS(root), passcode)
func Read(root string, passcode []byte) ([]Account, error) {
	return ReadFS(os.DirFS(root), passcode)
}

// ReadFS reads Telegram Desktop accounts info from given FS root.
func ReadFS(root fs.FS, passcode []byte) ([]Account, error) {
	keyDataFile, err := open(root, "key_data")
	if err != nil {
		return nil, errors.Wrap(err, "open key_data")
	}

	kd, err := readKeyData(keyDataFile, passcode)
	if err != nil {
		return nil, err
	}
	if len(kd.accountsIDx) < 1 {
		return nil, ErrNoAccounts
	}

	r := make([]Account, 0, len(kd.accountsIDx))
	for _, account := range kd.accountsIDx {
		var keyFile = fileKey("data")
		if account > 0 {
			keyFile = fileKey(fmt.Sprintf("data#%d", account+1))
		}

		mtpDataFile, err := open(root, keyFile)
		if err != nil {
			return nil, errors.Wrap(err, "open key_data")
		}

		mtpData, err := readMTPData(mtpDataFile, kd.localKey)
		if err != nil {
			return nil, errors.Wrap(err, "read mtp")
		}

		a := Account{
			IDx:           account,
			Authorization: mtpData,
		}
		mtpConfigFile, err := open(root, path.Join(keyFile, "config"))
		if err == nil {
			mtpConfig, err := readMTPConfig(mtpConfigFile, kd.localKey)
			// HACK: ignoring error, because config is optional.
			if err == nil {
				a.Config = mtpConfig
			}
		}

		r = append(r, a)
	}

	return r, nil
}
