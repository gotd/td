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

// ContactsBlocked represents TL type `contacts.blocked#ade1591`.
// Full list of blocked users.
//
// See https://core.telegram.org/constructor/contacts.blocked for reference.
type ContactsBlocked struct {
	// List of blocked users
	Blocked []PeerBlocked
	// Blocked chats
	Chats []ChatClass
	// List of users
	Users []UserClass
}

// ContactsBlockedTypeID is TL type id of ContactsBlocked.
const ContactsBlockedTypeID = 0xade1591

func (b *ContactsBlocked) Zero() bool {
	if b == nil {
		return true
	}
	if !(b.Blocked == nil) {
		return false
	}
	if !(b.Chats == nil) {
		return false
	}
	if !(b.Users == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (b *ContactsBlocked) String() string {
	if b == nil {
		return "ContactsBlocked(nil)"
	}
	type Alias ContactsBlocked
	return fmt.Sprintf("ContactsBlocked%+v", Alias(*b))
}

// FillFrom fills ContactsBlocked from given interface.
func (b *ContactsBlocked) FillFrom(from interface {
	GetBlocked() (value []PeerBlocked)
	GetChats() (value []ChatClass)
	GetUsers() (value []UserClass)
}) {
	b.Blocked = from.GetBlocked()
	b.Chats = from.GetChats()
	b.Users = from.GetUsers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ContactsBlocked) TypeID() uint32 {
	return ContactsBlockedTypeID
}

// TypeName returns name of type in TL schema.
func (*ContactsBlocked) TypeName() string {
	return "contacts.blocked"
}

// TypeInfo returns info about TL type.
func (b *ContactsBlocked) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "contacts.blocked",
		ID:   ContactsBlockedTypeID,
	}
	if b == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Blocked",
			SchemaName: "blocked",
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
func (b *ContactsBlocked) Encode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't encode contacts.blocked#ade1591 as nil")
	}
	buf.PutID(ContactsBlockedTypeID)
	return b.EncodeBare(buf)
}

// EncodeBare implements bin.BareEncoder.
func (b *ContactsBlocked) EncodeBare(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't encode contacts.blocked#ade1591 as nil")
	}
	buf.PutVectorHeader(len(b.Blocked))
	for idx, v := range b.Blocked {
		if err := v.Encode(buf); err != nil {
			return fmt.Errorf("unable to encode contacts.blocked#ade1591: field blocked element with index %d: %w", idx, err)
		}
	}
	buf.PutVectorHeader(len(b.Chats))
	for idx, v := range b.Chats {
		if v == nil {
			return fmt.Errorf("unable to encode contacts.blocked#ade1591: field chats element with index %d is nil", idx)
		}
		if err := v.Encode(buf); err != nil {
			return fmt.Errorf("unable to encode contacts.blocked#ade1591: field chats element with index %d: %w", idx, err)
		}
	}
	buf.PutVectorHeader(len(b.Users))
	for idx, v := range b.Users {
		if v == nil {
			return fmt.Errorf("unable to encode contacts.blocked#ade1591: field users element with index %d is nil", idx)
		}
		if err := v.Encode(buf); err != nil {
			return fmt.Errorf("unable to encode contacts.blocked#ade1591: field users element with index %d: %w", idx, err)
		}
	}
	return nil
}

// GetBlocked returns value of Blocked field.
func (b *ContactsBlocked) GetBlocked() (value []PeerBlocked) {
	return b.Blocked
}

// GetChats returns value of Chats field.
func (b *ContactsBlocked) GetChats() (value []ChatClass) {
	return b.Chats
}

// MapChats returns field Chats wrapped in ChatClassArray helper.
func (b *ContactsBlocked) MapChats() (value ChatClassArray) {
	return ChatClassArray(b.Chats)
}

// GetUsers returns value of Users field.
func (b *ContactsBlocked) GetUsers() (value []UserClass) {
	return b.Users
}

// MapUsers returns field Users wrapped in UserClassArray helper.
func (b *ContactsBlocked) MapUsers() (value UserClassArray) {
	return UserClassArray(b.Users)
}

// Decode implements bin.Decoder.
func (b *ContactsBlocked) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't decode contacts.blocked#ade1591 to nil")
	}
	if err := buf.ConsumeID(ContactsBlockedTypeID); err != nil {
		return fmt.Errorf("unable to decode contacts.blocked#ade1591: %w", err)
	}
	return b.DecodeBare(buf)
}

// DecodeBare implements bin.BareDecoder.
func (b *ContactsBlocked) DecodeBare(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't decode contacts.blocked#ade1591 to nil")
	}
	{
		headerLen, err := buf.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode contacts.blocked#ade1591: field blocked: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value PeerBlocked
			if err := value.Decode(buf); err != nil {
				return fmt.Errorf("unable to decode contacts.blocked#ade1591: field blocked: %w", err)
			}
			b.Blocked = append(b.Blocked, value)
		}
	}
	{
		headerLen, err := buf.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode contacts.blocked#ade1591: field chats: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeChat(buf)
			if err != nil {
				return fmt.Errorf("unable to decode contacts.blocked#ade1591: field chats: %w", err)
			}
			b.Chats = append(b.Chats, value)
		}
	}
	{
		headerLen, err := buf.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode contacts.blocked#ade1591: field users: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeUser(buf)
			if err != nil {
				return fmt.Errorf("unable to decode contacts.blocked#ade1591: field users: %w", err)
			}
			b.Users = append(b.Users, value)
		}
	}
	return nil
}

// construct implements constructor of ContactsBlockedClass.
func (b ContactsBlocked) construct() ContactsBlockedClass { return &b }

// Ensuring interfaces in compile-time for ContactsBlocked.
var (
	_ bin.Encoder     = &ContactsBlocked{}
	_ bin.Decoder     = &ContactsBlocked{}
	_ bin.BareEncoder = &ContactsBlocked{}
	_ bin.BareDecoder = &ContactsBlocked{}

	_ ContactsBlockedClass = &ContactsBlocked{}
)

// ContactsBlockedSlice represents TL type `contacts.blockedSlice#e1664194`.
// Incomplete list of blocked users.
//
// See https://core.telegram.org/constructor/contacts.blockedSlice for reference.
type ContactsBlockedSlice struct {
	// Total number of elements in the list
	Count int
	// List of blocked users
	Blocked []PeerBlocked
	// Blocked chats
	Chats []ChatClass
	// List of users
	Users []UserClass
}

// ContactsBlockedSliceTypeID is TL type id of ContactsBlockedSlice.
const ContactsBlockedSliceTypeID = 0xe1664194

func (b *ContactsBlockedSlice) Zero() bool {
	if b == nil {
		return true
	}
	if !(b.Count == 0) {
		return false
	}
	if !(b.Blocked == nil) {
		return false
	}
	if !(b.Chats == nil) {
		return false
	}
	if !(b.Users == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (b *ContactsBlockedSlice) String() string {
	if b == nil {
		return "ContactsBlockedSlice(nil)"
	}
	type Alias ContactsBlockedSlice
	return fmt.Sprintf("ContactsBlockedSlice%+v", Alias(*b))
}

// FillFrom fills ContactsBlockedSlice from given interface.
func (b *ContactsBlockedSlice) FillFrom(from interface {
	GetCount() (value int)
	GetBlocked() (value []PeerBlocked)
	GetChats() (value []ChatClass)
	GetUsers() (value []UserClass)
}) {
	b.Count = from.GetCount()
	b.Blocked = from.GetBlocked()
	b.Chats = from.GetChats()
	b.Users = from.GetUsers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ContactsBlockedSlice) TypeID() uint32 {
	return ContactsBlockedSliceTypeID
}

// TypeName returns name of type in TL schema.
func (*ContactsBlockedSlice) TypeName() string {
	return "contacts.blockedSlice"
}

// TypeInfo returns info about TL type.
func (b *ContactsBlockedSlice) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "contacts.blockedSlice",
		ID:   ContactsBlockedSliceTypeID,
	}
	if b == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Count",
			SchemaName: "count",
		},
		{
			Name:       "Blocked",
			SchemaName: "blocked",
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
func (b *ContactsBlockedSlice) Encode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't encode contacts.blockedSlice#e1664194 as nil")
	}
	buf.PutID(ContactsBlockedSliceTypeID)
	return b.EncodeBare(buf)
}

// EncodeBare implements bin.BareEncoder.
func (b *ContactsBlockedSlice) EncodeBare(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't encode contacts.blockedSlice#e1664194 as nil")
	}
	buf.PutInt(b.Count)
	buf.PutVectorHeader(len(b.Blocked))
	for idx, v := range b.Blocked {
		if err := v.Encode(buf); err != nil {
			return fmt.Errorf("unable to encode contacts.blockedSlice#e1664194: field blocked element with index %d: %w", idx, err)
		}
	}
	buf.PutVectorHeader(len(b.Chats))
	for idx, v := range b.Chats {
		if v == nil {
			return fmt.Errorf("unable to encode contacts.blockedSlice#e1664194: field chats element with index %d is nil", idx)
		}
		if err := v.Encode(buf); err != nil {
			return fmt.Errorf("unable to encode contacts.blockedSlice#e1664194: field chats element with index %d: %w", idx, err)
		}
	}
	buf.PutVectorHeader(len(b.Users))
	for idx, v := range b.Users {
		if v == nil {
			return fmt.Errorf("unable to encode contacts.blockedSlice#e1664194: field users element with index %d is nil", idx)
		}
		if err := v.Encode(buf); err != nil {
			return fmt.Errorf("unable to encode contacts.blockedSlice#e1664194: field users element with index %d: %w", idx, err)
		}
	}
	return nil
}

// GetCount returns value of Count field.
func (b *ContactsBlockedSlice) GetCount() (value int) {
	return b.Count
}

// GetBlocked returns value of Blocked field.
func (b *ContactsBlockedSlice) GetBlocked() (value []PeerBlocked) {
	return b.Blocked
}

// GetChats returns value of Chats field.
func (b *ContactsBlockedSlice) GetChats() (value []ChatClass) {
	return b.Chats
}

// MapChats returns field Chats wrapped in ChatClassArray helper.
func (b *ContactsBlockedSlice) MapChats() (value ChatClassArray) {
	return ChatClassArray(b.Chats)
}

// GetUsers returns value of Users field.
func (b *ContactsBlockedSlice) GetUsers() (value []UserClass) {
	return b.Users
}

// MapUsers returns field Users wrapped in UserClassArray helper.
func (b *ContactsBlockedSlice) MapUsers() (value UserClassArray) {
	return UserClassArray(b.Users)
}

// Decode implements bin.Decoder.
func (b *ContactsBlockedSlice) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't decode contacts.blockedSlice#e1664194 to nil")
	}
	if err := buf.ConsumeID(ContactsBlockedSliceTypeID); err != nil {
		return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: %w", err)
	}
	return b.DecodeBare(buf)
}

// DecodeBare implements bin.BareDecoder.
func (b *ContactsBlockedSlice) DecodeBare(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't decode contacts.blockedSlice#e1664194 to nil")
	}
	{
		value, err := buf.Int()
		if err != nil {
			return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: field count: %w", err)
		}
		b.Count = value
	}
	{
		headerLen, err := buf.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: field blocked: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value PeerBlocked
			if err := value.Decode(buf); err != nil {
				return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: field blocked: %w", err)
			}
			b.Blocked = append(b.Blocked, value)
		}
	}
	{
		headerLen, err := buf.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: field chats: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeChat(buf)
			if err != nil {
				return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: field chats: %w", err)
			}
			b.Chats = append(b.Chats, value)
		}
	}
	{
		headerLen, err := buf.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: field users: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeUser(buf)
			if err != nil {
				return fmt.Errorf("unable to decode contacts.blockedSlice#e1664194: field users: %w", err)
			}
			b.Users = append(b.Users, value)
		}
	}
	return nil
}

// construct implements constructor of ContactsBlockedClass.
func (b ContactsBlockedSlice) construct() ContactsBlockedClass { return &b }

// Ensuring interfaces in compile-time for ContactsBlockedSlice.
var (
	_ bin.Encoder     = &ContactsBlockedSlice{}
	_ bin.Decoder     = &ContactsBlockedSlice{}
	_ bin.BareEncoder = &ContactsBlockedSlice{}
	_ bin.BareDecoder = &ContactsBlockedSlice{}

	_ ContactsBlockedClass = &ContactsBlockedSlice{}
)

// ContactsBlockedClass represents contacts.Blocked generic type.
//
// See https://core.telegram.org/type/contacts.Blocked for reference.
//
// Example:
//  g, err := tg.DecodeContactsBlocked(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.ContactsBlocked: // contacts.blocked#ade1591
//  case *tg.ContactsBlockedSlice: // contacts.blockedSlice#e1664194
//  default: panic(v)
//  }
type ContactsBlockedClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() ContactsBlockedClass

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

	// List of blocked users
	GetBlocked() (value []PeerBlocked)
	// Blocked chats
	GetChats() (value []ChatClass)
	// Blocked chats
	MapChats() (value ChatClassArray)
	// List of users
	GetUsers() (value []UserClass)
	// List of users
	MapUsers() (value UserClassArray)
}

// DecodeContactsBlocked implements binary de-serialization for ContactsBlockedClass.
func DecodeContactsBlocked(buf *bin.Buffer) (ContactsBlockedClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case ContactsBlockedTypeID:
		// Decoding contacts.blocked#ade1591.
		v := ContactsBlocked{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode ContactsBlockedClass: %w", err)
		}
		return &v, nil
	case ContactsBlockedSliceTypeID:
		// Decoding contacts.blockedSlice#e1664194.
		v := ContactsBlockedSlice{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode ContactsBlockedClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode ContactsBlockedClass: %w", bin.NewUnexpectedID(id))
	}
}

// ContactsBlocked boxes the ContactsBlockedClass providing a helper.
type ContactsBlockedBox struct {
	Blocked ContactsBlockedClass
}

// Decode implements bin.Decoder for ContactsBlockedBox.
func (b *ContactsBlockedBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode ContactsBlockedBox to nil")
	}
	v, err := DecodeContactsBlocked(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.Blocked = v
	return nil
}

// Encode implements bin.Encode for ContactsBlockedBox.
func (b *ContactsBlockedBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.Blocked == nil {
		return fmt.Errorf("unable to encode ContactsBlockedClass as nil")
	}
	return b.Blocked.Encode(buf)
}

// ContactsBlockedClassArray is adapter for slice of ContactsBlockedClass.
type ContactsBlockedClassArray []ContactsBlockedClass

// Sort sorts slice of ContactsBlockedClass.
func (s ContactsBlockedClassArray) Sort(less func(a, b ContactsBlockedClass) bool) ContactsBlockedClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of ContactsBlockedClass.
func (s ContactsBlockedClassArray) SortStable(less func(a, b ContactsBlockedClass) bool) ContactsBlockedClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of ContactsBlockedClass.
func (s ContactsBlockedClassArray) Retain(keep func(x ContactsBlockedClass) bool) ContactsBlockedClassArray {
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
func (s ContactsBlockedClassArray) First() (v ContactsBlockedClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s ContactsBlockedClassArray) Last() (v ContactsBlockedClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *ContactsBlockedClassArray) PopFirst() (v ContactsBlockedClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero ContactsBlockedClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *ContactsBlockedClassArray) Pop() (v ContactsBlockedClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsContactsBlocked returns copy with only ContactsBlocked constructors.
func (s ContactsBlockedClassArray) AsContactsBlocked() (to ContactsBlockedArray) {
	for _, elem := range s {
		value, ok := elem.(*ContactsBlocked)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsContactsBlockedSlice returns copy with only ContactsBlockedSlice constructors.
func (s ContactsBlockedClassArray) AsContactsBlockedSlice() (to ContactsBlockedSliceArray) {
	for _, elem := range s {
		value, ok := elem.(*ContactsBlockedSlice)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// ContactsBlockedArray is adapter for slice of ContactsBlocked.
type ContactsBlockedArray []ContactsBlocked

// Sort sorts slice of ContactsBlocked.
func (s ContactsBlockedArray) Sort(less func(a, b ContactsBlocked) bool) ContactsBlockedArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of ContactsBlocked.
func (s ContactsBlockedArray) SortStable(less func(a, b ContactsBlocked) bool) ContactsBlockedArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of ContactsBlocked.
func (s ContactsBlockedArray) Retain(keep func(x ContactsBlocked) bool) ContactsBlockedArray {
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
func (s ContactsBlockedArray) First() (v ContactsBlocked, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s ContactsBlockedArray) Last() (v ContactsBlocked, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *ContactsBlockedArray) PopFirst() (v ContactsBlocked, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero ContactsBlocked
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *ContactsBlockedArray) Pop() (v ContactsBlocked, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// ContactsBlockedSliceArray is adapter for slice of ContactsBlockedSlice.
type ContactsBlockedSliceArray []ContactsBlockedSlice

// Sort sorts slice of ContactsBlockedSlice.
func (s ContactsBlockedSliceArray) Sort(less func(a, b ContactsBlockedSlice) bool) ContactsBlockedSliceArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of ContactsBlockedSlice.
func (s ContactsBlockedSliceArray) SortStable(less func(a, b ContactsBlockedSlice) bool) ContactsBlockedSliceArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of ContactsBlockedSlice.
func (s ContactsBlockedSliceArray) Retain(keep func(x ContactsBlockedSlice) bool) ContactsBlockedSliceArray {
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
func (s ContactsBlockedSliceArray) First() (v ContactsBlockedSlice, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s ContactsBlockedSliceArray) Last() (v ContactsBlockedSlice, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *ContactsBlockedSliceArray) PopFirst() (v ContactsBlockedSlice, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero ContactsBlockedSlice
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *ContactsBlockedSliceArray) Pop() (v ContactsBlockedSlice, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
