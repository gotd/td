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

// TestGetDifferenceRequest represents TL type `testGetDifference#68226325`.
type TestGetDifferenceRequest struct {
}

// TestGetDifferenceRequestTypeID is TL type id of TestGetDifferenceRequest.
const TestGetDifferenceRequestTypeID = 0x68226325

// Ensuring interfaces in compile-time for TestGetDifferenceRequest.
var (
	_ bin.Encoder     = &TestGetDifferenceRequest{}
	_ bin.Decoder     = &TestGetDifferenceRequest{}
	_ bin.BareEncoder = &TestGetDifferenceRequest{}
	_ bin.BareDecoder = &TestGetDifferenceRequest{}
)

func (t *TestGetDifferenceRequest) Zero() bool {
	if t == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (t *TestGetDifferenceRequest) String() string {
	if t == nil {
		return "TestGetDifferenceRequest(nil)"
	}
	type Alias TestGetDifferenceRequest
	return fmt.Sprintf("TestGetDifferenceRequest%+v", Alias(*t))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*TestGetDifferenceRequest) TypeID() uint32 {
	return TestGetDifferenceRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*TestGetDifferenceRequest) TypeName() string {
	return "testGetDifference"
}

// TypeInfo returns info about TL type.
func (t *TestGetDifferenceRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "testGetDifference",
		ID:   TestGetDifferenceRequestTypeID,
	}
	if t == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (t *TestGetDifferenceRequest) Encode(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't encode testGetDifference#68226325 as nil")
	}
	b.PutID(TestGetDifferenceRequestTypeID)
	return t.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (t *TestGetDifferenceRequest) EncodeBare(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't encode testGetDifference#68226325 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (t *TestGetDifferenceRequest) Decode(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't decode testGetDifference#68226325 to nil")
	}
	if err := b.ConsumeID(TestGetDifferenceRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode testGetDifference#68226325: %w", err)
	}
	return t.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (t *TestGetDifferenceRequest) DecodeBare(b *bin.Buffer) error {
	if t == nil {
		return fmt.Errorf("can't decode testGetDifference#68226325 to nil")
	}
	return nil
}

// TestGetDifference invokes method testGetDifference#68226325 returning error if any.
func (c *Client) TestGetDifference(ctx context.Context) error {
	var ok Ok

	request := &TestGetDifferenceRequest{}
	if err := c.rpc.Invoke(ctx, request, &ok); err != nil {
		return err
	}
	return nil
}