package crypto

import (
	"crypto/sha1" // #nosec
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
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

// DecodeJSON decode AuthKey from object with base64-encoded key and integer ID.
func (a *AuthKey) DecodeJSON(d *jx.Decoder) error {
	return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
		switch string(key) {
		case "value":
			data, err := d.Base64()
			if err != nil {
				return errors.Wrap(err, "decode value")
			}
			copy(a.Value[:], data)
		case "id":
			id, err := d.Int64()
			if err != nil {
				return errors.Wrap(err, "decode id")
			}
			a.SetIntID(id)
		default:
			return d.Skip()
		}

		return nil
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AuthKey) UnmarshalJSON(data []byte) error {
	return a.DecodeJSON(jx.DecodeBytes(data))
}

// EncodeJSON encodes AuthKey as object with base64-encoded key and integer ID.
func (a AuthKey) EncodeJSON(e *jx.Encoder) error {
	e.ObjStart()
	e.FieldStart("value")
	e.Base64(a.Value[:])
	e.FieldStart("id")
	e.Int64(a.IntID())
	e.ObjEnd()
	return nil
}

// MarshalJSON implements json.Marshaler.
func (a AuthKey) MarshalJSON() ([]byte, error) {
	e := jx.GetEncoder()
	if err := a.EncodeJSON(e); err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

// Zero reports whether Key is zero value.
func (a AuthKey) Zero() bool {
	return a == AuthKey{}
}

// IntID returns key fingerprint (ID) as int64.
func (a AuthKey) IntID() int64 {
	return int64(binary.LittleEndian.Uint64(a.ID[:]))
}

// SetIntID sets key fingerprint (ID) as int64.
func (a *AuthKey) SetIntID(v int64) {
	binary.LittleEndian.PutUint64(a.ID[:], uint64(v))
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
