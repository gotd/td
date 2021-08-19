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

// SecurePlainPhone represents TL type `securePlainPhone#7d6099dd`.
// Phone number to use in telegram passport¹: it must be verified, first »².
//
// Links:
//  1) https://core.telegram.org/passport
//  2) https://core.telegram.org/passport/encryption#secureplaindata
//
// See https://core.telegram.org/constructor/securePlainPhone for reference.
type SecurePlainPhone struct {
	// Phone number
	Phone string
}

// SecurePlainPhoneTypeID is TL type id of SecurePlainPhone.
const SecurePlainPhoneTypeID = 0x7d6099dd

func (s *SecurePlainPhone) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Phone == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SecurePlainPhone) String() string {
	if s == nil {
		return "SecurePlainPhone(nil)"
	}
	type Alias SecurePlainPhone
	return fmt.Sprintf("SecurePlainPhone%+v", Alias(*s))
}

// FillFrom fills SecurePlainPhone from given interface.
func (s *SecurePlainPhone) FillFrom(from interface {
	GetPhone() (value string)
}) {
	s.Phone = from.GetPhone()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SecurePlainPhone) TypeID() uint32 {
	return SecurePlainPhoneTypeID
}

// TypeName returns name of type in TL schema.
func (*SecurePlainPhone) TypeName() string {
	return "securePlainPhone"
}

// TypeInfo returns info about TL type.
func (s *SecurePlainPhone) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "securePlainPhone",
		ID:   SecurePlainPhoneTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Phone",
			SchemaName: "phone",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SecurePlainPhone) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "securePlainPhone#7d6099dd",
		}
	}
	b.PutID(SecurePlainPhoneTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SecurePlainPhone) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "securePlainPhone#7d6099dd",
		}
	}
	b.PutString(s.Phone)
	return nil
}

// GetPhone returns value of Phone field.
func (s *SecurePlainPhone) GetPhone() (value string) {
	return s.Phone
}

// Decode implements bin.Decoder.
func (s *SecurePlainPhone) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "securePlainPhone#7d6099dd",
		}
	}
	if err := b.ConsumeID(SecurePlainPhoneTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "securePlainPhone#7d6099dd",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SecurePlainPhone) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "securePlainPhone#7d6099dd",
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "securePlainPhone#7d6099dd",
				FieldName:  "phone",
				Underlying: err,
			}
		}
		s.Phone = value
	}
	return nil
}

// construct implements constructor of SecurePlainDataClass.
func (s SecurePlainPhone) construct() SecurePlainDataClass { return &s }

// Ensuring interfaces in compile-time for SecurePlainPhone.
var (
	_ bin.Encoder     = &SecurePlainPhone{}
	_ bin.Decoder     = &SecurePlainPhone{}
	_ bin.BareEncoder = &SecurePlainPhone{}
	_ bin.BareDecoder = &SecurePlainPhone{}

	_ SecurePlainDataClass = &SecurePlainPhone{}
)

// SecurePlainEmail represents TL type `securePlainEmail#21ec5a5f`.
// Email address to use in telegram passport¹: it must be verified, first »².
//
// Links:
//  1) https://core.telegram.org/passport
//  2) https://core.telegram.org/passport/encryption#secureplaindata
//
// See https://core.telegram.org/constructor/securePlainEmail for reference.
type SecurePlainEmail struct {
	// Email address
	Email string
}

// SecurePlainEmailTypeID is TL type id of SecurePlainEmail.
const SecurePlainEmailTypeID = 0x21ec5a5f

func (s *SecurePlainEmail) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Email == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SecurePlainEmail) String() string {
	if s == nil {
		return "SecurePlainEmail(nil)"
	}
	type Alias SecurePlainEmail
	return fmt.Sprintf("SecurePlainEmail%+v", Alias(*s))
}

// FillFrom fills SecurePlainEmail from given interface.
func (s *SecurePlainEmail) FillFrom(from interface {
	GetEmail() (value string)
}) {
	s.Email = from.GetEmail()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SecurePlainEmail) TypeID() uint32 {
	return SecurePlainEmailTypeID
}

// TypeName returns name of type in TL schema.
func (*SecurePlainEmail) TypeName() string {
	return "securePlainEmail"
}

// TypeInfo returns info about TL type.
func (s *SecurePlainEmail) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "securePlainEmail",
		ID:   SecurePlainEmailTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Email",
			SchemaName: "email",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SecurePlainEmail) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "securePlainEmail#21ec5a5f",
		}
	}
	b.PutID(SecurePlainEmailTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SecurePlainEmail) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "securePlainEmail#21ec5a5f",
		}
	}
	b.PutString(s.Email)
	return nil
}

// GetEmail returns value of Email field.
func (s *SecurePlainEmail) GetEmail() (value string) {
	return s.Email
}

// Decode implements bin.Decoder.
func (s *SecurePlainEmail) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "securePlainEmail#21ec5a5f",
		}
	}
	if err := b.ConsumeID(SecurePlainEmailTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "securePlainEmail#21ec5a5f",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SecurePlainEmail) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "securePlainEmail#21ec5a5f",
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "securePlainEmail#21ec5a5f",
				FieldName:  "email",
				Underlying: err,
			}
		}
		s.Email = value
	}
	return nil
}

// construct implements constructor of SecurePlainDataClass.
func (s SecurePlainEmail) construct() SecurePlainDataClass { return &s }

// Ensuring interfaces in compile-time for SecurePlainEmail.
var (
	_ bin.Encoder     = &SecurePlainEmail{}
	_ bin.Decoder     = &SecurePlainEmail{}
	_ bin.BareEncoder = &SecurePlainEmail{}
	_ bin.BareDecoder = &SecurePlainEmail{}

	_ SecurePlainDataClass = &SecurePlainEmail{}
)

// SecurePlainDataClass represents SecurePlainData generic type.
//
// See https://core.telegram.org/type/SecurePlainData for reference.
//
// Example:
//  g, err := tg.DecodeSecurePlainData(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.SecurePlainPhone: // securePlainPhone#7d6099dd
//  case *tg.SecurePlainEmail: // securePlainEmail#21ec5a5f
//  default: panic(v)
//  }
type SecurePlainDataClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() SecurePlainDataClass

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

// DecodeSecurePlainData implements binary de-serialization for SecurePlainDataClass.
func DecodeSecurePlainData(buf *bin.Buffer) (SecurePlainDataClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case SecurePlainPhoneTypeID:
		// Decoding securePlainPhone#7d6099dd.
		v := SecurePlainPhone{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "SecurePlainDataClass",
				Underlying: err,
			}
		}
		return &v, nil
	case SecurePlainEmailTypeID:
		// Decoding securePlainEmail#21ec5a5f.
		v := SecurePlainEmail{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "SecurePlainDataClass",
				Underlying: err,
			}
		}
		return &v, nil
	default:
		return nil, &bin.DecodeError{
			TypeName:   "SecurePlainDataClass",
			Underlying: bin.NewUnexpectedID(id),
		}
	}
}

// SecurePlainData boxes the SecurePlainDataClass providing a helper.
type SecurePlainDataBox struct {
	SecurePlainData SecurePlainDataClass
}

// Decode implements bin.Decoder for SecurePlainDataBox.
func (b *SecurePlainDataBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "SecurePlainDataBox",
		}
	}
	v, err := DecodeSecurePlainData(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.SecurePlainData = v
	return nil
}

// Encode implements bin.Encode for SecurePlainDataBox.
func (b *SecurePlainDataBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.SecurePlainData == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "SecurePlainDataBox",
		}
	}
	return b.SecurePlainData.Encode(buf)
}

// SecurePlainDataClassArray is adapter for slice of SecurePlainDataClass.
type SecurePlainDataClassArray []SecurePlainDataClass

// Sort sorts slice of SecurePlainDataClass.
func (s SecurePlainDataClassArray) Sort(less func(a, b SecurePlainDataClass) bool) SecurePlainDataClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of SecurePlainDataClass.
func (s SecurePlainDataClassArray) SortStable(less func(a, b SecurePlainDataClass) bool) SecurePlainDataClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of SecurePlainDataClass.
func (s SecurePlainDataClassArray) Retain(keep func(x SecurePlainDataClass) bool) SecurePlainDataClassArray {
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
func (s SecurePlainDataClassArray) First() (v SecurePlainDataClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s SecurePlainDataClassArray) Last() (v SecurePlainDataClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *SecurePlainDataClassArray) PopFirst() (v SecurePlainDataClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero SecurePlainDataClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *SecurePlainDataClassArray) Pop() (v SecurePlainDataClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsSecurePlainPhone returns copy with only SecurePlainPhone constructors.
func (s SecurePlainDataClassArray) AsSecurePlainPhone() (to SecurePlainPhoneArray) {
	for _, elem := range s {
		value, ok := elem.(*SecurePlainPhone)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsSecurePlainEmail returns copy with only SecurePlainEmail constructors.
func (s SecurePlainDataClassArray) AsSecurePlainEmail() (to SecurePlainEmailArray) {
	for _, elem := range s {
		value, ok := elem.(*SecurePlainEmail)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// SecurePlainPhoneArray is adapter for slice of SecurePlainPhone.
type SecurePlainPhoneArray []SecurePlainPhone

// Sort sorts slice of SecurePlainPhone.
func (s SecurePlainPhoneArray) Sort(less func(a, b SecurePlainPhone) bool) SecurePlainPhoneArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of SecurePlainPhone.
func (s SecurePlainPhoneArray) SortStable(less func(a, b SecurePlainPhone) bool) SecurePlainPhoneArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of SecurePlainPhone.
func (s SecurePlainPhoneArray) Retain(keep func(x SecurePlainPhone) bool) SecurePlainPhoneArray {
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
func (s SecurePlainPhoneArray) First() (v SecurePlainPhone, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s SecurePlainPhoneArray) Last() (v SecurePlainPhone, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *SecurePlainPhoneArray) PopFirst() (v SecurePlainPhone, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero SecurePlainPhone
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *SecurePlainPhoneArray) Pop() (v SecurePlainPhone, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// SecurePlainEmailArray is adapter for slice of SecurePlainEmail.
type SecurePlainEmailArray []SecurePlainEmail

// Sort sorts slice of SecurePlainEmail.
func (s SecurePlainEmailArray) Sort(less func(a, b SecurePlainEmail) bool) SecurePlainEmailArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of SecurePlainEmail.
func (s SecurePlainEmailArray) SortStable(less func(a, b SecurePlainEmail) bool) SecurePlainEmailArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of SecurePlainEmail.
func (s SecurePlainEmailArray) Retain(keep func(x SecurePlainEmail) bool) SecurePlainEmailArray {
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
func (s SecurePlainEmailArray) First() (v SecurePlainEmail, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s SecurePlainEmailArray) Last() (v SecurePlainEmail, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *SecurePlainEmailArray) PopFirst() (v SecurePlainEmail, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero SecurePlainEmail
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *SecurePlainEmailArray) Pop() (v SecurePlainEmail, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
