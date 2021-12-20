// Package auth provides authentication on top of tg.Client.
package auth

import (
	"github.com/gotd/td/tgerr"
)

// IsKeyUnregistered reports whether err is AUTH_KEY_UNREGISTERED error.
//
// Deprecated: use IsUnauthorized.
func IsKeyUnregistered(err error) bool {
	return tgerr.Is(err, "AUTH_KEY_UNREGISTERED")
}

// IsUnauthorized reports whether err is 401 UNAUTHORIZED.
//
// https://core.telegram.org/api/errors#401-unauthorized
func IsUnauthorized(err error) bool {
	return tgerr.IsCode(err, 401)
}
