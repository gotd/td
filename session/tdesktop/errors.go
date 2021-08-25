package tdesktop

import (
	"errors"
	"fmt"
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

// ErrKeyInfoDecrypt is returned when key data decrypt fails.
// It can happen if passed passcode is wrong.
var ErrKeyInfoDecrypt = errors.New("key data decrypt")
