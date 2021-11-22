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

// SetScopeNotificationSettingsRequest represents TL type `setScopeNotificationSettings#85cfb63a`.
type SetScopeNotificationSettingsRequest struct {
	// Types of chats for which to change the notification settings
	Scope NotificationSettingsScopeClass
	// The new notification settings for the given scope
	NotificationSettings ScopeNotificationSettings
}

// SetScopeNotificationSettingsRequestTypeID is TL type id of SetScopeNotificationSettingsRequest.
const SetScopeNotificationSettingsRequestTypeID = 0x85cfb63a

// Ensuring interfaces in compile-time for SetScopeNotificationSettingsRequest.
var (
	_ bin.Encoder     = &SetScopeNotificationSettingsRequest{}
	_ bin.Decoder     = &SetScopeNotificationSettingsRequest{}
	_ bin.BareEncoder = &SetScopeNotificationSettingsRequest{}
	_ bin.BareDecoder = &SetScopeNotificationSettingsRequest{}
)

func (s *SetScopeNotificationSettingsRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Scope == nil) {
		return false
	}
	if !(s.NotificationSettings.Zero()) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SetScopeNotificationSettingsRequest) String() string {
	if s == nil {
		return "SetScopeNotificationSettingsRequest(nil)"
	}
	type Alias SetScopeNotificationSettingsRequest
	return fmt.Sprintf("SetScopeNotificationSettingsRequest%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SetScopeNotificationSettingsRequest) TypeID() uint32 {
	return SetScopeNotificationSettingsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*SetScopeNotificationSettingsRequest) TypeName() string {
	return "setScopeNotificationSettings"
}

// TypeInfo returns info about TL type.
func (s *SetScopeNotificationSettingsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "setScopeNotificationSettings",
		ID:   SetScopeNotificationSettingsRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Scope",
			SchemaName: "scope",
		},
		{
			Name:       "NotificationSettings",
			SchemaName: "notification_settings",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *SetScopeNotificationSettingsRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode setScopeNotificationSettings#85cfb63a as nil")
	}
	b.PutID(SetScopeNotificationSettingsRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SetScopeNotificationSettingsRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode setScopeNotificationSettings#85cfb63a as nil")
	}
	if s.Scope == nil {
		return fmt.Errorf("unable to encode setScopeNotificationSettings#85cfb63a: field scope is nil")
	}
	if err := s.Scope.Encode(b); err != nil {
		return fmt.Errorf("unable to encode setScopeNotificationSettings#85cfb63a: field scope: %w", err)
	}
	if err := s.NotificationSettings.Encode(b); err != nil {
		return fmt.Errorf("unable to encode setScopeNotificationSettings#85cfb63a: field notification_settings: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *SetScopeNotificationSettingsRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode setScopeNotificationSettings#85cfb63a to nil")
	}
	if err := b.ConsumeID(SetScopeNotificationSettingsRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode setScopeNotificationSettings#85cfb63a: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SetScopeNotificationSettingsRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode setScopeNotificationSettings#85cfb63a to nil")
	}
	{
		value, err := DecodeNotificationSettingsScope(b)
		if err != nil {
			return fmt.Errorf("unable to decode setScopeNotificationSettings#85cfb63a: field scope: %w", err)
		}
		s.Scope = value
	}
	{
		if err := s.NotificationSettings.Decode(b); err != nil {
			return fmt.Errorf("unable to decode setScopeNotificationSettings#85cfb63a: field notification_settings: %w", err)
		}
	}
	return nil
}

// GetScope returns value of Scope field.
func (s *SetScopeNotificationSettingsRequest) GetScope() (value NotificationSettingsScopeClass) {
	return s.Scope
}

// GetNotificationSettings returns value of NotificationSettings field.
func (s *SetScopeNotificationSettingsRequest) GetNotificationSettings() (value ScopeNotificationSettings) {
	return s.NotificationSettings
}

// SetScopeNotificationSettings invokes method setScopeNotificationSettings#85cfb63a returning error if any.
func (c *Client) SetScopeNotificationSettings(ctx context.Context, request *SetScopeNotificationSettingsRequest) error {
	var ok Ok

	if err := c.rpc.Invoke(ctx, request, &ok); err != nil {
		return err
	}
	return nil
}