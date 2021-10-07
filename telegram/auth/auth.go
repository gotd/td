// Package auth provides authentication on top of tg.Client.
package auth

import (
	"github.com/nnqq/td/tgerr"
)

// IsKeyUnregistered reports whether err is AUTH_KEY_UNREGISTERED error.
func IsKeyUnregistered(err error) bool {
	return tgerr.Is(err, "AUTH_KEY_UNREGISTERED")
}
