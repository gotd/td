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
	"github.com/gotd/td/tdjson"
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
	_ = tdjson.Encoder{}
)

// AccountEmailVerified represents TL type `account.emailVerified#2b96cd1b`.
// The email was verified correctly.
//
// See https://core.telegram.org/constructor/account.emailVerified for reference.
type AccountEmailVerified struct {
	// The verified email address.
	Email string
}

// AccountEmailVerifiedTypeID is TL type id of AccountEmailVerified.
const AccountEmailVerifiedTypeID = 0x2b96cd1b

// construct implements constructor of AccountEmailVerifiedClass.
func (e AccountEmailVerified) construct() AccountEmailVerifiedClass { return &e }

// Ensuring interfaces in compile-time for AccountEmailVerified.
var (
	_ bin.Encoder     = &AccountEmailVerified{}
	_ bin.Decoder     = &AccountEmailVerified{}
	_ bin.BareEncoder = &AccountEmailVerified{}
	_ bin.BareDecoder = &AccountEmailVerified{}

	_ AccountEmailVerifiedClass = &AccountEmailVerified{}
)

func (e *AccountEmailVerified) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Email == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *AccountEmailVerified) String() string {
	if e == nil {
		return "AccountEmailVerified(nil)"
	}
	type Alias AccountEmailVerified
	return fmt.Sprintf("AccountEmailVerified%+v", Alias(*e))
}

// FillFrom fills AccountEmailVerified from given interface.
func (e *AccountEmailVerified) FillFrom(from interface {
	GetEmail() (value string)
}) {
	e.Email = from.GetEmail()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*AccountEmailVerified) TypeID() uint32 {
	return AccountEmailVerifiedTypeID
}

// TypeName returns name of type in TL schema.
func (*AccountEmailVerified) TypeName() string {
	return "account.emailVerified"
}

// TypeInfo returns info about TL type.
func (e *AccountEmailVerified) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "account.emailVerified",
		ID:   AccountEmailVerifiedTypeID,
	}
	if e == nil {
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
func (e *AccountEmailVerified) Encode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode account.emailVerified#2b96cd1b as nil")
	}
	b.PutID(AccountEmailVerifiedTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *AccountEmailVerified) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode account.emailVerified#2b96cd1b as nil")
	}
	b.PutString(e.Email)
	return nil
}

// Decode implements bin.Decoder.
func (e *AccountEmailVerified) Decode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode account.emailVerified#2b96cd1b to nil")
	}
	if err := b.ConsumeID(AccountEmailVerifiedTypeID); err != nil {
		return fmt.Errorf("unable to decode account.emailVerified#2b96cd1b: %w", err)
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *AccountEmailVerified) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode account.emailVerified#2b96cd1b to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode account.emailVerified#2b96cd1b: field email: %w", err)
		}
		e.Email = value
	}
	return nil
}

// GetEmail returns value of Email field.
func (e *AccountEmailVerified) GetEmail() (value string) {
	if e == nil {
		return
	}
	return e.Email
}

// AccountEmailVerifiedLogin represents TL type `account.emailVerifiedLogin#e1bb0d61`.
// The email was verified correctly, and a login code was just sent to it.
//
// See https://core.telegram.org/constructor/account.emailVerifiedLogin for reference.
type AccountEmailVerifiedLogin struct {
	// The verified email address.
	Email string
	// Info about the sent login code¹
	//
	// Links:
	//  1) https://core.telegram.org/api/auth
	SentCode AuthSentCodeClass
}

// AccountEmailVerifiedLoginTypeID is TL type id of AccountEmailVerifiedLogin.
const AccountEmailVerifiedLoginTypeID = 0xe1bb0d61

// construct implements constructor of AccountEmailVerifiedClass.
func (e AccountEmailVerifiedLogin) construct() AccountEmailVerifiedClass { return &e }

// Ensuring interfaces in compile-time for AccountEmailVerifiedLogin.
var (
	_ bin.Encoder     = &AccountEmailVerifiedLogin{}
	_ bin.Decoder     = &AccountEmailVerifiedLogin{}
	_ bin.BareEncoder = &AccountEmailVerifiedLogin{}
	_ bin.BareDecoder = &AccountEmailVerifiedLogin{}

	_ AccountEmailVerifiedClass = &AccountEmailVerifiedLogin{}
)

func (e *AccountEmailVerifiedLogin) Zero() bool {
	if e == nil {
		return true
	}
	if !(e.Email == "") {
		return false
	}
	if !(e.SentCode == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (e *AccountEmailVerifiedLogin) String() string {
	if e == nil {
		return "AccountEmailVerifiedLogin(nil)"
	}
	type Alias AccountEmailVerifiedLogin
	return fmt.Sprintf("AccountEmailVerifiedLogin%+v", Alias(*e))
}

// FillFrom fills AccountEmailVerifiedLogin from given interface.
func (e *AccountEmailVerifiedLogin) FillFrom(from interface {
	GetEmail() (value string)
	GetSentCode() (value AuthSentCodeClass)
}) {
	e.Email = from.GetEmail()
	e.SentCode = from.GetSentCode()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*AccountEmailVerifiedLogin) TypeID() uint32 {
	return AccountEmailVerifiedLoginTypeID
}

// TypeName returns name of type in TL schema.
func (*AccountEmailVerifiedLogin) TypeName() string {
	return "account.emailVerifiedLogin"
}

// TypeInfo returns info about TL type.
func (e *AccountEmailVerifiedLogin) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "account.emailVerifiedLogin",
		ID:   AccountEmailVerifiedLoginTypeID,
	}
	if e == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Email",
			SchemaName: "email",
		},
		{
			Name:       "SentCode",
			SchemaName: "sent_code",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (e *AccountEmailVerifiedLogin) Encode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode account.emailVerifiedLogin#e1bb0d61 as nil")
	}
	b.PutID(AccountEmailVerifiedLoginTypeID)
	return e.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (e *AccountEmailVerifiedLogin) EncodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't encode account.emailVerifiedLogin#e1bb0d61 as nil")
	}
	b.PutString(e.Email)
	if e.SentCode == nil {
		return fmt.Errorf("unable to encode account.emailVerifiedLogin#e1bb0d61: field sent_code is nil")
	}
	if err := e.SentCode.Encode(b); err != nil {
		return fmt.Errorf("unable to encode account.emailVerifiedLogin#e1bb0d61: field sent_code: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (e *AccountEmailVerifiedLogin) Decode(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode account.emailVerifiedLogin#e1bb0d61 to nil")
	}
	if err := b.ConsumeID(AccountEmailVerifiedLoginTypeID); err != nil {
		return fmt.Errorf("unable to decode account.emailVerifiedLogin#e1bb0d61: %w", err)
	}
	return e.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (e *AccountEmailVerifiedLogin) DecodeBare(b *bin.Buffer) error {
	if e == nil {
		return fmt.Errorf("can't decode account.emailVerifiedLogin#e1bb0d61 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode account.emailVerifiedLogin#e1bb0d61: field email: %w", err)
		}
		e.Email = value
	}
	{
		value, err := DecodeAuthSentCode(b)
		if err != nil {
			return fmt.Errorf("unable to decode account.emailVerifiedLogin#e1bb0d61: field sent_code: %w", err)
		}
		e.SentCode = value
	}
	return nil
}

// GetEmail returns value of Email field.
func (e *AccountEmailVerifiedLogin) GetEmail() (value string) {
	if e == nil {
		return
	}
	return e.Email
}

// GetSentCode returns value of SentCode field.
func (e *AccountEmailVerifiedLogin) GetSentCode() (value AuthSentCodeClass) {
	if e == nil {
		return
	}
	return e.SentCode
}

// AccountEmailVerifiedClassName is schema name of AccountEmailVerifiedClass.
const AccountEmailVerifiedClassName = "account.EmailVerified"

// AccountEmailVerifiedClass represents account.EmailVerified generic type.
//
// See https://core.telegram.org/type/account.EmailVerified for reference.
//
// Constructors:
//   - [AccountEmailVerified]
//   - [AccountEmailVerifiedLogin]
//
// Example:
//
//	g, err := tg.DecodeAccountEmailVerified(buf)
//	if err != nil {
//	    panic(err)
//	}
//	switch v := g.(type) {
//	case *tg.AccountEmailVerified: // account.emailVerified#2b96cd1b
//	case *tg.AccountEmailVerifiedLogin: // account.emailVerifiedLogin#e1bb0d61
//	default: panic(v)
//	}
type AccountEmailVerifiedClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() AccountEmailVerifiedClass

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

	// The verified email address.
	GetEmail() (value string)
}

// DecodeAccountEmailVerified implements binary de-serialization for AccountEmailVerifiedClass.
func DecodeAccountEmailVerified(buf *bin.Buffer) (AccountEmailVerifiedClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case AccountEmailVerifiedTypeID:
		// Decoding account.emailVerified#2b96cd1b.
		v := AccountEmailVerified{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode AccountEmailVerifiedClass: %w", err)
		}
		return &v, nil
	case AccountEmailVerifiedLoginTypeID:
		// Decoding account.emailVerifiedLogin#e1bb0d61.
		v := AccountEmailVerifiedLogin{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode AccountEmailVerifiedClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode AccountEmailVerifiedClass: %w", bin.NewUnexpectedID(id))
	}
}

// AccountEmailVerified boxes the AccountEmailVerifiedClass providing a helper.
type AccountEmailVerifiedBox struct {
	EmailVerified AccountEmailVerifiedClass
}

// Decode implements bin.Decoder for AccountEmailVerifiedBox.
func (b *AccountEmailVerifiedBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode AccountEmailVerifiedBox to nil")
	}
	v, err := DecodeAccountEmailVerified(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.EmailVerified = v
	return nil
}

// Encode implements bin.Encode for AccountEmailVerifiedBox.
func (b *AccountEmailVerifiedBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.EmailVerified == nil {
		return fmt.Errorf("unable to encode AccountEmailVerifiedClass as nil")
	}
	return b.EmailVerified.Encode(buf)
}
