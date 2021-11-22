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

// PersonalDetails represents TL type `personalDetails#c0b869b7`.
type PersonalDetails struct {
	// First name of the user written in English; 1-255 characters
	FirstName string
	// Middle name of the user written in English; 0-255 characters
	MiddleName string
	// Last name of the user written in English; 1-255 characters
	LastName string
	// Native first name of the user; 1-255 characters
	NativeFirstName string
	// Native middle name of the user; 0-255 characters
	NativeMiddleName string
	// Native last name of the user; 1-255 characters
	NativeLastName string
	// Birthdate of the user
	Birthdate Date
	// Gender of the user, "male" or "female"
	Gender string
	// A two-letter ISO 3166-1 alpha-2 country code of the user's country
	CountryCode string
	// A two-letter ISO 3166-1 alpha-2 country code of the user's residence country
	ResidenceCountryCode string
}

// PersonalDetailsTypeID is TL type id of PersonalDetails.
const PersonalDetailsTypeID = 0xc0b869b7

// Ensuring interfaces in compile-time for PersonalDetails.
var (
	_ bin.Encoder     = &PersonalDetails{}
	_ bin.Decoder     = &PersonalDetails{}
	_ bin.BareEncoder = &PersonalDetails{}
	_ bin.BareDecoder = &PersonalDetails{}
)

func (p *PersonalDetails) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.FirstName == "") {
		return false
	}
	if !(p.MiddleName == "") {
		return false
	}
	if !(p.LastName == "") {
		return false
	}
	if !(p.NativeFirstName == "") {
		return false
	}
	if !(p.NativeMiddleName == "") {
		return false
	}
	if !(p.NativeLastName == "") {
		return false
	}
	if !(p.Birthdate.Zero()) {
		return false
	}
	if !(p.Gender == "") {
		return false
	}
	if !(p.CountryCode == "") {
		return false
	}
	if !(p.ResidenceCountryCode == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PersonalDetails) String() string {
	if p == nil {
		return "PersonalDetails(nil)"
	}
	type Alias PersonalDetails
	return fmt.Sprintf("PersonalDetails%+v", Alias(*p))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PersonalDetails) TypeID() uint32 {
	return PersonalDetailsTypeID
}

// TypeName returns name of type in TL schema.
func (*PersonalDetails) TypeName() string {
	return "personalDetails"
}

// TypeInfo returns info about TL type.
func (p *PersonalDetails) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "personalDetails",
		ID:   PersonalDetailsTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "FirstName",
			SchemaName: "first_name",
		},
		{
			Name:       "MiddleName",
			SchemaName: "middle_name",
		},
		{
			Name:       "LastName",
			SchemaName: "last_name",
		},
		{
			Name:       "NativeFirstName",
			SchemaName: "native_first_name",
		},
		{
			Name:       "NativeMiddleName",
			SchemaName: "native_middle_name",
		},
		{
			Name:       "NativeLastName",
			SchemaName: "native_last_name",
		},
		{
			Name:       "Birthdate",
			SchemaName: "birthdate",
		},
		{
			Name:       "Gender",
			SchemaName: "gender",
		},
		{
			Name:       "CountryCode",
			SchemaName: "country_code",
		},
		{
			Name:       "ResidenceCountryCode",
			SchemaName: "residence_country_code",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PersonalDetails) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode personalDetails#c0b869b7 as nil")
	}
	b.PutID(PersonalDetailsTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PersonalDetails) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode personalDetails#c0b869b7 as nil")
	}
	b.PutString(p.FirstName)
	b.PutString(p.MiddleName)
	b.PutString(p.LastName)
	b.PutString(p.NativeFirstName)
	b.PutString(p.NativeMiddleName)
	b.PutString(p.NativeLastName)
	if err := p.Birthdate.Encode(b); err != nil {
		return fmt.Errorf("unable to encode personalDetails#c0b869b7: field birthdate: %w", err)
	}
	b.PutString(p.Gender)
	b.PutString(p.CountryCode)
	b.PutString(p.ResidenceCountryCode)
	return nil
}

// Decode implements bin.Decoder.
func (p *PersonalDetails) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode personalDetails#c0b869b7 to nil")
	}
	if err := b.ConsumeID(PersonalDetailsTypeID); err != nil {
		return fmt.Errorf("unable to decode personalDetails#c0b869b7: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PersonalDetails) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode personalDetails#c0b869b7 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field first_name: %w", err)
		}
		p.FirstName = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field middle_name: %w", err)
		}
		p.MiddleName = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field last_name: %w", err)
		}
		p.LastName = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field native_first_name: %w", err)
		}
		p.NativeFirstName = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field native_middle_name: %w", err)
		}
		p.NativeMiddleName = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field native_last_name: %w", err)
		}
		p.NativeLastName = value
	}
	{
		if err := p.Birthdate.Decode(b); err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field birthdate: %w", err)
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field gender: %w", err)
		}
		p.Gender = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field country_code: %w", err)
		}
		p.CountryCode = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode personalDetails#c0b869b7: field residence_country_code: %w", err)
		}
		p.ResidenceCountryCode = value
	}
	return nil
}

// GetFirstName returns value of FirstName field.
func (p *PersonalDetails) GetFirstName() (value string) {
	return p.FirstName
}

// GetMiddleName returns value of MiddleName field.
func (p *PersonalDetails) GetMiddleName() (value string) {
	return p.MiddleName
}

// GetLastName returns value of LastName field.
func (p *PersonalDetails) GetLastName() (value string) {
	return p.LastName
}

// GetNativeFirstName returns value of NativeFirstName field.
func (p *PersonalDetails) GetNativeFirstName() (value string) {
	return p.NativeFirstName
}

// GetNativeMiddleName returns value of NativeMiddleName field.
func (p *PersonalDetails) GetNativeMiddleName() (value string) {
	return p.NativeMiddleName
}

// GetNativeLastName returns value of NativeLastName field.
func (p *PersonalDetails) GetNativeLastName() (value string) {
	return p.NativeLastName
}

// GetBirthdate returns value of Birthdate field.
func (p *PersonalDetails) GetBirthdate() (value Date) {
	return p.Birthdate
}

// GetGender returns value of Gender field.
func (p *PersonalDetails) GetGender() (value string) {
	return p.Gender
}

// GetCountryCode returns value of CountryCode field.
func (p *PersonalDetails) GetCountryCode() (value string) {
	return p.CountryCode
}

// GetResidenceCountryCode returns value of ResidenceCountryCode field.
func (p *PersonalDetails) GetResidenceCountryCode() (value string) {
	return p.ResidenceCountryCode
}