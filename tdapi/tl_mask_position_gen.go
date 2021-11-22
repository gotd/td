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

// MaskPosition represents TL type `maskPosition#82fbb63e`.
type MaskPosition struct {
	// Part of the face, relative to which the mask should be placed
	Point MaskPointClass
	// Shift by X-axis measured in widths of the mask scaled to the face size, from left to
	// right. (For example, -1.0 will place the mask just to the left of the default mask
	// position)
	XShift float64
	// Shift by Y-axis measured in heights of the mask scaled to the face size, from top to
	// bottom. (For example, 1.0 will place the mask just below the default mask position)
	YShift float64
	// Mask scaling coefficient. (For example, 2.0 means a doubled size)
	Scale float64
}

// MaskPositionTypeID is TL type id of MaskPosition.
const MaskPositionTypeID = 0x82fbb63e

// Ensuring interfaces in compile-time for MaskPosition.
var (
	_ bin.Encoder     = &MaskPosition{}
	_ bin.Decoder     = &MaskPosition{}
	_ bin.BareEncoder = &MaskPosition{}
	_ bin.BareDecoder = &MaskPosition{}
)

func (m *MaskPosition) Zero() bool {
	if m == nil {
		return true
	}
	if !(m.Point == nil) {
		return false
	}
	if !(m.XShift == 0) {
		return false
	}
	if !(m.YShift == 0) {
		return false
	}
	if !(m.Scale == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (m *MaskPosition) String() string {
	if m == nil {
		return "MaskPosition(nil)"
	}
	type Alias MaskPosition
	return fmt.Sprintf("MaskPosition%+v", Alias(*m))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MaskPosition) TypeID() uint32 {
	return MaskPositionTypeID
}

// TypeName returns name of type in TL schema.
func (*MaskPosition) TypeName() string {
	return "maskPosition"
}

// TypeInfo returns info about TL type.
func (m *MaskPosition) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "maskPosition",
		ID:   MaskPositionTypeID,
	}
	if m == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Point",
			SchemaName: "point",
		},
		{
			Name:       "XShift",
			SchemaName: "x_shift",
		},
		{
			Name:       "YShift",
			SchemaName: "y_shift",
		},
		{
			Name:       "Scale",
			SchemaName: "scale",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (m *MaskPosition) Encode(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't encode maskPosition#82fbb63e as nil")
	}
	b.PutID(MaskPositionTypeID)
	return m.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (m *MaskPosition) EncodeBare(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't encode maskPosition#82fbb63e as nil")
	}
	if m.Point == nil {
		return fmt.Errorf("unable to encode maskPosition#82fbb63e: field point is nil")
	}
	if err := m.Point.Encode(b); err != nil {
		return fmt.Errorf("unable to encode maskPosition#82fbb63e: field point: %w", err)
	}
	b.PutDouble(m.XShift)
	b.PutDouble(m.YShift)
	b.PutDouble(m.Scale)
	return nil
}

// Decode implements bin.Decoder.
func (m *MaskPosition) Decode(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't decode maskPosition#82fbb63e to nil")
	}
	if err := b.ConsumeID(MaskPositionTypeID); err != nil {
		return fmt.Errorf("unable to decode maskPosition#82fbb63e: %w", err)
	}
	return m.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (m *MaskPosition) DecodeBare(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't decode maskPosition#82fbb63e to nil")
	}
	{
		value, err := DecodeMaskPoint(b)
		if err != nil {
			return fmt.Errorf("unable to decode maskPosition#82fbb63e: field point: %w", err)
		}
		m.Point = value
	}
	{
		value, err := b.Double()
		if err != nil {
			return fmt.Errorf("unable to decode maskPosition#82fbb63e: field x_shift: %w", err)
		}
		m.XShift = value
	}
	{
		value, err := b.Double()
		if err != nil {
			return fmt.Errorf("unable to decode maskPosition#82fbb63e: field y_shift: %w", err)
		}
		m.YShift = value
	}
	{
		value, err := b.Double()
		if err != nil {
			return fmt.Errorf("unable to decode maskPosition#82fbb63e: field scale: %w", err)
		}
		m.Scale = value
	}
	return nil
}

// GetPoint returns value of Point field.
func (m *MaskPosition) GetPoint() (value MaskPointClass) {
	return m.Point
}

// GetXShift returns value of XShift field.
func (m *MaskPosition) GetXShift() (value float64) {
	return m.XShift
}

// GetYShift returns value of YShift field.
func (m *MaskPosition) GetYShift() (value float64) {
	return m.YShift
}

// GetScale returns value of Scale field.
func (m *MaskPosition) GetScale() (value float64) {
	return m.Scale
}