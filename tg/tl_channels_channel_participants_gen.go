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

// ChannelsChannelParticipants represents TL type `channels.channelParticipants#9ab0feaf`.
// Represents multiple channel participants
//
// See https://core.telegram.org/constructor/channels.channelParticipants for reference.
type ChannelsChannelParticipants struct {
	// Total number of participants that correspond to the given query
	Count int
	// Participants
	Participants []ChannelParticipantClass
	// Chats field of ChannelsChannelParticipants.
	Chats []ChatClass
	// Users mentioned in participant info
	Users []UserClass
}

// ChannelsChannelParticipantsTypeID is TL type id of ChannelsChannelParticipants.
const ChannelsChannelParticipantsTypeID = 0x9ab0feaf

func (c *ChannelsChannelParticipants) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.Count == 0) {
		return false
	}
	if !(c.Participants == nil) {
		return false
	}
	if !(c.Chats == nil) {
		return false
	}
	if !(c.Users == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *ChannelsChannelParticipants) String() string {
	if c == nil {
		return "ChannelsChannelParticipants(nil)"
	}
	type Alias ChannelsChannelParticipants
	return fmt.Sprintf("ChannelsChannelParticipants%+v", Alias(*c))
}

// FillFrom fills ChannelsChannelParticipants from given interface.
func (c *ChannelsChannelParticipants) FillFrom(from interface {
	GetCount() (value int)
	GetParticipants() (value []ChannelParticipantClass)
	GetChats() (value []ChatClass)
	GetUsers() (value []UserClass)
}) {
	c.Count = from.GetCount()
	c.Participants = from.GetParticipants()
	c.Chats = from.GetChats()
	c.Users = from.GetUsers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChannelsChannelParticipants) TypeID() uint32 {
	return ChannelsChannelParticipantsTypeID
}

// TypeName returns name of type in TL schema.
func (*ChannelsChannelParticipants) TypeName() string {
	return "channels.channelParticipants"
}

// TypeInfo returns info about TL type.
func (c *ChannelsChannelParticipants) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "channels.channelParticipants",
		ID:   ChannelsChannelParticipantsTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Count",
			SchemaName: "count",
		},
		{
			Name:       "Participants",
			SchemaName: "participants",
		},
		{
			Name:       "Chats",
			SchemaName: "chats",
		},
		{
			Name:       "Users",
			SchemaName: "users",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *ChannelsChannelParticipants) Encode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode channels.channelParticipants#9ab0feaf as nil")
	}
	b.PutID(ChannelsChannelParticipantsTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *ChannelsChannelParticipants) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode channels.channelParticipants#9ab0feaf as nil")
	}
	b.PutInt(c.Count)
	b.PutVectorHeader(len(c.Participants))
	for idx, v := range c.Participants {
		if v == nil {
			return fmt.Errorf("unable to encode channels.channelParticipants#9ab0feaf: field participants element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode channels.channelParticipants#9ab0feaf: field participants element with index %d: %w", idx, err)
		}
	}
	b.PutVectorHeader(len(c.Chats))
	for idx, v := range c.Chats {
		if v == nil {
			return fmt.Errorf("unable to encode channels.channelParticipants#9ab0feaf: field chats element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode channels.channelParticipants#9ab0feaf: field chats element with index %d: %w", idx, err)
		}
	}
	b.PutVectorHeader(len(c.Users))
	for idx, v := range c.Users {
		if v == nil {
			return fmt.Errorf("unable to encode channels.channelParticipants#9ab0feaf: field users element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode channels.channelParticipants#9ab0feaf: field users element with index %d: %w", idx, err)
		}
	}
	return nil
}

// GetCount returns value of Count field.
func (c *ChannelsChannelParticipants) GetCount() (value int) {
	return c.Count
}

// GetParticipants returns value of Participants field.
func (c *ChannelsChannelParticipants) GetParticipants() (value []ChannelParticipantClass) {
	return c.Participants
}

// MapParticipants returns field Participants wrapped in ChannelParticipantClassArray helper.
func (c *ChannelsChannelParticipants) MapParticipants() (value ChannelParticipantClassArray) {
	return ChannelParticipantClassArray(c.Participants)
}

// GetChats returns value of Chats field.
func (c *ChannelsChannelParticipants) GetChats() (value []ChatClass) {
	return c.Chats
}

// MapChats returns field Chats wrapped in ChatClassArray helper.
func (c *ChannelsChannelParticipants) MapChats() (value ChatClassArray) {
	return ChatClassArray(c.Chats)
}

// GetUsers returns value of Users field.
func (c *ChannelsChannelParticipants) GetUsers() (value []UserClass) {
	return c.Users
}

// MapUsers returns field Users wrapped in UserClassArray helper.
func (c *ChannelsChannelParticipants) MapUsers() (value UserClassArray) {
	return UserClassArray(c.Users)
}

// Decode implements bin.Decoder.
func (c *ChannelsChannelParticipants) Decode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode channels.channelParticipants#9ab0feaf to nil")
	}
	if err := b.ConsumeID(ChannelsChannelParticipantsTypeID); err != nil {
		return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: %w", err)
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *ChannelsChannelParticipants) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode channels.channelParticipants#9ab0feaf to nil")
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: field count: %w", err)
		}
		c.Count = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: field participants: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeChannelParticipant(b)
			if err != nil {
				return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: field participants: %w", err)
			}
			c.Participants = append(c.Participants, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: field chats: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeChat(b)
			if err != nil {
				return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: field chats: %w", err)
			}
			c.Chats = append(c.Chats, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: field users: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeUser(b)
			if err != nil {
				return fmt.Errorf("unable to decode channels.channelParticipants#9ab0feaf: field users: %w", err)
			}
			c.Users = append(c.Users, value)
		}
	}
	return nil
}

// construct implements constructor of ChannelsChannelParticipantsClass.
func (c ChannelsChannelParticipants) construct() ChannelsChannelParticipantsClass { return &c }

// Ensuring interfaces in compile-time for ChannelsChannelParticipants.
var (
	_ bin.Encoder     = &ChannelsChannelParticipants{}
	_ bin.Decoder     = &ChannelsChannelParticipants{}
	_ bin.BareEncoder = &ChannelsChannelParticipants{}
	_ bin.BareDecoder = &ChannelsChannelParticipants{}

	_ ChannelsChannelParticipantsClass = &ChannelsChannelParticipants{}
)

// ChannelsChannelParticipantsNotModified represents TL type `channels.channelParticipantsNotModified#f0173fe9`.
// No new participant info could be found
//
// See https://core.telegram.org/constructor/channels.channelParticipantsNotModified for reference.
type ChannelsChannelParticipantsNotModified struct {
}

// ChannelsChannelParticipantsNotModifiedTypeID is TL type id of ChannelsChannelParticipantsNotModified.
const ChannelsChannelParticipantsNotModifiedTypeID = 0xf0173fe9

func (c *ChannelsChannelParticipantsNotModified) Zero() bool {
	if c == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (c *ChannelsChannelParticipantsNotModified) String() string {
	if c == nil {
		return "ChannelsChannelParticipantsNotModified(nil)"
	}
	type Alias ChannelsChannelParticipantsNotModified
	return fmt.Sprintf("ChannelsChannelParticipantsNotModified%+v", Alias(*c))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChannelsChannelParticipantsNotModified) TypeID() uint32 {
	return ChannelsChannelParticipantsNotModifiedTypeID
}

// TypeName returns name of type in TL schema.
func (*ChannelsChannelParticipantsNotModified) TypeName() string {
	return "channels.channelParticipantsNotModified"
}

// TypeInfo returns info about TL type.
func (c *ChannelsChannelParticipantsNotModified) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "channels.channelParticipantsNotModified",
		ID:   ChannelsChannelParticipantsNotModifiedTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (c *ChannelsChannelParticipantsNotModified) Encode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode channels.channelParticipantsNotModified#f0173fe9 as nil")
	}
	b.PutID(ChannelsChannelParticipantsNotModifiedTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *ChannelsChannelParticipantsNotModified) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't encode channels.channelParticipantsNotModified#f0173fe9 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (c *ChannelsChannelParticipantsNotModified) Decode(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode channels.channelParticipantsNotModified#f0173fe9 to nil")
	}
	if err := b.ConsumeID(ChannelsChannelParticipantsNotModifiedTypeID); err != nil {
		return fmt.Errorf("unable to decode channels.channelParticipantsNotModified#f0173fe9: %w", err)
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *ChannelsChannelParticipantsNotModified) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return fmt.Errorf("can't decode channels.channelParticipantsNotModified#f0173fe9 to nil")
	}
	return nil
}

// construct implements constructor of ChannelsChannelParticipantsClass.
func (c ChannelsChannelParticipantsNotModified) construct() ChannelsChannelParticipantsClass {
	return &c
}

// Ensuring interfaces in compile-time for ChannelsChannelParticipantsNotModified.
var (
	_ bin.Encoder     = &ChannelsChannelParticipantsNotModified{}
	_ bin.Decoder     = &ChannelsChannelParticipantsNotModified{}
	_ bin.BareEncoder = &ChannelsChannelParticipantsNotModified{}
	_ bin.BareDecoder = &ChannelsChannelParticipantsNotModified{}

	_ ChannelsChannelParticipantsClass = &ChannelsChannelParticipantsNotModified{}
)

// ChannelsChannelParticipantsClass represents channels.ChannelParticipants generic type.
//
// See https://core.telegram.org/type/channels.ChannelParticipants for reference.
//
// Example:
//  g, err := tg.DecodeChannelsChannelParticipants(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.ChannelsChannelParticipants: // channels.channelParticipants#9ab0feaf
//  case *tg.ChannelsChannelParticipantsNotModified: // channels.channelParticipantsNotModified#f0173fe9
//  default: panic(v)
//  }
type ChannelsChannelParticipantsClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() ChannelsChannelParticipantsClass

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

	// AsModified tries to map ChannelsChannelParticipantsClass to ChannelsChannelParticipants.
	AsModified() (*ChannelsChannelParticipants, bool)
}

// AsModified tries to map ChannelsChannelParticipants to ChannelsChannelParticipants.
func (c *ChannelsChannelParticipants) AsModified() (*ChannelsChannelParticipants, bool) {
	return c, true
}

// AsModified tries to map ChannelsChannelParticipantsNotModified to ChannelsChannelParticipants.
func (c *ChannelsChannelParticipantsNotModified) AsModified() (*ChannelsChannelParticipants, bool) {
	return nil, false
}

// DecodeChannelsChannelParticipants implements binary de-serialization for ChannelsChannelParticipantsClass.
func DecodeChannelsChannelParticipants(buf *bin.Buffer) (ChannelsChannelParticipantsClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case ChannelsChannelParticipantsTypeID:
		// Decoding channels.channelParticipants#9ab0feaf.
		v := ChannelsChannelParticipants{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode ChannelsChannelParticipantsClass: %w", err)
		}
		return &v, nil
	case ChannelsChannelParticipantsNotModifiedTypeID:
		// Decoding channels.channelParticipantsNotModified#f0173fe9.
		v := ChannelsChannelParticipantsNotModified{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode ChannelsChannelParticipantsClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode ChannelsChannelParticipantsClass: %w", bin.NewUnexpectedID(id))
	}
}

// ChannelsChannelParticipants boxes the ChannelsChannelParticipantsClass providing a helper.
type ChannelsChannelParticipantsBox struct {
	ChannelParticipants ChannelsChannelParticipantsClass
}

// Decode implements bin.Decoder for ChannelsChannelParticipantsBox.
func (b *ChannelsChannelParticipantsBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode ChannelsChannelParticipantsBox to nil")
	}
	v, err := DecodeChannelsChannelParticipants(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.ChannelParticipants = v
	return nil
}

// Encode implements bin.Encode for ChannelsChannelParticipantsBox.
func (b *ChannelsChannelParticipantsBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.ChannelParticipants == nil {
		return fmt.Errorf("unable to encode ChannelsChannelParticipantsClass as nil")
	}
	return b.ChannelParticipants.Encode(buf)
}

// ChannelsChannelParticipantsClassArray is adapter for slice of ChannelsChannelParticipantsClass.
type ChannelsChannelParticipantsClassArray []ChannelsChannelParticipantsClass

// Sort sorts slice of ChannelsChannelParticipantsClass.
func (s ChannelsChannelParticipantsClassArray) Sort(less func(a, b ChannelsChannelParticipantsClass) bool) ChannelsChannelParticipantsClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of ChannelsChannelParticipantsClass.
func (s ChannelsChannelParticipantsClassArray) SortStable(less func(a, b ChannelsChannelParticipantsClass) bool) ChannelsChannelParticipantsClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of ChannelsChannelParticipantsClass.
func (s ChannelsChannelParticipantsClassArray) Retain(keep func(x ChannelsChannelParticipantsClass) bool) ChannelsChannelParticipantsClassArray {
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
func (s ChannelsChannelParticipantsClassArray) First() (v ChannelsChannelParticipantsClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s ChannelsChannelParticipantsClassArray) Last() (v ChannelsChannelParticipantsClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *ChannelsChannelParticipantsClassArray) PopFirst() (v ChannelsChannelParticipantsClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero ChannelsChannelParticipantsClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *ChannelsChannelParticipantsClassArray) Pop() (v ChannelsChannelParticipantsClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsChannelsChannelParticipants returns copy with only ChannelsChannelParticipants constructors.
func (s ChannelsChannelParticipantsClassArray) AsChannelsChannelParticipants() (to ChannelsChannelParticipantsArray) {
	for _, elem := range s {
		value, ok := elem.(*ChannelsChannelParticipants)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AppendOnlyModified appends only Modified constructors to
// given slice.
func (s ChannelsChannelParticipantsClassArray) AppendOnlyModified(to []*ChannelsChannelParticipants) []*ChannelsChannelParticipants {
	for _, elem := range s {
		value, ok := elem.AsModified()
		if !ok {
			continue
		}
		to = append(to, value)
	}

	return to
}

// AsModified returns copy with only Modified constructors.
func (s ChannelsChannelParticipantsClassArray) AsModified() (to []*ChannelsChannelParticipants) {
	return s.AppendOnlyModified(to)
}

// FirstAsModified returns first element of slice (if exists).
func (s ChannelsChannelParticipantsClassArray) FirstAsModified() (v *ChannelsChannelParticipants, ok bool) {
	value, ok := s.First()
	if !ok {
		return
	}
	return value.AsModified()
}

// LastAsModified returns last element of slice (if exists).
func (s ChannelsChannelParticipantsClassArray) LastAsModified() (v *ChannelsChannelParticipants, ok bool) {
	value, ok := s.Last()
	if !ok {
		return
	}
	return value.AsModified()
}

// PopFirstAsModified returns element of slice (if exists).
func (s *ChannelsChannelParticipantsClassArray) PopFirstAsModified() (v *ChannelsChannelParticipants, ok bool) {
	value, ok := s.PopFirst()
	if !ok {
		return
	}
	return value.AsModified()
}

// PopAsModified returns element of slice (if exists).
func (s *ChannelsChannelParticipantsClassArray) PopAsModified() (v *ChannelsChannelParticipants, ok bool) {
	value, ok := s.Pop()
	if !ok {
		return
	}
	return value.AsModified()
}

// ChannelsChannelParticipantsArray is adapter for slice of ChannelsChannelParticipants.
type ChannelsChannelParticipantsArray []ChannelsChannelParticipants

// Sort sorts slice of ChannelsChannelParticipants.
func (s ChannelsChannelParticipantsArray) Sort(less func(a, b ChannelsChannelParticipants) bool) ChannelsChannelParticipantsArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of ChannelsChannelParticipants.
func (s ChannelsChannelParticipantsArray) SortStable(less func(a, b ChannelsChannelParticipants) bool) ChannelsChannelParticipantsArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of ChannelsChannelParticipants.
func (s ChannelsChannelParticipantsArray) Retain(keep func(x ChannelsChannelParticipants) bool) ChannelsChannelParticipantsArray {
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
func (s ChannelsChannelParticipantsArray) First() (v ChannelsChannelParticipants, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s ChannelsChannelParticipantsArray) Last() (v ChannelsChannelParticipants, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *ChannelsChannelParticipantsArray) PopFirst() (v ChannelsChannelParticipants, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero ChannelsChannelParticipants
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *ChannelsChannelParticipantsArray) Pop() (v ChannelsChannelParticipants, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
