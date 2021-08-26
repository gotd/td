// Package tdesktop contains Telegram Desktop session decoder.
package tdesktop

import (
	"fmt"
	"io/fs"
	"os"

	"golang.org/x/xerrors"
)

// Account is a Telegram user account representation in Telegram Desktop storage.
type Account struct {
	// IDx is an internal Telegram Desktop account ID.
	IDx uint32
	// Authorization contains Telegram user and MTProto sessions.
	Authorization MTPAuthorization
}

// Read reads accounts info from given Telegram Desktop tdata root.
// Shorthand for:
//
//	ReadFS(os.DirFS(root), passcode)
//
func Read(root string, passcode []byte) ([]Account, error) {
	return ReadFS(os.DirFS(root), passcode)
}

// ReadFS reads Telegram Desktop accounts info from given FS root.
func ReadFS(root fs.FS, passcode []byte) ([]Account, error) {
	keyDataFile, err := open(root, "key_data")
	if err != nil {
		return nil, xerrors.Errorf("open key_data: %w", err)
	}

	kd, err := readKeyData(keyDataFile, passcode)
	if err != nil {
		return nil, err
	}
	if len(kd.accountsIDx) < 1 {
		return nil, xerrors.New("tdesktop data does not contain accounts")
	}

	r := make([]Account, 0, len(kd.accountsIDx))
	for _, account := range kd.accountsIDx {
		var keyFile = fileKey("data")
		if account > 0 {
			keyFile = fileKey(fmt.Sprintf("data#%d", account+1))
		}

		tgf, err := open(root, keyFile)
		if err != nil {
			return nil, xerrors.Errorf("open key_data: %w", err)
		}

		mtp, err := readMTPData(tgf, kd.localKey)
		if err != nil {
			return nil, xerrors.Errorf("read mtp: %w", err)
		}

		r = append(r, Account{
			IDx:           account,
			Authorization: mtp,
		})
	}

	return r, nil
}
