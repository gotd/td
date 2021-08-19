// Code generated by gotdgen, DO NOT EDIT.

package tg

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/multierr"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdp"
	"github.com/gotd/td/tgerr"
)

// No-op definition for keeping imports.
var (
	_ = bin.Buffer{}
	_ = context.Background()
	_ = fmt.Stringer(nil)
	_ = strings.Builder{}
	_ = errors.Is
	_ = multierr.AppendInto
	_ = sort.Ints
	_ = tdp.Format
	_ = tgerr.Error{}
)

// PasswordKdfAlgoUnknown represents TL type `passwordKdfAlgoUnknown#d45ab096`.
// Unknown KDF (most likely, the client is outdated and does not support the specified
// KDF algorithm)
//
// See https://core.telegram.org/constructor/passwordKdfAlgoUnknown for reference.
type PasswordKdfAlgoUnknown struct {
}

// PasswordKdfAlgoUnknownTypeID is TL type id of PasswordKdfAlgoUnknown.
const PasswordKdfAlgoUnknownTypeID = 0xd45ab096

func (p *PasswordKdfAlgoUnknown) Zero() bool {
	if p == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (p *PasswordKdfAlgoUnknown) String() string {
	if p == nil {
		return "PasswordKdfAlgoUnknown(nil)"
	}
	type Alias PasswordKdfAlgoUnknown
	return fmt.Sprintf("PasswordKdfAlgoUnknown%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PasswordKdfAlgoUnknown) TypeID() uint32 {
	return PasswordKdfAlgoUnknownTypeID
}

// TypeName returns name of type in TL schema.
func (*PasswordKdfAlgoUnknown) TypeName() string {
	return "passwordKdfAlgoUnknown"
}

// TypeInfo returns info about TL type.
func (p *PasswordKdfAlgoUnknown) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "passwordKdfAlgoUnknown",
		ID:   PasswordKdfAlgoUnknownTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (p *PasswordKdfAlgoUnknown) Encode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "passwordKdfAlgoUnknown#d45ab096",
		}
	}
	b.PutID(PasswordKdfAlgoUnknownTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PasswordKdfAlgoUnknown) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "passwordKdfAlgoUnknown#d45ab096",
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (p *PasswordKdfAlgoUnknown) Decode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "passwordKdfAlgoUnknown#d45ab096",
		}
	}
	if err := b.ConsumeID(PasswordKdfAlgoUnknownTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "passwordKdfAlgoUnknown#d45ab096",
			Underlying: err,
		}
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PasswordKdfAlgoUnknown) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "passwordKdfAlgoUnknown#d45ab096",
		}
	}
	return nil
}

// construct implements constructor of PasswordKdfAlgoClass.
func (p PasswordKdfAlgoUnknown) construct() PasswordKdfAlgoClass { return &p }

// Ensuring interfaces in compile-time for PasswordKdfAlgoUnknown.
var (
	_ bin.Encoder     = &PasswordKdfAlgoUnknown{}
	_ bin.Decoder     = &PasswordKdfAlgoUnknown{}
	_ bin.BareEncoder = &PasswordKdfAlgoUnknown{}
	_ bin.BareDecoder = &PasswordKdfAlgoUnknown{}

	_ PasswordKdfAlgoClass = &PasswordKdfAlgoUnknown{}
)

// PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow represents TL type `passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a`.
// This key derivation algorithm defines that SRP 2FA login¹ must be used
//
// Links:
//  1) https://core.telegram.org/api/srp
//
// See https://core.telegram.org/constructor/passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow for reference.
type PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow struct {
	// One of two salts used by the derivation function (see SRP 2FA login¹)
	//
	// Links:
	//  1) https://core.telegram.org/api/srp
	Salt1 []byte
	// One of two salts used by the derivation function (see SRP 2FA login¹)
	//
	// Links:
	//  1) https://core.telegram.org/api/srp
	Salt2 []byte
	// Base (see SRP 2FA login¹)
	//
	// Links:
	//  1) https://core.telegram.org/api/srp
	G int
	// 2048-bit modulus (see SRP 2FA login¹)
	//
	// Links:
	//  1) https://core.telegram.org/api/srp
	P []byte
}

// PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowTypeID is TL type id of PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow.
const PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowTypeID = 0x3a912d4a

func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.Salt1 == nil) {
		return false
	}
	if !(p.Salt2 == nil) {
		return false
	}
	if !(p.G == 0) {
		return false
	}
	if !(p.P == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) String() string {
	if p == nil {
		return "PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow(nil)"
	}
	type Alias PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow
	return fmt.Sprintf("PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow%+v", Alias(*p))
}

// FillFrom fills PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow from given interface.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) FillFrom(from interface {
	GetSalt1() (value []byte)
	GetSalt2() (value []byte)
	GetG() (value int)
	GetP() (value []byte)
}) {
	p.Salt1 = from.GetSalt1()
	p.Salt2 = from.GetSalt2()
	p.G = from.GetG()
	p.P = from.GetP()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) TypeID() uint32 {
	return PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowTypeID
}

// TypeName returns name of type in TL schema.
func (*PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) TypeName() string {
	return "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow"
}

// TypeInfo returns info about TL type.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow",
		ID:   PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Salt1",
			SchemaName: "salt1",
		},
		{
			Name:       "Salt2",
			SchemaName: "salt2",
		},
		{
			Name:       "G",
			SchemaName: "g",
		},
		{
			Name:       "P",
			SchemaName: "p",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) Encode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
		}
	}
	b.PutID(PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
		}
	}
	b.PutBytes(p.Salt1)
	b.PutBytes(p.Salt2)
	b.PutInt(p.G)
	b.PutBytes(p.P)
	return nil
}

// GetSalt1 returns value of Salt1 field.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) GetSalt1() (value []byte) {
	return p.Salt1
}

// GetSalt2 returns value of Salt2 field.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) GetSalt2() (value []byte) {
	return p.Salt2
}

// GetG returns value of G field.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) GetG() (value int) {
	return p.G
}

// GetP returns value of P field.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) GetP() (value []byte) {
	return p.P
}

// Decode implements bin.Decoder.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) Decode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
		}
	}
	if err := b.ConsumeID(PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
			Underlying: err,
		}
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
		}
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
				FieldName:  "salt1",
				Underlying: err,
			}
		}
		p.Salt1 = value
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
				FieldName:  "salt2",
				Underlying: err,
			}
		}
		p.Salt2 = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
				FieldName:  "g",
				Underlying: err,
			}
		}
		p.G = value
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a",
				FieldName:  "p",
				Underlying: err,
			}
		}
		p.P = value
	}
	return nil
}

// construct implements constructor of PasswordKdfAlgoClass.
func (p PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) construct() PasswordKdfAlgoClass {
	return &p
}

// Ensuring interfaces in compile-time for PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow.
var (
	_ bin.Encoder     = &PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{}
	_ bin.Decoder     = &PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{}
	_ bin.BareEncoder = &PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{}
	_ bin.BareDecoder = &PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{}

	_ PasswordKdfAlgoClass = &PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{}
)

// PasswordKdfAlgoClass represents PasswordKdfAlgo generic type.
//
// See https://core.telegram.org/type/PasswordKdfAlgo for reference.
//
// Example:
//  g, err := tg.DecodePasswordKdfAlgo(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.PasswordKdfAlgoUnknown: // passwordKdfAlgoUnknown#d45ab096
//  case *tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow: // passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a
//  default: panic(v)
//  }
type PasswordKdfAlgoClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() PasswordKdfAlgoClass

	// TypeID returns type id in TL schema.
	//
	// See https://core.telegram.org/mtproto/TL-tl#remarks.
	TypeID() uint32
	// TypeName returns name of type in TL schema.
	TypeName() string
	// String implements fmt.Stringer.
	String() string
	// Zero returns true if current object has a zero value.
	Zero() bool
}

// DecodePasswordKdfAlgo implements binary de-serialization for PasswordKdfAlgoClass.
func DecodePasswordKdfAlgo(buf *bin.Buffer) (PasswordKdfAlgoClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case PasswordKdfAlgoUnknownTypeID:
		// Decoding passwordKdfAlgoUnknown#d45ab096.
		v := PasswordKdfAlgoUnknown{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "PasswordKdfAlgoClass",
				Underlying: err,
			}
		}
		return &v, nil
	case PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowTypeID:
		// Decoding passwordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow#3a912d4a.
		v := PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "PasswordKdfAlgoClass",
				Underlying: err,
			}
		}
		return &v, nil
	default:
		return nil, &bin.DecodeError{
			TypeName:   "PasswordKdfAlgoClass",
			Underlying: bin.NewUnexpectedID(id),
		}
	}
}

// PasswordKdfAlgo boxes the PasswordKdfAlgoClass providing a helper.
type PasswordKdfAlgoBox struct {
	PasswordKdfAlgo PasswordKdfAlgoClass
}

// Decode implements bin.Decoder for PasswordKdfAlgoBox.
func (b *PasswordKdfAlgoBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "PasswordKdfAlgoBox",
		}
	}
	v, err := DecodePasswordKdfAlgo(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.PasswordKdfAlgo = v
	return nil
}

// Encode implements bin.Encode for PasswordKdfAlgoBox.
func (b *PasswordKdfAlgoBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.PasswordKdfAlgo == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "PasswordKdfAlgoBox",
		}
	}
	return b.PasswordKdfAlgo.Encode(buf)
}

// PasswordKdfAlgoClassArray is adapter for slice of PasswordKdfAlgoClass.
type PasswordKdfAlgoClassArray []PasswordKdfAlgoClass

// Sort sorts slice of PasswordKdfAlgoClass.
func (s PasswordKdfAlgoClassArray) Sort(less func(a, b PasswordKdfAlgoClass) bool) PasswordKdfAlgoClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of PasswordKdfAlgoClass.
func (s PasswordKdfAlgoClassArray) SortStable(less func(a, b PasswordKdfAlgoClass) bool) PasswordKdfAlgoClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of PasswordKdfAlgoClass.
func (s PasswordKdfAlgoClassArray) Retain(keep func(x PasswordKdfAlgoClass) bool) PasswordKdfAlgoClassArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s PasswordKdfAlgoClassArray) First() (v PasswordKdfAlgoClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s PasswordKdfAlgoClassArray) Last() (v PasswordKdfAlgoClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *PasswordKdfAlgoClassArray) PopFirst() (v PasswordKdfAlgoClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero PasswordKdfAlgoClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *PasswordKdfAlgoClassArray) Pop() (v PasswordKdfAlgoClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsPasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow returns copy with only PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow constructors.
func (s PasswordKdfAlgoClassArray) AsPasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow() (to PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) {
	for _, elem := range s {
		value, ok := elem.(*PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray is adapter for slice of PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow.
type PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray []PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow

// Sort sorts slice of PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow.
func (s PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) Sort(less func(a, b PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) bool) PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow.
func (s PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) SortStable(less func(a, b PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) bool) PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow.
func (s PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) Retain(keep func(x PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow) bool) PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) First() (v PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) Last() (v PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) PopFirst() (v PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPowArray) Pop() (v PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
