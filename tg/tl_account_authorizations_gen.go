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

// AccountAuthorizations represents TL type `account.authorizations#1250abde`.
// Logged-in sessions
//
// See https://core.telegram.org/constructor/account.authorizations for reference.
type AccountAuthorizations struct {
	// Logged-in sessions
	Authorizations []Authorization
}

// AccountAuthorizationsTypeID is TL type id of AccountAuthorizations.
const AccountAuthorizationsTypeID = 0x1250abde

func (a *AccountAuthorizations) Zero() bool {
	if a == nil {
		return true
	}
	if !(a.Authorizations == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (a *AccountAuthorizations) String() string {
	if a == nil {
		return "AccountAuthorizations(nil)"
	}
	type Alias AccountAuthorizations
	return fmt.Sprintf("AccountAuthorizations%+v", Alias(*a))
}

// FillFrom fills AccountAuthorizations from given interface.
func (a *AccountAuthorizations) FillFrom(from interface {
	GetAuthorizations() (value []Authorization)
}) {
	a.Authorizations = from.GetAuthorizations()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*AccountAuthorizations) TypeID() uint32 {
	return AccountAuthorizationsTypeID
}

// TypeName returns name of type in TL schema.
func (*AccountAuthorizations) TypeName() string {
	return "account.authorizations"
}

// TypeInfo returns info about TL type.
func (a *AccountAuthorizations) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "account.authorizations",
		ID:   AccountAuthorizationsTypeID,
	}
	if a == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Authorizations",
			SchemaName: "authorizations",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (a *AccountAuthorizations) Encode(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "account.authorizations#1250abde",
		}
	}
	b.PutID(AccountAuthorizationsTypeID)
	return a.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (a *AccountAuthorizations) EncodeBare(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "account.authorizations#1250abde",
		}
	}
	b.PutVectorHeader(len(a.Authorizations))
	for idx, v := range a.Authorizations {
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "account.authorizations#1250abde",
				FieldName: "authorizations",
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

// GetAuthorizations returns value of Authorizations field.
func (a *AccountAuthorizations) GetAuthorizations() (value []Authorization) {
	return a.Authorizations
}

// Decode implements bin.Decoder.
func (a *AccountAuthorizations) Decode(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "account.authorizations#1250abde",
		}
	}
	if err := b.ConsumeID(AccountAuthorizationsTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "account.authorizations#1250abde",
			Underlying: err,
		}
	}
	return a.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (a *AccountAuthorizations) DecodeBare(b *bin.Buffer) error {
	if a == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "account.authorizations#1250abde",
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "account.authorizations#1250abde",
				FieldName:  "authorizations",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			a.Authorizations = make([]Authorization, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value Authorization
			if err := value.Decode(b); err != nil {
				return &bin.FieldError{
					Action:     "decode",
					BareField:  false,
					TypeName:   "account.authorizations#1250abde",
					FieldName:  "authorizations",
					Underlying: err,
				}
			}
			a.Authorizations = append(a.Authorizations, value)
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for AccountAuthorizations.
var (
	_ bin.Encoder     = &AccountAuthorizations{}
	_ bin.Decoder     = &AccountAuthorizations{}
	_ bin.BareEncoder = &AccountAuthorizations{}
	_ bin.BareDecoder = &AccountAuthorizations{}
)
