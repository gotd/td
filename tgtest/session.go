package tgtest

import (
	"encoding/hex"

	"go.uber.org/atomic"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/td/internal/crypto"
)

// Session represents connection session.
type Session struct {
	// ID is a Session ID.
	ID int64
	// AuthKey is a attached key.
	AuthKey crypto.AuthKey
	// Layer is Telegram schema layer.
	// NB: may be zero (not set).
	Layer atomic.Int32
}

// MarshalLogObject implements zap.ObjectMarshaler.
func (s Session) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddInt64("session_id", s.ID)
	encoder.AddString("key_id", hex.EncodeToString(s.AuthKey.ID[:]))
	if l := s.Layer.Load(); l != 0 {
		encoder.AddInt32("layer", l)
	}
	return nil
}
