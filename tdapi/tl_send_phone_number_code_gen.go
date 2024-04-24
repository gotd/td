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

// SendPhoneNumberCodeRequest represents TL type `sendPhoneNumberCode#409e3d10`.
type SendPhoneNumberCodeRequest struct {
	// The phone number, in international format
	PhoneNumber string
	// Settings for the authentication of the user's phone number; pass null to use default
	// settings
	Settings PhoneNumberAuthenticationSettings
	// Type of the request for which the code is sent
	Type PhoneNumberCodeTypeClass
}

// SendPhoneNumberCodeRequestTypeID is TL type id of SendPhoneNumberCodeRequest.
const SendPhoneNumberCodeRequestTypeID = 0x409e3d10

// Ensuring interfaces in compile-time for SendPhoneNumberCodeRequest.
var (
	_ bin.Encoder     = &SendPhoneNumberCodeRequest{}
	_ bin.Decoder     = &SendPhoneNumberCodeRequest{}
	_ bin.BareEncoder = &SendPhoneNumberCodeRequest{}
	_ bin.BareDecoder = &SendPhoneNumberCodeRequest{}
)

func (s *SendPhoneNumberCodeRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.PhoneNumber == "") {
		return false
	}
	if !(s.Settings.Zero()) {
		return false
	}
	if !(s.Type == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SendPhoneNumberCodeRequest) String() string {
	if s == nil {
		return "SendPhoneNumberCodeRequest(nil)"
	}
	type Alias SendPhoneNumberCodeRequest
	return fmt.Sprintf("SendPhoneNumberCodeRequest%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SendPhoneNumberCodeRequest) TypeID() uint32 {
	return SendPhoneNumberCodeRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*SendPhoneNumberCodeRequest) TypeName() string {
	return "sendPhoneNumberCode"
}

// TypeInfo returns info about TL type.
func (s *SendPhoneNumberCodeRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "sendPhoneNumberCode",
		ID:   SendPhoneNumberCodeRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "PhoneNumber",
			SchemaName: "phone_number",
		},
		{
			Name:       "Settings",
			SchemaName: "settings",
		},
		{
			Name:       "Type",
			SchemaName: "type",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SendPhoneNumberCodeRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode sendPhoneNumberCode#409e3d10 as nil")
	}
	b.PutID(SendPhoneNumberCodeRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SendPhoneNumberCodeRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode sendPhoneNumberCode#409e3d10 as nil")
	}
	b.PutString(s.PhoneNumber)
	if err := s.Settings.Encode(b); err != nil {
		return fmt.Errorf("unable to encode sendPhoneNumberCode#409e3d10: field settings: %w", err)
	}
	if s.Type == nil {
		return fmt.Errorf("unable to encode sendPhoneNumberCode#409e3d10: field type is nil")
	}
	if err := s.Type.Encode(b); err != nil {
		return fmt.Errorf("unable to encode sendPhoneNumberCode#409e3d10: field type: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *SendPhoneNumberCodeRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode sendPhoneNumberCode#409e3d10 to nil")
	}
	if err := b.ConsumeID(SendPhoneNumberCodeRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SendPhoneNumberCodeRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode sendPhoneNumberCode#409e3d10 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: field phone_number: %w", err)
		}
		s.PhoneNumber = value
	}
	{
		if err := s.Settings.Decode(b); err != nil {
			return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: field settings: %w", err)
		}
	}
	{
		value, err := DecodePhoneNumberCodeType(b)
		if err != nil {
			return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: field type: %w", err)
		}
		s.Type = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (s *SendPhoneNumberCodeRequest) EncodeTDLibJSON(b tdjson.Encoder) error {
	if s == nil {
		return fmt.Errorf("can't encode sendPhoneNumberCode#409e3d10 as nil")
	}
	b.ObjStart()
	b.PutID("sendPhoneNumberCode")
	b.Comma()
	b.FieldStart("phone_number")
	b.PutString(s.PhoneNumber)
	b.Comma()
	b.FieldStart("settings")
	if err := s.Settings.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode sendPhoneNumberCode#409e3d10: field settings: %w", err)
	}
	b.Comma()
	b.FieldStart("type")
	if s.Type == nil {
		return fmt.Errorf("unable to encode sendPhoneNumberCode#409e3d10: field type is nil")
	}
	if err := s.Type.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode sendPhoneNumberCode#409e3d10: field type: %w", err)
	}
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (s *SendPhoneNumberCodeRequest) DecodeTDLibJSON(b tdjson.Decoder) error {
	if s == nil {
		return fmt.Errorf("can't decode sendPhoneNumberCode#409e3d10 to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("sendPhoneNumberCode"); err != nil {
				return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: %w", err)
			}
		case "phone_number":
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: field phone_number: %w", err)
			}
			s.PhoneNumber = value
		case "settings":
			if err := s.Settings.DecodeTDLibJSON(b); err != nil {
				return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: field settings: %w", err)
			}
		case "type":
			value, err := DecodeTDLibJSONPhoneNumberCodeType(b)
			if err != nil {
				return fmt.Errorf("unable to decode sendPhoneNumberCode#409e3d10: field type: %w", err)
			}
			s.Type = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetPhoneNumber returns value of PhoneNumber field.
func (s *SendPhoneNumberCodeRequest) GetPhoneNumber() (value string) {
	if s == nil {
		return
	}
	return s.PhoneNumber
}

// GetSettings returns value of Settings field.
func (s *SendPhoneNumberCodeRequest) GetSettings() (value PhoneNumberAuthenticationSettings) {
	if s == nil {
		return
	}
	return s.Settings
}

// GetType returns value of Type field.
func (s *SendPhoneNumberCodeRequest) GetType() (value PhoneNumberCodeTypeClass) {
	if s == nil {
		return
	}
	return s.Type
}

// SendPhoneNumberCode invokes method sendPhoneNumberCode#409e3d10 returning error if any.
func (c *Client) SendPhoneNumberCode(ctx context.Context, request *SendPhoneNumberCodeRequest) (*AuthenticationCodeInfo, error) {
	var result AuthenticationCodeInfo

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}