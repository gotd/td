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

// MessagesExportedChatInvite represents TL type `messages.exportedChatInvite#1871be50`.
//
// See https://core.telegram.org/constructor/messages.exportedChatInvite for reference.
type MessagesExportedChatInvite struct {
	// Invite field of MessagesExportedChatInvite.
	Invite ChatInviteExported
	// Users field of MessagesExportedChatInvite.
	Users []UserClass
}

// MessagesExportedChatInviteTypeID is TL type id of MessagesExportedChatInvite.
const MessagesExportedChatInviteTypeID = 0x1871be50

func (e *MessagesExportedChatInvite) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Invite.Zero()) {
		return false
	}
	if !(e.Users == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *MessagesExportedChatInvite) String() string {
	if e == nil {
		return "MessagesExportedChatInvite(nil)"
	}
	type Alias MessagesExportedChatInvite
	return fmt.Sprintf("MessagesExportedChatInvite%+v", Alias(*e))
}

// FillFrom fills MessagesExportedChatInvite from given interface.
func (e *MessagesExportedChatInvite) FillFrom(from interface {
	GetInvite() (value ChatInviteExported)
	GetUsers() (value []UserClass)
}) {
	e.Invite = from.GetInvite()
	e.Users = from.GetUsers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesExportedChatInvite) TypeID() uint32 {
	return MessagesExportedChatInviteTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesExportedChatInvite) TypeName() string {
	return "messages.exportedChatInvite"
}

// TypeInfo returns info about TL type.
func (e *MessagesExportedChatInvite) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.exportedChatInvite",
		ID:   MessagesExportedChatInviteTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Invite",
			SchemaName: "invite",
		},
		{
			Name:       "Users",
			SchemaName: "users",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (e *MessagesExportedChatInvite) Encode(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.exportedChatInvite#1871be50",
		}
	}
	b.PutID(MessagesExportedChatInviteTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *MessagesExportedChatInvite) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.exportedChatInvite#1871be50",
		}
	}
	if err := e.Invite.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.exportedChatInvite#1871be50",
			FieldName:  "invite",
			Underlying: err,
		}
	}
	b.PutVectorHeader(len(e.Users))
	for idx, v := range e.Users {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.exportedChatInvite#1871be50",
				FieldName: "users",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<User>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.exportedChatInvite#1871be50",
				FieldName: "users",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	return nil
}

// GetInvite returns value of Invite field.
func (e *MessagesExportedChatInvite) GetInvite() (value ChatInviteExported) {
	return e.Invite
}

// GetUsers returns value of Users field.
func (e *MessagesExportedChatInvite) GetUsers() (value []UserClass) {
	return e.Users
}

// MapUsers returns field Users wrapped in UserClassArray helper.
func (e *MessagesExportedChatInvite) MapUsers() (value UserClassArray) {
	return UserClassArray(e.Users)
}

// Decode implements bin.Decoder.
func (e *MessagesExportedChatInvite) Decode(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.exportedChatInvite#1871be50",
		}
	}
	if err := b.ConsumeID(MessagesExportedChatInviteTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.exportedChatInvite#1871be50",
			Underlying: err,
		}
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *MessagesExportedChatInvite) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.exportedChatInvite#1871be50",
		}
	}
	{
		if err := e.Invite.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.exportedChatInvite#1871be50",
				FieldName:  "invite",
				Underlying: err,
			}
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.exportedChatInvite#1871be50",
				FieldName:  "users",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			e.Users = make([]UserClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeUser(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "messages.exportedChatInvite#1871be50",
					FieldName:  "users",
					Underlying: err,
				}
			}
			e.Users = append(e.Users, value)
		}
	}
	return nil
}

// construct implements constructor of MessagesExportedChatInviteClass.
func (e MessagesExportedChatInvite) construct() MessagesExportedChatInviteClass { return &e }

// Ensuring interfaces in compile-time for MessagesExportedChatInvite.
var (
	_ bin.Encoder     = &MessagesExportedChatInvite{}
	_ bin.Decoder     = &MessagesExportedChatInvite{}
	_ bin.BareEncoder = &MessagesExportedChatInvite{}
	_ bin.BareDecoder = &MessagesExportedChatInvite{}

	_ MessagesExportedChatInviteClass = &MessagesExportedChatInvite{}
)

// MessagesExportedChatInviteReplaced represents TL type `messages.exportedChatInviteReplaced#222600ef`.
//
// See https://core.telegram.org/constructor/messages.exportedChatInviteReplaced for reference.
type MessagesExportedChatInviteReplaced struct {
	// Invite field of MessagesExportedChatInviteReplaced.
	Invite ChatInviteExported
	// NewInvite field of MessagesExportedChatInviteReplaced.
	NewInvite ChatInviteExported
	// Users field of MessagesExportedChatInviteReplaced.
	Users []UserClass
}

// MessagesExportedChatInviteReplacedTypeID is TL type id of MessagesExportedChatInviteReplaced.
const MessagesExportedChatInviteReplacedTypeID = 0x222600ef

func (e *MessagesExportedChatInviteReplaced) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Invite.Zero()) {
		return false
	}
	if !(e.NewInvite.Zero()) {
		return false
	}
	if !(e.Users == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *MessagesExportedChatInviteReplaced) String() string {
	if e == nil {
		return "MessagesExportedChatInviteReplaced(nil)"
	}
	type Alias MessagesExportedChatInviteReplaced
	return fmt.Sprintf("MessagesExportedChatInviteReplaced%+v", Alias(*e))
}

// FillFrom fills MessagesExportedChatInviteReplaced from given interface.
func (e *MessagesExportedChatInviteReplaced) FillFrom(from interface {
	GetInvite() (value ChatInviteExported)
	GetNewInvite() (value ChatInviteExported)
	GetUsers() (value []UserClass)
}) {
	e.Invite = from.GetInvite()
	e.NewInvite = from.GetNewInvite()
	e.Users = from.GetUsers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesExportedChatInviteReplaced) TypeID() uint32 {
	return MessagesExportedChatInviteReplacedTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesExportedChatInviteReplaced) TypeName() string {
	return "messages.exportedChatInviteReplaced"
}

// TypeInfo returns info about TL type.
func (e *MessagesExportedChatInviteReplaced) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.exportedChatInviteReplaced",
		ID:   MessagesExportedChatInviteReplacedTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Invite",
			SchemaName: "invite",
		},
		{
			Name:       "NewInvite",
			SchemaName: "new_invite",
		},
		{
			Name:       "Users",
			SchemaName: "users",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (e *MessagesExportedChatInviteReplaced) Encode(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.exportedChatInviteReplaced#222600ef",
		}
	}
	b.PutID(MessagesExportedChatInviteReplacedTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *MessagesExportedChatInviteReplaced) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.exportedChatInviteReplaced#222600ef",
		}
	}
	if err := e.Invite.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.exportedChatInviteReplaced#222600ef",
			FieldName:  "invite",
			Underlying: err,
		}
	}
	if err := e.NewInvite.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.exportedChatInviteReplaced#222600ef",
			FieldName:  "new_invite",
			Underlying: err,
		}
	}
	b.PutVectorHeader(len(e.Users))
	for idx, v := range e.Users {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.exportedChatInviteReplaced#222600ef",
				FieldName: "users",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<User>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.exportedChatInviteReplaced#222600ef",
				FieldName: "users",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	return nil
}

// GetInvite returns value of Invite field.
func (e *MessagesExportedChatInviteReplaced) GetInvite() (value ChatInviteExported) {
	return e.Invite
}

// GetNewInvite returns value of NewInvite field.
func (e *MessagesExportedChatInviteReplaced) GetNewInvite() (value ChatInviteExported) {
	return e.NewInvite
}

// GetUsers returns value of Users field.
func (e *MessagesExportedChatInviteReplaced) GetUsers() (value []UserClass) {
	return e.Users
}

// MapUsers returns field Users wrapped in UserClassArray helper.
func (e *MessagesExportedChatInviteReplaced) MapUsers() (value UserClassArray) {
	return UserClassArray(e.Users)
}

// Decode implements bin.Decoder.
func (e *MessagesExportedChatInviteReplaced) Decode(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.exportedChatInviteReplaced#222600ef",
		}
	}
	if err := b.ConsumeID(MessagesExportedChatInviteReplacedTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.exportedChatInviteReplaced#222600ef",
			Underlying: err,
		}
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *MessagesExportedChatInviteReplaced) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.exportedChatInviteReplaced#222600ef",
		}
	}
	{
		if err := e.Invite.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.exportedChatInviteReplaced#222600ef",
				FieldName:  "invite",
				Underlying: err,
			}
		}
	}
	{
		if err := e.NewInvite.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.exportedChatInviteReplaced#222600ef",
				FieldName:  "new_invite",
				Underlying: err,
			}
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.exportedChatInviteReplaced#222600ef",
				FieldName:  "users",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			e.Users = make([]UserClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeUser(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "messages.exportedChatInviteReplaced#222600ef",
					FieldName:  "users",
					Underlying: err,
				}
			}
			e.Users = append(e.Users, value)
		}
	}
	return nil
}

// construct implements constructor of MessagesExportedChatInviteClass.
func (e MessagesExportedChatInviteReplaced) construct() MessagesExportedChatInviteClass { return &e }

// Ensuring interfaces in compile-time for MessagesExportedChatInviteReplaced.
var (
	_ bin.Encoder     = &MessagesExportedChatInviteReplaced{}
	_ bin.Decoder     = &MessagesExportedChatInviteReplaced{}
	_ bin.BareEncoder = &MessagesExportedChatInviteReplaced{}
	_ bin.BareDecoder = &MessagesExportedChatInviteReplaced{}

	_ MessagesExportedChatInviteClass = &MessagesExportedChatInviteReplaced{}
)

// MessagesExportedChatInviteClass represents messages.ExportedChatInvite generic type.
//
// See https://core.telegram.org/type/messages.ExportedChatInvite for reference.
//
// Example:
//  g, err := tg.DecodeMessagesExportedChatInvite(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.MessagesExportedChatInvite: // messages.exportedChatInvite#1871be50
//  case *tg.MessagesExportedChatInviteReplaced: // messages.exportedChatInviteReplaced#222600ef
//  default: panic(v)
//  }
type MessagesExportedChatInviteClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() MessagesExportedChatInviteClass

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

	// Invite field of MessagesExportedChatInvite.
	GetInvite() (value ChatInviteExported)

	// Users field of MessagesExportedChatInvite.
	GetUsers() (value []UserClass)
	// Users field of MessagesExportedChatInvite.
	MapUsers() (value UserClassArray)
}

// DecodeMessagesExportedChatInvite implements binary de-serialization for MessagesExportedChatInviteClass.
func DecodeMessagesExportedChatInvite(buf *bin.Buffer) (MessagesExportedChatInviteClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case MessagesExportedChatInviteTypeID:
		// Decoding messages.exportedChatInvite#1871be50.
		v := MessagesExportedChatInvite{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "MessagesExportedChatInviteClass",
				Underlying: err,
			}
		}
		return &v, nil
	case MessagesExportedChatInviteReplacedTypeID:
		// Decoding messages.exportedChatInviteReplaced#222600ef.
		v := MessagesExportedChatInviteReplaced{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "MessagesExportedChatInviteClass",
				Underlying: err,
			}
		}
		return &v, nil
	default:
		return nil, &bin.DecodeError{
			TypeName:   "MessagesExportedChatInviteClass",
			Underlying: bin.NewUnexpectedID(id),
		}
	}
}

// MessagesExportedChatInvite boxes the MessagesExportedChatInviteClass providing a helper.
type MessagesExportedChatInviteBox struct {
	ExportedChatInvite MessagesExportedChatInviteClass
}

// Decode implements bin.Decoder for MessagesExportedChatInviteBox.
func (b *MessagesExportedChatInviteBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "MessagesExportedChatInviteBox",
		}
	}
	v, err := DecodeMessagesExportedChatInvite(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.ExportedChatInvite = v
	return nil
}

// Encode implements bin.Encode for MessagesExportedChatInviteBox.
func (b *MessagesExportedChatInviteBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.ExportedChatInvite == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "MessagesExportedChatInviteBox",
		}
	}
	return b.ExportedChatInvite.Encode(buf)
}

// MessagesExportedChatInviteClassArray is adapter for slice of MessagesExportedChatInviteClass.
type MessagesExportedChatInviteClassArray []MessagesExportedChatInviteClass

// Sort sorts slice of MessagesExportedChatInviteClass.
func (s MessagesExportedChatInviteClassArray) Sort(less func(a, b MessagesExportedChatInviteClass) bool) MessagesExportedChatInviteClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of MessagesExportedChatInviteClass.
func (s MessagesExportedChatInviteClassArray) SortStable(less func(a, b MessagesExportedChatInviteClass) bool) MessagesExportedChatInviteClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of MessagesExportedChatInviteClass.
func (s MessagesExportedChatInviteClassArray) Retain(keep func(x MessagesExportedChatInviteClass) bool) MessagesExportedChatInviteClassArray {
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
func (s MessagesExportedChatInviteClassArray) First() (v MessagesExportedChatInviteClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s MessagesExportedChatInviteClassArray) Last() (v MessagesExportedChatInviteClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *MessagesExportedChatInviteClassArray) PopFirst() (v MessagesExportedChatInviteClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero MessagesExportedChatInviteClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *MessagesExportedChatInviteClassArray) Pop() (v MessagesExportedChatInviteClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsMessagesExportedChatInvite returns copy with only MessagesExportedChatInvite constructors.
func (s MessagesExportedChatInviteClassArray) AsMessagesExportedChatInvite() (to MessagesExportedChatInviteArray) {
	for _, elem := range s {
		value, ok := elem.(*MessagesExportedChatInvite)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsMessagesExportedChatInviteReplaced returns copy with only MessagesExportedChatInviteReplaced constructors.
func (s MessagesExportedChatInviteClassArray) AsMessagesExportedChatInviteReplaced() (to MessagesExportedChatInviteReplacedArray) {
	for _, elem := range s {
		value, ok := elem.(*MessagesExportedChatInviteReplaced)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// MessagesExportedChatInviteArray is adapter for slice of MessagesExportedChatInvite.
type MessagesExportedChatInviteArray []MessagesExportedChatInvite

// Sort sorts slice of MessagesExportedChatInvite.
func (s MessagesExportedChatInviteArray) Sort(less func(a, b MessagesExportedChatInvite) bool) MessagesExportedChatInviteArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of MessagesExportedChatInvite.
func (s MessagesExportedChatInviteArray) SortStable(less func(a, b MessagesExportedChatInvite) bool) MessagesExportedChatInviteArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of MessagesExportedChatInvite.
func (s MessagesExportedChatInviteArray) Retain(keep func(x MessagesExportedChatInvite) bool) MessagesExportedChatInviteArray {
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
func (s MessagesExportedChatInviteArray) First() (v MessagesExportedChatInvite, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s MessagesExportedChatInviteArray) Last() (v MessagesExportedChatInvite, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *MessagesExportedChatInviteArray) PopFirst() (v MessagesExportedChatInvite, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero MessagesExportedChatInvite
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *MessagesExportedChatInviteArray) Pop() (v MessagesExportedChatInvite, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// MessagesExportedChatInviteReplacedArray is adapter for slice of MessagesExportedChatInviteReplaced.
type MessagesExportedChatInviteReplacedArray []MessagesExportedChatInviteReplaced

// Sort sorts slice of MessagesExportedChatInviteReplaced.
func (s MessagesExportedChatInviteReplacedArray) Sort(less func(a, b MessagesExportedChatInviteReplaced) bool) MessagesExportedChatInviteReplacedArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of MessagesExportedChatInviteReplaced.
func (s MessagesExportedChatInviteReplacedArray) SortStable(less func(a, b MessagesExportedChatInviteReplaced) bool) MessagesExportedChatInviteReplacedArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of MessagesExportedChatInviteReplaced.
func (s MessagesExportedChatInviteReplacedArray) Retain(keep func(x MessagesExportedChatInviteReplaced) bool) MessagesExportedChatInviteReplacedArray {
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
func (s MessagesExportedChatInviteReplacedArray) First() (v MessagesExportedChatInviteReplaced, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s MessagesExportedChatInviteReplacedArray) Last() (v MessagesExportedChatInviteReplaced, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *MessagesExportedChatInviteReplacedArray) PopFirst() (v MessagesExportedChatInviteReplaced, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero MessagesExportedChatInviteReplaced
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *MessagesExportedChatInviteReplacedArray) Pop() (v MessagesExportedChatInviteReplaced, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
