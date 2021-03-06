// Code generated by gotdgen, DO NOT EDIT.

package td

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

// UserAuth represents TL type `user.auth#f4815592`.
//
// See https://localhost:80/doc/constructor/user.auth for reference.
type UserAuth struct {
	// Foo field of UserAuth.
	Foo string
}

// UserAuthTypeID is TL type id of UserAuth.
const UserAuthTypeID = 0xf4815592

func (a *UserAuth) Zero() bool {
	if a == nil {
		return true
	}
	if !(a.Foo == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (a *UserAuth) String() string {
	if a == nil {
		return "UserAuth(nil)"
	}
	type Alias UserAuth
	return fmt.Sprintf("UserAuth%+v", Alias(*a))
}

// FillFrom fills UserAuth from given interface.
func (a *UserAuth) FillFrom(from interface {
	GetFoo() (value string)
}) {
	a.Foo = from.GetFoo()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*UserAuth) TypeID() uint32 {
	return UserAuthTypeID
}

// TypeName returns name of type in TL schema.
func (*UserAuth) TypeName() string {
	return "user.auth"
}

// TypeInfo returns info about TL type.
func (a *UserAuth) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "user.auth",
		ID:   UserAuthTypeID,
	}
	if a == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Foo",
			SchemaName: "foo",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (a *UserAuth) Encode(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't encode user.auth#f4815592 as nil")
	}
	b.PutID(UserAuthTypeID)
	return a.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (a *UserAuth) EncodeBare(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't encode user.auth#f4815592 as nil")
	}
	b.PutString(a.Foo)
	return nil
}

// GetFoo returns value of Foo field.
func (a *UserAuth) GetFoo() (value string) {
	return a.Foo
}

// Decode implements bin.Decoder.
func (a *UserAuth) Decode(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't decode user.auth#f4815592 to nil")
	}
	if err := b.ConsumeID(UserAuthTypeID); err != nil {
		return fmt.Errorf("unable to decode user.auth#f4815592: %w", err)
	}
	return a.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (a *UserAuth) DecodeBare(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't decode user.auth#f4815592 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode user.auth#f4815592: field foo: %w", err)
		}
		a.Foo = value
	}
	return nil
}

// construct implements constructor of UserAuthClass.
func (a UserAuth) construct() UserAuthClass { return &a }

// Ensuring interfaces in compile-time for UserAuth.
var (
	_ bin.Encoder     = &UserAuth{}
	_ bin.Decoder     = &UserAuth{}
	_ bin.BareEncoder = &UserAuth{}
	_ bin.BareDecoder = &UserAuth{}

	_ UserAuthClass = &UserAuth{}
)

// UserAuthPassword represents TL type `user.authPassword#5981e317`.
//
// See https://localhost:80/doc/constructor/user.authPassword for reference.
type UserAuthPassword struct {
	// Pwd field of UserAuthPassword.
	Pwd string
}

// UserAuthPasswordTypeID is TL type id of UserAuthPassword.
const UserAuthPasswordTypeID = 0x5981e317

func (a *UserAuthPassword) Zero() bool {
	if a == nil {
		return true
	}
	if !(a.Pwd == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (a *UserAuthPassword) String() string {
	if a == nil {
		return "UserAuthPassword(nil)"
	}
	type Alias UserAuthPassword
	return fmt.Sprintf("UserAuthPassword%+v", Alias(*a))
}

// FillFrom fills UserAuthPassword from given interface.
func (a *UserAuthPassword) FillFrom(from interface {
	GetPwd() (value string)
}) {
	a.Pwd = from.GetPwd()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*UserAuthPassword) TypeID() uint32 {
	return UserAuthPasswordTypeID
}

// TypeName returns name of type in TL schema.
func (*UserAuthPassword) TypeName() string {
	return "user.authPassword"
}

// TypeInfo returns info about TL type.
func (a *UserAuthPassword) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "user.authPassword",
		ID:   UserAuthPasswordTypeID,
	}
	if a == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Pwd",
			SchemaName: "pwd",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (a *UserAuthPassword) Encode(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't encode user.authPassword#5981e317 as nil")
	}
	b.PutID(UserAuthPasswordTypeID)
	return a.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (a *UserAuthPassword) EncodeBare(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't encode user.authPassword#5981e317 as nil")
	}
	b.PutString(a.Pwd)
	return nil
}

// GetPwd returns value of Pwd field.
func (a *UserAuthPassword) GetPwd() (value string) {
	return a.Pwd
}

// Decode implements bin.Decoder.
func (a *UserAuthPassword) Decode(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't decode user.authPassword#5981e317 to nil")
	}
	if err := b.ConsumeID(UserAuthPasswordTypeID); err != nil {
		return fmt.Errorf("unable to decode user.authPassword#5981e317: %w", err)
	}
	return a.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (a *UserAuthPassword) DecodeBare(b *bin.Buffer) error {
	if a == nil {
		return fmt.Errorf("can't decode user.authPassword#5981e317 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode user.authPassword#5981e317: field pwd: %w", err)
		}
		a.Pwd = value
	}
	return nil
}

// construct implements constructor of UserAuthClass.
func (a UserAuthPassword) construct() UserAuthClass { return &a }

// Ensuring interfaces in compile-time for UserAuthPassword.
var (
	_ bin.Encoder     = &UserAuthPassword{}
	_ bin.Decoder     = &UserAuthPassword{}
	_ bin.BareEncoder = &UserAuthPassword{}
	_ bin.BareDecoder = &UserAuthPassword{}

	_ UserAuthClass = &UserAuthPassword{}
)

// UserAuthClass represents user.Auth generic type.
//
// See https://localhost:80/doc/type/user.Auth for reference.
//
// Example:
//  g, err := td.DecodeUserAuth(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *td.UserAuth: // user.auth#f4815592
//  case *td.UserAuthPassword: // user.authPassword#5981e317
//  default: panic(v)
//  }
type UserAuthClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() UserAuthClass

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

// DecodeUserAuth implements binary de-serialization for UserAuthClass.
func DecodeUserAuth(buf *bin.Buffer) (UserAuthClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case UserAuthTypeID:
		// Decoding user.auth#f4815592.
		v := UserAuth{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode UserAuthClass: %w", err)
		}
		return &v, nil
	case UserAuthPasswordTypeID:
		// Decoding user.authPassword#5981e317.
		v := UserAuthPassword{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode UserAuthClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode UserAuthClass: %w", bin.NewUnexpectedID(id))
	}
}

// UserAuth boxes the UserAuthClass providing a helper.
type UserAuthBox struct {
	Auth UserAuthClass
}

// Decode implements bin.Decoder for UserAuthBox.
func (b *UserAuthBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode UserAuthBox to nil")
	}
	v, err := DecodeUserAuth(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.Auth = v
	return nil
}

// Encode implements bin.Encode for UserAuthBox.
func (b *UserAuthBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.Auth == nil {
		return fmt.Errorf("unable to encode UserAuthClass as nil")
	}
	return b.Auth.Encode(buf)
}

// UserAuthClassArray is adapter for slice of UserAuthClass.
type UserAuthClassArray []UserAuthClass

// Sort sorts slice of UserAuthClass.
func (s UserAuthClassArray) Sort(less func(a, b UserAuthClass) bool) UserAuthClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UserAuthClass.
func (s UserAuthClassArray) SortStable(less func(a, b UserAuthClass) bool) UserAuthClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UserAuthClass.
func (s UserAuthClassArray) Retain(keep func(x UserAuthClass) bool) UserAuthClassArray {
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
func (s UserAuthClassArray) First() (v UserAuthClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UserAuthClassArray) Last() (v UserAuthClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UserAuthClassArray) PopFirst() (v UserAuthClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UserAuthClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UserAuthClassArray) Pop() (v UserAuthClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsUserAuth returns copy with only UserAuth constructors.
func (s UserAuthClassArray) AsUserAuth() (to UserAuthArray) {
	for _, elem := range s {
		value, ok := elem.(*UserAuth)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsUserAuthPassword returns copy with only UserAuthPassword constructors.
func (s UserAuthClassArray) AsUserAuthPassword() (to UserAuthPasswordArray) {
	for _, elem := range s {
		value, ok := elem.(*UserAuthPassword)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// UserAuthArray is adapter for slice of UserAuth.
type UserAuthArray []UserAuth

// Sort sorts slice of UserAuth.
func (s UserAuthArray) Sort(less func(a, b UserAuth) bool) UserAuthArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UserAuth.
func (s UserAuthArray) SortStable(less func(a, b UserAuth) bool) UserAuthArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UserAuth.
func (s UserAuthArray) Retain(keep func(x UserAuth) bool) UserAuthArray {
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
func (s UserAuthArray) First() (v UserAuth, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UserAuthArray) Last() (v UserAuth, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UserAuthArray) PopFirst() (v UserAuth, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UserAuth
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UserAuthArray) Pop() (v UserAuth, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// UserAuthPasswordArray is adapter for slice of UserAuthPassword.
type UserAuthPasswordArray []UserAuthPassword

// Sort sorts slice of UserAuthPassword.
func (s UserAuthPasswordArray) Sort(less func(a, b UserAuthPassword) bool) UserAuthPasswordArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UserAuthPassword.
func (s UserAuthPasswordArray) SortStable(less func(a, b UserAuthPassword) bool) UserAuthPasswordArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UserAuthPassword.
func (s UserAuthPasswordArray) Retain(keep func(x UserAuthPassword) bool) UserAuthPasswordArray {
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
func (s UserAuthPasswordArray) First() (v UserAuthPassword, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UserAuthPasswordArray) Last() (v UserAuthPassword, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UserAuthPasswordArray) PopFirst() (v UserAuthPassword, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UserAuthPassword
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UserAuthPasswordArray) Pop() (v UserAuthPassword, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
