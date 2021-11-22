// Code generated by gotdgen, DO NOT EDIT.

package tdapi

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

// SecretChatStatePending represents TL type `secretChatStatePending#9e6c967c`.
type SecretChatStatePending struct {
}

// SecretChatStatePendingTypeID is TL type id of SecretChatStatePending.
const SecretChatStatePendingTypeID = 0x9e6c967c

// construct implements constructor of SecretChatStateClass.
func (s SecretChatStatePending) construct() SecretChatStateClass { return &s }

// Ensuring interfaces in compile-time for SecretChatStatePending.
var (
	_ bin.Encoder     = &SecretChatStatePending{}
	_ bin.Decoder     = &SecretChatStatePending{}
	_ bin.BareEncoder = &SecretChatStatePending{}
	_ bin.BareDecoder = &SecretChatStatePending{}

	_ SecretChatStateClass = &SecretChatStatePending{}
)

func (s *SecretChatStatePending) Zero() bool {
	if s == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (s *SecretChatStatePending) String() string {
	if s == nil {
		return "SecretChatStatePending(nil)"
	}
	type Alias SecretChatStatePending
	return fmt.Sprintf("SecretChatStatePending%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SecretChatStatePending) TypeID() uint32 {
	return SecretChatStatePendingTypeID
}

// TypeName returns name of type in TL schema.
func (*SecretChatStatePending) TypeName() string {
	return "secretChatStatePending"
}

// TypeInfo returns info about TL type.
func (s *SecretChatStatePending) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "secretChatStatePending",
		ID:   SecretChatStatePendingTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (s *SecretChatStatePending) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode secretChatStatePending#9e6c967c as nil")
	}
	b.PutID(SecretChatStatePendingTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SecretChatStatePending) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode secretChatStatePending#9e6c967c as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *SecretChatStatePending) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode secretChatStatePending#9e6c967c to nil")
	}
	if err := b.ConsumeID(SecretChatStatePendingTypeID); err != nil {
		return fmt.Errorf("unable to decode secretChatStatePending#9e6c967c: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SecretChatStatePending) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode secretChatStatePending#9e6c967c to nil")
	}
	return nil
}

// SecretChatStateReady represents TL type `secretChatStateReady#9ff4b7e9`.
type SecretChatStateReady struct {
}

// SecretChatStateReadyTypeID is TL type id of SecretChatStateReady.
const SecretChatStateReadyTypeID = 0x9ff4b7e9

// construct implements constructor of SecretChatStateClass.
func (s SecretChatStateReady) construct() SecretChatStateClass { return &s }

// Ensuring interfaces in compile-time for SecretChatStateReady.
var (
	_ bin.Encoder     = &SecretChatStateReady{}
	_ bin.Decoder     = &SecretChatStateReady{}
	_ bin.BareEncoder = &SecretChatStateReady{}
	_ bin.BareDecoder = &SecretChatStateReady{}

	_ SecretChatStateClass = &SecretChatStateReady{}
)

func (s *SecretChatStateReady) Zero() bool {
	if s == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (s *SecretChatStateReady) String() string {
	if s == nil {
		return "SecretChatStateReady(nil)"
	}
	type Alias SecretChatStateReady
	return fmt.Sprintf("SecretChatStateReady%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SecretChatStateReady) TypeID() uint32 {
	return SecretChatStateReadyTypeID
}

// TypeName returns name of type in TL schema.
func (*SecretChatStateReady) TypeName() string {
	return "secretChatStateReady"
}

// TypeInfo returns info about TL type.
func (s *SecretChatStateReady) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "secretChatStateReady",
		ID:   SecretChatStateReadyTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (s *SecretChatStateReady) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode secretChatStateReady#9ff4b7e9 as nil")
	}
	b.PutID(SecretChatStateReadyTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SecretChatStateReady) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode secretChatStateReady#9ff4b7e9 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *SecretChatStateReady) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode secretChatStateReady#9ff4b7e9 to nil")
	}
	if err := b.ConsumeID(SecretChatStateReadyTypeID); err != nil {
		return fmt.Errorf("unable to decode secretChatStateReady#9ff4b7e9: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SecretChatStateReady) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode secretChatStateReady#9ff4b7e9 to nil")
	}
	return nil
}

// SecretChatStateClosed represents TL type `secretChatStateClosed#8c1006ed`.
type SecretChatStateClosed struct {
}

// SecretChatStateClosedTypeID is TL type id of SecretChatStateClosed.
const SecretChatStateClosedTypeID = 0x8c1006ed

// construct implements constructor of SecretChatStateClass.
func (s SecretChatStateClosed) construct() SecretChatStateClass { return &s }

// Ensuring interfaces in compile-time for SecretChatStateClosed.
var (
	_ bin.Encoder     = &SecretChatStateClosed{}
	_ bin.Decoder     = &SecretChatStateClosed{}
	_ bin.BareEncoder = &SecretChatStateClosed{}
	_ bin.BareDecoder = &SecretChatStateClosed{}

	_ SecretChatStateClass = &SecretChatStateClosed{}
)

func (s *SecretChatStateClosed) Zero() bool {
	if s == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (s *SecretChatStateClosed) String() string {
	if s == nil {
		return "SecretChatStateClosed(nil)"
	}
	type Alias SecretChatStateClosed
	return fmt.Sprintf("SecretChatStateClosed%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SecretChatStateClosed) TypeID() uint32 {
	return SecretChatStateClosedTypeID
}

// TypeName returns name of type in TL schema.
func (*SecretChatStateClosed) TypeName() string {
	return "secretChatStateClosed"
}

// TypeInfo returns info about TL type.
func (s *SecretChatStateClosed) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "secretChatStateClosed",
		ID:   SecretChatStateClosedTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (s *SecretChatStateClosed) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode secretChatStateClosed#8c1006ed as nil")
	}
	b.PutID(SecretChatStateClosedTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SecretChatStateClosed) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode secretChatStateClosed#8c1006ed as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *SecretChatStateClosed) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode secretChatStateClosed#8c1006ed to nil")
	}
	if err := b.ConsumeID(SecretChatStateClosedTypeID); err != nil {
		return fmt.Errorf("unable to decode secretChatStateClosed#8c1006ed: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SecretChatStateClosed) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode secretChatStateClosed#8c1006ed to nil")
	}
	return nil
}

// SecretChatStateClass represents SecretChatState generic type.
//
// Example:
//  g, err := tdapi.DecodeSecretChatState(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tdapi.SecretChatStatePending: // secretChatStatePending#9e6c967c
//  case *tdapi.SecretChatStateReady: // secretChatStateReady#9ff4b7e9
//  case *tdapi.SecretChatStateClosed: // secretChatStateClosed#8c1006ed
//  default: panic(v)
//  }
type SecretChatStateClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() SecretChatStateClass

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

// DecodeSecretChatState implements binary de-serialization for SecretChatStateClass.
func DecodeSecretChatState(buf *bin.Buffer) (SecretChatStateClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case SecretChatStatePendingTypeID:
		// Decoding secretChatStatePending#9e6c967c.
		v := SecretChatStatePending{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode SecretChatStateClass: %w", err)
		}
		return &v, nil
	case SecretChatStateReadyTypeID:
		// Decoding secretChatStateReady#9ff4b7e9.
		v := SecretChatStateReady{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode SecretChatStateClass: %w", err)
		}
		return &v, nil
	case SecretChatStateClosedTypeID:
		// Decoding secretChatStateClosed#8c1006ed.
		v := SecretChatStateClosed{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode SecretChatStateClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode SecretChatStateClass: %w", bin.NewUnexpectedID(id))
	}
}

// SecretChatState boxes the SecretChatStateClass providing a helper.
type SecretChatStateBox struct {
	SecretChatState SecretChatStateClass
}

// Decode implements bin.Decoder for SecretChatStateBox.
func (b *SecretChatStateBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode SecretChatStateBox to nil")
	}
	v, err := DecodeSecretChatState(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.SecretChatState = v
	return nil
}

// Encode implements bin.Encode for SecretChatStateBox.
func (b *SecretChatStateBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.SecretChatState == nil {
		return fmt.Errorf("unable to encode SecretChatStateClass as nil")
	}
	return b.SecretChatState.Encode(buf)
}