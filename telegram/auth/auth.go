package auth

import (
	"github.com/gotd/td/tgerr"
)

func IsKeyUnregistered(err error) bool {
	return tgerr.Is(err, "AUTH_KEY_UNREGISTERED")
}
