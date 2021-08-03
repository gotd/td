package crypto

import (
	"crypto/sha1" // #nosec
	"encoding/hex"
	"fmt"

	"go.uber.org/zap/zapcore"
)

// See https://core.telegram.org/mtproto/description#defining-aes-key-and-initialization-vector

// Key represents 2048-bit authorization key value.
type Key [256]byte

func (k Key) String() string {
	// Never print key.
	return "(redacted)"
}

// Zero reports whether Key is zero value.
func (k Key) Zero() bool {
	return k == Key{}
}

// ID returns auth_key_id.
func (k Key) ID() [8]byte {
	raw := sha1.Sum(k[:]) // #nosec
	var id [8]byte
	copy(id[:], raw[12:])
	return id
}

// AuxHash returns aux_hash value of key.
func (k Key) AuxHash() [8]byte {
	raw := sha1.Sum(k[:]) // #nosec
	var id [8]byte
	copy(id[:], raw[0:8])
	return id
}

// WithID creates new AuthKey from Key.
func (k Key) WithID() AuthKey {
	return AuthKey{
		Value: k,
		ID:    k.ID(),
	}
}

// AuthKey is a Key with cached id.
type AuthKey struct {
	Value Key
	ID    [8]byte
}

// Zero reports whether Key is zero value.
func (a AuthKey) Zero() bool {
	return a == AuthKey{}
}

// String implements fmt.Stringer.
func (a AuthKey) String() string {
	return fmt.Sprintf("Key(id: %x)", a.ID)
}

// MarshalLogObject implements zap.ObjectMarshaler.
func (a AuthKey) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("id", hex.EncodeToString(a.ID[:]))
	return nil
}
