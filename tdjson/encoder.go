package tdjson

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"

	"github.com/go-faster/jx"

	"github.com/gotd/td/bin"
)

// Encoder is a simple wrapper around jx.Encoder to conform TL type system.
type Encoder struct {
	*jx.Encoder
}

// PutID serializes given typeID.
func (b Encoder) PutID(typeID string) {
	b.Encoder.FieldStart(TypeField)
	b.Encoder.Str(typeID)
}

// PutInt serializes v as signed 32-bit integer.
func (b Encoder) PutInt(v int) {
	b.Encoder.Int(v)
}

// PutBool serializes boolean.
func (b Encoder) PutBool(v bool) {
	b.Encoder.Bool(v)
}

// PutUint16 serializes unsigned 16-bit integer.
func (b Encoder) PutUint16(v uint16) {
	b.Encoder.UInt32(uint32(v))
}

// PutInt32 serializes signed 32-bit integer.
func (b Encoder) PutInt32(v int32) {
	b.Encoder.Int32(v)
}

// PutUint32 serializes unsigned 32-bit integer.
func (b Encoder) PutUint32(v uint32) {
	b.Encoder.UInt32(v)
}

// PutInt53 serializes v as int53.
func (b Encoder) PutInt53(v int64) {
	b.Encoder.Int64(v)
}

// PutLong serializes v as int64.
func (b Encoder) PutLong(v int64) {
	var buf [32]byte
	r := append(buf[:0], '"')
	r = strconv.AppendInt(r, v, 10)
	r = append(r, '"')
	b.Encoder.Raw(r)
}

// PutUint64 serializes v as unsigned 64-bit integer.
func (b Encoder) PutUint64(v uint64) {
	// FIXME(tdakkota): TDLib API has no uint64 fields
	// 	so this encoding may incorrect.
	b.Encoder.UInt64(v)
}

// PutDouble serializes v as 64-bit floating point.
func (b Encoder) PutDouble(v float64) {
	b.Encoder.Float64(v)
}

// PutInt128 serializes v as 128-bit signed integer.
func (b Encoder) PutInt128(v bin.Int128) {
	// FIXME(tdakkota): neither TDLib API nor Telegram API has no Int128/Int256 fields
	// 	so this encoding may incorrect.
	b.Encoder.Str(hex.EncodeToString(v[:]))
}

// PutInt256 serializes v as 256-bit signed integer.
func (b Encoder) PutInt256(v bin.Int256) {
	// FIXME(tdakkota): neither TDLib API not Telegram API has no Int128/Int256 fields
	// 	so this encoding may incorrect.
	b.Encoder.Str(hex.EncodeToString(v[:]))
}

// PutString serializes bare string.
func (b Encoder) PutString(s string) {
	b.Encoder.Str(s)
}

// PutBytes serializes bare byte string.
func (b Encoder) PutBytes(v []byte) {
	// See https://core.telegram.org/tdlib/docs/td__json__client_8h.html
	//
	// ... fields of bytes type are base64 encoded and then stored as String ...
	b.Encoder.Str(base64.RawURLEncoding.EncodeToString(v))
}
