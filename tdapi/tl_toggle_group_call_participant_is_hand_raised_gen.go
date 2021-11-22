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

// ToggleGroupCallParticipantIsHandRaisedRequest represents TL type `toggleGroupCallParticipantIsHandRaised#8efb63e1`.
type ToggleGroupCallParticipantIsHandRaisedRequest struct {
	// Group call identifier
	GroupCallID int32
	// Participant identifier
	ParticipantID MessageSenderClass
	// Pass true if the user's hand should be raised. Only self hand can be raised. Requires
	// groupCall.can_be_managed group call flag to lower other's hand
	IsHandRaised bool
}

// ToggleGroupCallParticipantIsHandRaisedRequestTypeID is TL type id of ToggleGroupCallParticipantIsHandRaisedRequest.
const ToggleGroupCallParticipantIsHandRaisedRequestTypeID = 0x8efb63e1

// Ensuring interfaces in compile-time for ToggleGroupCallParticipantIsHandRaisedRequest.
var (
	_ bin.Encoder     = &ToggleGroupCallParticipantIsHandRaisedRequest{}
	_ bin.Decoder     = &ToggleGroupCallParticipantIsHandRaisedRequest{}
	_ bin.BareEncoder = &ToggleGroupCallParticipantIsHandRaisedRequest{}
	_ bin.BareDecoder = &ToggleGroupCallParticipantIsHandRaisedRequest{}
)

func (t *ToggleGroupCallParticipantIsHandRaisedRequest) Zero() bool {
	if t == nil {
		return true
	}
	if !(t.GroupCallID == 0) {
		return false
	}
	if !(t.ParticipantID == nil) {
		return false
	}
	if !(t.IsHandRaised == false) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) String() string {
	if t == nil {
		return "ToggleGroupCallParticipantIsHandRaisedRequest(nil)"
	}
	type Alias ToggleGroupCallParticipantIsHandRaisedRequest
	return fmt.Sprintf("ToggleGroupCallParticipantIsHandRaisedRequest%+v", Alias(*t))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ToggleGroupCallParticipantIsHandRaisedRequest) TypeID() uint32 {
	return ToggleGroupCallParticipantIsHandRaisedRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ToggleGroupCallParticipantIsHandRaisedRequest) TypeName() string {
	return "toggleGroupCallParticipantIsHandRaised"
}

// TypeInfo returns info about TL type.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "toggleGroupCallParticipantIsHandRaised",
		ID:   ToggleGroupCallParticipantIsHandRaisedRequestTypeID,
	}
	if t == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "GroupCallID",
			SchemaName: "group_call_id",
		},
		{
			Name:       "ParticipantID",
			SchemaName: "participant_id",
		},
		{
			Name:       "IsHandRaised",
			SchemaName: "is_hand_raised",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) Encode(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't encode toggleGroupCallParticipantIsHandRaised#8efb63e1 as nil")
	}
	b.PutID(ToggleGroupCallParticipantIsHandRaisedRequestTypeID)
	return t.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) EncodeBare(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't encode toggleGroupCallParticipantIsHandRaised#8efb63e1 as nil")
	}
	b.PutInt32(t.GroupCallID)
	if t.ParticipantID == nil {
		return fmt.Errorf("unable to encode toggleGroupCallParticipantIsHandRaised#8efb63e1: field participant_id is nil")
	}
	if err := t.ParticipantID.Encode(b); err != nil {
		return fmt.Errorf("unable to encode toggleGroupCallParticipantIsHandRaised#8efb63e1: field participant_id: %w", err)
	}
	b.PutBool(t.IsHandRaised)
	return nil
}

// Decode implements bin.Decoder.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) Decode(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't decode toggleGroupCallParticipantIsHandRaised#8efb63e1 to nil")
	}
	if err := b.ConsumeID(ToggleGroupCallParticipantIsHandRaisedRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode toggleGroupCallParticipantIsHandRaised#8efb63e1: %w", err)
	}
	return t.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) DecodeBare(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't decode toggleGroupCallParticipantIsHandRaised#8efb63e1 to nil")
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode toggleGroupCallParticipantIsHandRaised#8efb63e1: field group_call_id: %w", err)
		}
		t.GroupCallID = value
	}
	{
		value, err := DecodeMessageSender(b)
		if err != nil {
			return fmt.Errorf("unable to decode toggleGroupCallParticipantIsHandRaised#8efb63e1: field participant_id: %w", err)
		}
		t.ParticipantID = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode toggleGroupCallParticipantIsHandRaised#8efb63e1: field is_hand_raised: %w", err)
		}
		t.IsHandRaised = value
	}
	return nil
}

// GetGroupCallID returns value of GroupCallID field.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) GetGroupCallID() (value int32) {
	return t.GroupCallID
}

// GetParticipantID returns value of ParticipantID field.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) GetParticipantID() (value MessageSenderClass) {
	return t.ParticipantID
}

// GetIsHandRaised returns value of IsHandRaised field.
func (t *ToggleGroupCallParticipantIsHandRaisedRequest) GetIsHandRaised() (value bool) {
	return t.IsHandRaised
}

// ToggleGroupCallParticipantIsHandRaised invokes method toggleGroupCallParticipantIsHandRaised#8efb63e1 returning error if any.
func (c *Client) ToggleGroupCallParticipantIsHandRaised(ctx context.Context, request *ToggleGroupCallParticipantIsHandRaisedRequest) error {
	var ok Ok

	if err := c.rpc.Invoke(ctx, request, &ok); err != nil {
		return err
	}
	return nil
}