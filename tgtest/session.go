package tgtest

import (
	"encoding/hex"

	"go.uber.org/zap/zapcore"

	"github.com/nnqq/td/internal/crypto"
)

// Session represents connection session.
type Session struct {
	// ID is a Session ID.
	ID int64
	// AuthKey is an attached key.
	AuthKey crypto.AuthKey
}

// MarshalLogObject implements zap.ObjectMarshaler.
func (s Session) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddInt64("session_id", s.ID)
	encoder.AddString("key_id", hex.EncodeToString(s.AuthKey.ID[:]))
	return nil
}
