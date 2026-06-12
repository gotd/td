package tgtest

import (
	"encoding/hex"

	"github.com/gotd/log"

	"github.com/gotd/td/crypto"
)

// Session represents connection session.
type Session struct {
	// ID is a Session ID.
	ID int64
	// AuthKey is an attached key.
	AuthKey crypto.AuthKey
}

// LogAttr returns the session as an inline log group.
func (s Session) LogAttr() log.Attr {
	return log.Group("",
		log.Int64("session_id", s.ID),
		log.String("key_id", hex.EncodeToString(s.AuthKey.ID[:])),
	)
}
