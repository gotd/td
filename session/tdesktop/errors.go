package tdesktop

import (
	"fmt"

	"github.com/go-faster/errors"
)

// WrongMagicError is returned when tdesktop data file
// has wrong magic header.
type WrongMagicError struct {
	Magic [4]byte
}

// Error implements error.
func (w *WrongMagicError) Error() string {
	return fmt.Sprintf("wrong magic %+v", w.Magic)
}

var (
	// ErrKeyInfoDecrypt is returned when key data decrypt fails.
	// It can happen if passed passcode is wrong.
	ErrKeyInfoDecrypt = errors.New("key data decrypt")
	// ErrNoAccounts reports that decoded tdata does not contain any accounts info.
	ErrNoAccounts = errors.New("tdesktop data does not contain accounts")
)
