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

// GeoPointEmpty represents TL type `geoPointEmpty#1117dd5f`.
// Empty constructor.
//
// See https://core.telegram.org/constructor/geoPointEmpty for reference.
type GeoPointEmpty struct {
}

// GeoPointEmptyTypeID is TL type id of GeoPointEmpty.
const GeoPointEmptyTypeID = 0x1117dd5f

func (g *GeoPointEmpty) Zero() bool {
	if g == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (g *GeoPointEmpty) String() string {
	if g == nil {
		return "GeoPointEmpty(nil)"
	}
	type Alias GeoPointEmpty
	return fmt.Sprintf("GeoPointEmpty%+v", Alias(*g))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*GeoPointEmpty) TypeID() uint32 {
	return GeoPointEmptyTypeID
}

// TypeName returns name of type in TL schema.
func (*GeoPointEmpty) TypeName() string {
	return "geoPointEmpty"
}

// TypeInfo returns info about TL type.
func (g *GeoPointEmpty) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "geoPointEmpty",
		ID:   GeoPointEmptyTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (g *GeoPointEmpty) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode geoPointEmpty#1117dd5f as nil")
	}
	b.PutID(GeoPointEmptyTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *GeoPointEmpty) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode geoPointEmpty#1117dd5f as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *GeoPointEmpty) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode geoPointEmpty#1117dd5f to nil")
	}
	if err := b.ConsumeID(GeoPointEmptyTypeID); err != nil {
		return fmt.Errorf("unable to decode geoPointEmpty#1117dd5f: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *GeoPointEmpty) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode geoPointEmpty#1117dd5f to nil")
	}
	return nil
}

// construct implements constructor of GeoPointClass.
func (g GeoPointEmpty) construct() GeoPointClass { return &g }

// Ensuring interfaces in compile-time for GeoPointEmpty.
var (
	_ bin.Encoder     = &GeoPointEmpty{}
	_ bin.Decoder     = &GeoPointEmpty{}
	_ bin.BareEncoder = &GeoPointEmpty{}
	_ bin.BareDecoder = &GeoPointEmpty{}

	_ GeoPointClass = &GeoPointEmpty{}
)

// GeoPoint represents TL type `geoPoint#b2a2f663`.
// GeoPoint.
//
// See https://core.telegram.org/constructor/geoPoint for reference.
type GeoPoint struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Longtitude
	Long float64
	// Latitude
	Lat float64
	// Access hash
	AccessHash int64
	// The estimated horizontal accuracy of the location, in meters; as defined by the sender.
	//
	// Use SetAccuracyRadius and GetAccuracyRadius helpers.
	AccuracyRadius int
}

// GeoPointTypeID is TL type id of GeoPoint.
const GeoPointTypeID = 0xb2a2f663

func (g *GeoPoint) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Flags.Zero()) {
		return false
	}
	if !(g.Long == 0) {
		return false
	}
	if !(g.Lat == 0) {
		return false
	}
	if !(g.AccessHash == 0) {
		return false
	}
	if !(g.AccuracyRadius == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *GeoPoint) String() string {
	if g == nil {
		return "GeoPoint(nil)"
	}
	type Alias GeoPoint
	return fmt.Sprintf("GeoPoint%+v", Alias(*g))
}

// FillFrom fills GeoPoint from given interface.
func (g *GeoPoint) FillFrom(from interface {
	GetLong() (value float64)
	GetLat() (value float64)
	GetAccessHash() (value int64)
	GetAccuracyRadius() (value int, ok bool)
}) {
	g.Long = from.GetLong()
	g.Lat = from.GetLat()
	g.AccessHash = from.GetAccessHash()
	if val, ok := from.GetAccuracyRadius(); ok {
		g.AccuracyRadius = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*GeoPoint) TypeID() uint32 {
	return GeoPointTypeID
}

// TypeName returns name of type in TL schema.
func (*GeoPoint) TypeName() string {
	return "geoPoint"
}

// TypeInfo returns info about TL type.
func (g *GeoPoint) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "geoPoint",
		ID:   GeoPointTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Long",
			SchemaName: "long",
		},
		{
			Name:       "Lat",
			SchemaName: "lat",
		},
		{
			Name:       "AccessHash",
			SchemaName: "access_hash",
		},
		{
			Name:       "AccuracyRadius",
			SchemaName: "accuracy_radius",
			Null:       !g.Flags.Has(0),
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *GeoPoint) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode geoPoint#b2a2f663 as nil")
	}
	b.PutID(GeoPointTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *GeoPoint) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode geoPoint#b2a2f663 as nil")
	}
	if !(g.AccuracyRadius == 0) {
		g.Flags.Set(0)
	}
	if err := g.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode geoPoint#b2a2f663: field flags: %w", err)
	}
	b.PutDouble(g.Long)
	b.PutDouble(g.Lat)
	b.PutLong(g.AccessHash)
	if g.Flags.Has(0) {
		b.PutInt(g.AccuracyRadius)
	}
	return nil
}

// GetLong returns value of Long field.
func (g *GeoPoint) GetLong() (value float64) {
	return g.Long
}

// GetLat returns value of Lat field.
func (g *GeoPoint) GetLat() (value float64) {
	return g.Lat
}

// GetAccessHash returns value of AccessHash field.
func (g *GeoPoint) GetAccessHash() (value int64) {
	return g.AccessHash
}

// SetAccuracyRadius sets value of AccuracyRadius conditional field.
func (g *GeoPoint) SetAccuracyRadius(value int) {
	g.Flags.Set(0)
	g.AccuracyRadius = value
}

// GetAccuracyRadius returns value of AccuracyRadius conditional field and
// boolean which is true if field was set.
func (g *GeoPoint) GetAccuracyRadius() (value int, ok bool) {
	if !g.Flags.Has(0) {
		return value, false
	}
	return g.AccuracyRadius, true
}

// Decode implements bin.Decoder.
func (g *GeoPoint) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode geoPoint#b2a2f663 to nil")
	}
	if err := b.ConsumeID(GeoPointTypeID); err != nil {
		return fmt.Errorf("unable to decode geoPoint#b2a2f663: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *GeoPoint) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode geoPoint#b2a2f663 to nil")
	}
	{
		if err := g.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode geoPoint#b2a2f663: field flags: %w", err)
		}
	}
	{
		value, err := b.Double()
		if err != nil {
			return fmt.Errorf("unable to decode geoPoint#b2a2f663: field long: %w", err)
		}
		g.Long = value
	}
	{
		value, err := b.Double()
		if err != nil {
			return fmt.Errorf("unable to decode geoPoint#b2a2f663: field lat: %w", err)
		}
		g.Lat = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode geoPoint#b2a2f663: field access_hash: %w", err)
		}
		g.AccessHash = value
	}
	if g.Flags.Has(0) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode geoPoint#b2a2f663: field accuracy_radius: %w", err)
		}
		g.AccuracyRadius = value
	}
	return nil
}

// construct implements constructor of GeoPointClass.
func (g GeoPoint) construct() GeoPointClass { return &g }

// Ensuring interfaces in compile-time for GeoPoint.
var (
	_ bin.Encoder     = &GeoPoint{}
	_ bin.Decoder     = &GeoPoint{}
	_ bin.BareEncoder = &GeoPoint{}
	_ bin.BareDecoder = &GeoPoint{}

	_ GeoPointClass = &GeoPoint{}
)

// GeoPointClass represents GeoPoint generic type.
//
// See https://core.telegram.org/type/GeoPoint for reference.
//
// Example:
//  g, err := tg.DecodeGeoPoint(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.GeoPointEmpty: // geoPointEmpty#1117dd5f
//  case *tg.GeoPoint: // geoPoint#b2a2f663
//  default: panic(v)
//  }
type GeoPointClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() GeoPointClass

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

	// AsNotEmpty tries to map GeoPointClass to GeoPoint.
	AsNotEmpty() (*GeoPoint, bool)
}

// AsNotEmpty tries to map GeoPointEmpty to GeoPoint.
func (g *GeoPointEmpty) AsNotEmpty() (*GeoPoint, bool) {
	return nil, false
}

// AsNotEmpty tries to map GeoPoint to GeoPoint.
func (g *GeoPoint) AsNotEmpty() (*GeoPoint, bool) {
	return g, true
}

// DecodeGeoPoint implements binary de-serialization for GeoPointClass.
func DecodeGeoPoint(buf *bin.Buffer) (GeoPointClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case GeoPointEmptyTypeID:
		// Decoding geoPointEmpty#1117dd5f.
		v := GeoPointEmpty{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode GeoPointClass: %w", err)
		}
		return &v, nil
	case GeoPointTypeID:
		// Decoding geoPoint#b2a2f663.
		v := GeoPoint{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode GeoPointClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode GeoPointClass: %w", bin.NewUnexpectedID(id))
	}
}

// GeoPoint boxes the GeoPointClass providing a helper.
type GeoPointBox struct {
	GeoPoint GeoPointClass
}

// Decode implements bin.Decoder for GeoPointBox.
func (b *GeoPointBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode GeoPointBox to nil")
	}
	v, err := DecodeGeoPoint(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.GeoPoint = v
	return nil
}

// Encode implements bin.Encode for GeoPointBox.
func (b *GeoPointBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.GeoPoint == nil {
		return fmt.Errorf("unable to encode GeoPointClass as nil")
	}
	return b.GeoPoint.Encode(buf)
}

// GeoPointClassArray is adapter for slice of GeoPointClass.
type GeoPointClassArray []GeoPointClass

// Sort sorts slice of GeoPointClass.
func (s GeoPointClassArray) Sort(less func(a, b GeoPointClass) bool) GeoPointClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of GeoPointClass.
func (s GeoPointClassArray) SortStable(less func(a, b GeoPointClass) bool) GeoPointClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of GeoPointClass.
func (s GeoPointClassArray) Retain(keep func(x GeoPointClass) bool) GeoPointClassArray {
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
func (s GeoPointClassArray) First() (v GeoPointClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s GeoPointClassArray) Last() (v GeoPointClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *GeoPointClassArray) PopFirst() (v GeoPointClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero GeoPointClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *GeoPointClassArray) Pop() (v GeoPointClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsGeoPoint returns copy with only GeoPoint constructors.
func (s GeoPointClassArray) AsGeoPoint() (to GeoPointArray) {
	for _, elem := range s {
		value, ok := elem.(*GeoPoint)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AppendOnlyNotEmpty appends only NotEmpty constructors to
// given slice.
func (s GeoPointClassArray) AppendOnlyNotEmpty(to []*GeoPoint) []*GeoPoint {
	for _, elem := range s {
		value, ok := elem.AsNotEmpty()
		if !ok {
			continue
		}
		to = append(to, value)
	}

	return to
}

// AsNotEmpty returns copy with only NotEmpty constructors.
func (s GeoPointClassArray) AsNotEmpty() (to []*GeoPoint) {
	return s.AppendOnlyNotEmpty(to)
}

// FirstAsNotEmpty returns first element of slice (if exists).
func (s GeoPointClassArray) FirstAsNotEmpty() (v *GeoPoint, ok bool) {
	value, ok := s.First()
	if !ok {
		return
	}
	return value.AsNotEmpty()
}

// LastAsNotEmpty returns last element of slice (if exists).
func (s GeoPointClassArray) LastAsNotEmpty() (v *GeoPoint, ok bool) {
	value, ok := s.Last()
	if !ok {
		return
	}
	return value.AsNotEmpty()
}

// PopFirstAsNotEmpty returns element of slice (if exists).
func (s *GeoPointClassArray) PopFirstAsNotEmpty() (v *GeoPoint, ok bool) {
	value, ok := s.PopFirst()
	if !ok {
		return
	}
	return value.AsNotEmpty()
}

// PopAsNotEmpty returns element of slice (if exists).
func (s *GeoPointClassArray) PopAsNotEmpty() (v *GeoPoint, ok bool) {
	value, ok := s.Pop()
	if !ok {
		return
	}
	return value.AsNotEmpty()
}

// GeoPointArray is adapter for slice of GeoPoint.
type GeoPointArray []GeoPoint

// Sort sorts slice of GeoPoint.
func (s GeoPointArray) Sort(less func(a, b GeoPoint) bool) GeoPointArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of GeoPoint.
func (s GeoPointArray) SortStable(less func(a, b GeoPoint) bool) GeoPointArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of GeoPoint.
func (s GeoPointArray) Retain(keep func(x GeoPoint) bool) GeoPointArray {
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
func (s GeoPointArray) First() (v GeoPoint, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s GeoPointArray) Last() (v GeoPoint, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *GeoPointArray) PopFirst() (v GeoPoint, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero GeoPoint
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *GeoPointArray) Pop() (v GeoPoint, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
