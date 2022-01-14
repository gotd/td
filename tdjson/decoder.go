package tdjson

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/gotd/td/bin"
)

// Decoder is a simple wrapper around jx.Decoder to conform TL type system.
type Decoder struct {
	*jx.Decoder
}

// Obj calls f for every key in object, using byte slice as key.
//
// The key value is valid only until f is not returned.
func (b Decoder) Obj(cb func(d Decoder, key []byte) error) error {
	return b.Decoder.ObjBytes(func(d *jx.Decoder, key []byte) error {
		return cb(Decoder{Decoder: d}, key)
	})
}

// Arr decodes array and invokes callback on each array element.
func (b Decoder) Arr(cb func(d Decoder) error) error {
	return b.Decoder.Arr(func(d *jx.Decoder) error {
		return cb(Decoder{Decoder: d})
	})
}

// ID deserializes given typeID.
func (b Decoder) ID() (string, error) {
	return b.Decoder.Str()
}

// FindTypeID tries to find @type field or returns error.
func (b Decoder) FindTypeID() (string, error) {
	var (
		found bool
		typ   string
	)
	if err := b.Decoder.Capture(func(d *jx.Decoder) error {
		return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
			if found || string(key) != TypeField {
				return d.Skip()
			}

			t, err := d.Str()
			if err != nil {
				return err
			}
			typ = t
			found = true
			return nil
		})
	}); err != nil {
		return "", err
	}
	if !found {
		return "", ErrTypeIDNotFound
	}
	return typ, nil
}

// ConsumeID deserializes given typeID.
func (b Decoder) ConsumeID(id string) error {
	v, err := b.Decoder.Str()
	if err != nil {
		return err
	}
	if v != id {
		return NewUnexpectedID(id)
	}
	return nil
}

// Int deserializes signed 32-bit integer.
func (b Decoder) Int() (int, error) {
	return b.Decoder.Int()
}

// Bool deserializes boolean.
func (b Decoder) Bool() (bool, error) {
	return b.Decoder.Bool()
}

// Uint16 deserializes unsigned 16-bit integer.
func (b Decoder) Uint16() (uint16, error) {
	v, err := b.Decoder.UInt32()
	if err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// Int32 deserializes signed 32-bit integer.
func (b Decoder) Int32() (int32, error) {
	return b.Decoder.Int32()
}

// Uint32 deserializes unsigned 32-bit integer.
func (b Decoder) Uint32() (uint32, error) {
	return b.Decoder.UInt32()
}

// Int53 deserializes int53.
func (b Decoder) Int53() (int64, error) {
	return b.Decoder.Int64()
}

// Long deserializes int64.
func (b Decoder) Long() (int64, error) {
	n, err := b.Decoder.Num()
	if err != nil {
		return 0, err
	}
	return n.Int64()
}

// Uint64 deserializes unsigned 64-bit integer.
func (b Decoder) Uint64() (uint64, error) {
	return b.Decoder.UInt64()
}

// Double deserializes 64-bit floating point.
func (b Decoder) Double() (float64, error) {
	return b.Decoder.Float64()
}

// Int128 deserializes 128-bit signed integer.
func (b Decoder) Int128() (bin.Int128, error) {
	// FIXME(tdakkota): neither TDLib API not Telegram API has no Int128/Int256 fields
	// 	so this encoding may incorrect.
	v, err := b.Decoder.Str()
	if err != nil {
		return bin.Int128{}, err
	}

	var result bin.Int128
	if l := hex.DecodedLen(len(v)); l != len(result) {
		return bin.Int128{}, errors.Wrapf(err, "invalid length %d", l)
	}

	if _, err := hex.Decode(result[:], []byte(v)); err != nil {
		return bin.Int128{}, err
	}

	return result, nil
}

// Int256 deserializes 256-bit signed integer.
func (b Decoder) Int256() (bin.Int256, error) {
	// FIXME(tdakkota): neither TDLib API not Telegram API has no Int128/Int256 fields
	// 	so this encoding may incorrect.
	v, err := b.Decoder.StrBytes()
	if err != nil {
		return bin.Int256{}, err
	}

	var result bin.Int256
	if l := hex.DecodedLen(len(v)); l != len(result) {
		return bin.Int256{}, errors.Wrapf(err, "invalid length %d", l)
	}

	if _, err := hex.Decode(result[:], v); err != nil {
		return bin.Int256{}, err
	}

	return result, nil
}

// String deserializes bare string.
func (b Decoder) String() (string, error) {
	return b.Decoder.Str()
}

// Bytes deserializes bare byte string.
func (b Decoder) Bytes() ([]byte, error) {
	// See https://core.telegram.org/tdlib/docs/td__json__client_8h.html
	//
	// ... fields of bytes type are base64 encoded and then stored as String ...
	enc := base64.RawStdEncoding

	v, err := b.Decoder.StrBytes()
	if err != nil {
		return nil, err
	}

	result := make([]byte, enc.DecodedLen(len(v)))
	if _, err := enc.Decode(result, v); err != nil {
		return nil, err
	}

	return result, nil
}
