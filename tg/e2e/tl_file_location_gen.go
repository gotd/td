// Code generated by gotdgen, DO NOT EDIT.

package e2e

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

// FileLocationUnavailable represents TL type `fileLocationUnavailable#7c596b46`.
//
// See https://core.telegram.org/constructor/fileLocationUnavailable for reference.
type FileLocationUnavailable struct {
	// VolumeID field of FileLocationUnavailable.
	VolumeID int64
	// LocalID field of FileLocationUnavailable.
	LocalID int
	// Secret field of FileLocationUnavailable.
	Secret int64
}

// FileLocationUnavailableTypeID is TL type id of FileLocationUnavailable.
const FileLocationUnavailableTypeID = 0x7c596b46

func (f *FileLocationUnavailable) Zero() bool {
	if f == nil {
		return true
	}
	if !(f.VolumeID == 0) {
		return false
	}
	if !(f.LocalID == 0) {
		return false
	}
	if !(f.Secret == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (f *FileLocationUnavailable) String() string {
	if f == nil {
		return "FileLocationUnavailable(nil)"
	}
	type Alias FileLocationUnavailable
	return fmt.Sprintf("FileLocationUnavailable%+v", Alias(*f))
}

// FillFrom fills FileLocationUnavailable from given interface.
func (f *FileLocationUnavailable) FillFrom(from interface {
	GetVolumeID() (value int64)
	GetLocalID() (value int)
	GetSecret() (value int64)
}) {
	f.VolumeID = from.GetVolumeID()
	f.LocalID = from.GetLocalID()
	f.Secret = from.GetSecret()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*FileLocationUnavailable) TypeID() uint32 {
	return FileLocationUnavailableTypeID
}

// TypeName returns name of type in TL schema.
func (*FileLocationUnavailable) TypeName() string {
	return "fileLocationUnavailable"
}

// TypeInfo returns info about TL type.
func (f *FileLocationUnavailable) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "fileLocationUnavailable",
		ID:   FileLocationUnavailableTypeID,
	}
	if f == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "VolumeID",
			SchemaName: "volume_id",
		},
		{
			Name:       "LocalID",
			SchemaName: "local_id",
		},
		{
			Name:       "Secret",
			SchemaName: "secret",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (f *FileLocationUnavailable) Encode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode fileLocationUnavailable#7c596b46 as nil")
	}
	b.PutID(FileLocationUnavailableTypeID)
	return f.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (f *FileLocationUnavailable) EncodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode fileLocationUnavailable#7c596b46 as nil")
	}
	b.PutLong(f.VolumeID)
	b.PutInt(f.LocalID)
	b.PutLong(f.Secret)
	return nil
}

// GetVolumeID returns value of VolumeID field.
func (f *FileLocationUnavailable) GetVolumeID() (value int64) {
	return f.VolumeID
}

// GetLocalID returns value of LocalID field.
func (f *FileLocationUnavailable) GetLocalID() (value int) {
	return f.LocalID
}

// GetSecret returns value of Secret field.
func (f *FileLocationUnavailable) GetSecret() (value int64) {
	return f.Secret
}

// Decode implements bin.Decoder.
func (f *FileLocationUnavailable) Decode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode fileLocationUnavailable#7c596b46 to nil")
	}
	if err := b.ConsumeID(FileLocationUnavailableTypeID); err != nil {
		return fmt.Errorf("unable to decode fileLocationUnavailable#7c596b46: %w", err)
	}
	return f.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (f *FileLocationUnavailable) DecodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode fileLocationUnavailable#7c596b46 to nil")
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode fileLocationUnavailable#7c596b46: field volume_id: %w", err)
		}
		f.VolumeID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode fileLocationUnavailable#7c596b46: field local_id: %w", err)
		}
		f.LocalID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode fileLocationUnavailable#7c596b46: field secret: %w", err)
		}
		f.Secret = value
	}
	return nil
}

// construct implements constructor of FileLocationClass.
func (f FileLocationUnavailable) construct() FileLocationClass { return &f }

// Ensuring interfaces in compile-time for FileLocationUnavailable.
var (
	_ bin.Encoder     = &FileLocationUnavailable{}
	_ bin.Decoder     = &FileLocationUnavailable{}
	_ bin.BareEncoder = &FileLocationUnavailable{}
	_ bin.BareDecoder = &FileLocationUnavailable{}

	_ FileLocationClass = &FileLocationUnavailable{}
)

// FileLocation represents TL type `fileLocation#53d69076`.
//
// See https://core.telegram.org/constructor/fileLocation for reference.
type FileLocation struct {
	// DCID field of FileLocation.
	DCID int
	// VolumeID field of FileLocation.
	VolumeID int64
	// LocalID field of FileLocation.
	LocalID int
	// Secret field of FileLocation.
	Secret int64
}

// FileLocationTypeID is TL type id of FileLocation.
const FileLocationTypeID = 0x53d69076

func (f *FileLocation) Zero() bool {
	if f == nil {
		return true
	}
	if !(f.DCID == 0) {
		return false
	}
	if !(f.VolumeID == 0) {
		return false
	}
	if !(f.LocalID == 0) {
		return false
	}
	if !(f.Secret == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (f *FileLocation) String() string {
	if f == nil {
		return "FileLocation(nil)"
	}
	type Alias FileLocation
	return fmt.Sprintf("FileLocation%+v", Alias(*f))
}

// FillFrom fills FileLocation from given interface.
func (f *FileLocation) FillFrom(from interface {
	GetDCID() (value int)
	GetVolumeID() (value int64)
	GetLocalID() (value int)
	GetSecret() (value int64)
}) {
	f.DCID = from.GetDCID()
	f.VolumeID = from.GetVolumeID()
	f.LocalID = from.GetLocalID()
	f.Secret = from.GetSecret()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*FileLocation) TypeID() uint32 {
	return FileLocationTypeID
}

// TypeName returns name of type in TL schema.
func (*FileLocation) TypeName() string {
	return "fileLocation"
}

// TypeInfo returns info about TL type.
func (f *FileLocation) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "fileLocation",
		ID:   FileLocationTypeID,
	}
	if f == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "DCID",
			SchemaName: "dc_id",
		},
		{
			Name:       "VolumeID",
			SchemaName: "volume_id",
		},
		{
			Name:       "LocalID",
			SchemaName: "local_id",
		},
		{
			Name:       "Secret",
			SchemaName: "secret",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (f *FileLocation) Encode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode fileLocation#53d69076 as nil")
	}
	b.PutID(FileLocationTypeID)
	return f.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (f *FileLocation) EncodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't encode fileLocation#53d69076 as nil")
	}
	b.PutInt(f.DCID)
	b.PutLong(f.VolumeID)
	b.PutInt(f.LocalID)
	b.PutLong(f.Secret)
	return nil
}

// GetDCID returns value of DCID field.
func (f *FileLocation) GetDCID() (value int) {
	return f.DCID
}

// GetVolumeID returns value of VolumeID field.
func (f *FileLocation) GetVolumeID() (value int64) {
	return f.VolumeID
}

// GetLocalID returns value of LocalID field.
func (f *FileLocation) GetLocalID() (value int) {
	return f.LocalID
}

// GetSecret returns value of Secret field.
func (f *FileLocation) GetSecret() (value int64) {
	return f.Secret
}

// Decode implements bin.Decoder.
func (f *FileLocation) Decode(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode fileLocation#53d69076 to nil")
	}
	if err := b.ConsumeID(FileLocationTypeID); err != nil {
		return fmt.Errorf("unable to decode fileLocation#53d69076: %w", err)
	}
	return f.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (f *FileLocation) DecodeBare(b *bin.Buffer) error {
	if f == nil {
		return fmt.Errorf("can't decode fileLocation#53d69076 to nil")
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode fileLocation#53d69076: field dc_id: %w", err)
		}
		f.DCID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode fileLocation#53d69076: field volume_id: %w", err)
		}
		f.VolumeID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode fileLocation#53d69076: field local_id: %w", err)
		}
		f.LocalID = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode fileLocation#53d69076: field secret: %w", err)
		}
		f.Secret = value
	}
	return nil
}

// construct implements constructor of FileLocationClass.
func (f FileLocation) construct() FileLocationClass { return &f }

// Ensuring interfaces in compile-time for FileLocation.
var (
	_ bin.Encoder     = &FileLocation{}
	_ bin.Decoder     = &FileLocation{}
	_ bin.BareEncoder = &FileLocation{}
	_ bin.BareDecoder = &FileLocation{}

	_ FileLocationClass = &FileLocation{}
)

// FileLocationClass represents FileLocation generic type.
//
// See https://core.telegram.org/type/FileLocation for reference.
//
// Example:
//  g, err := e2e.DecodeFileLocation(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *e2e.FileLocationUnavailable: // fileLocationUnavailable#7c596b46
//  case *e2e.FileLocation: // fileLocation#53d69076
//  default: panic(v)
//  }
type FileLocationClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() FileLocationClass

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

	// VolumeID field of FileLocationUnavailable.
	GetVolumeID() (value int64)

	// LocalID field of FileLocationUnavailable.
	GetLocalID() (value int)

	// Secret field of FileLocationUnavailable.
	GetSecret() (value int64)

	// AsAvailable tries to map FileLocationClass to FileLocation.
	AsAvailable() (*FileLocation, bool)
}

// AsAvailable tries to map FileLocationUnavailable to FileLocation.
func (f *FileLocationUnavailable) AsAvailable() (*FileLocation, bool) {
	return nil, false
}

// AsAvailable tries to map FileLocation to FileLocation.
func (f *FileLocation) AsAvailable() (*FileLocation, bool) {
	return f, true
}

// DecodeFileLocation implements binary de-serialization for FileLocationClass.
func DecodeFileLocation(buf *bin.Buffer) (FileLocationClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case FileLocationUnavailableTypeID:
		// Decoding fileLocationUnavailable#7c596b46.
		v := FileLocationUnavailable{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode FileLocationClass: %w", err)
		}
		return &v, nil
	case FileLocationTypeID:
		// Decoding fileLocation#53d69076.
		v := FileLocation{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode FileLocationClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode FileLocationClass: %w", bin.NewUnexpectedID(id))
	}
}

// FileLocation boxes the FileLocationClass providing a helper.
type FileLocationBox struct {
	FileLocation FileLocationClass
}

// Decode implements bin.Decoder for FileLocationBox.
func (b *FileLocationBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode FileLocationBox to nil")
	}
	v, err := DecodeFileLocation(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.FileLocation = v
	return nil
}

// Encode implements bin.Encode for FileLocationBox.
func (b *FileLocationBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.FileLocation == nil {
		return fmt.Errorf("unable to encode FileLocationClass as nil")
	}
	return b.FileLocation.Encode(buf)
}

// FileLocationClassArray is adapter for slice of FileLocationClass.
type FileLocationClassArray []FileLocationClass

// Sort sorts slice of FileLocationClass.
func (s FileLocationClassArray) Sort(less func(a, b FileLocationClass) bool) FileLocationClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of FileLocationClass.
func (s FileLocationClassArray) SortStable(less func(a, b FileLocationClass) bool) FileLocationClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of FileLocationClass.
func (s FileLocationClassArray) Retain(keep func(x FileLocationClass) bool) FileLocationClassArray {
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
func (s FileLocationClassArray) First() (v FileLocationClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s FileLocationClassArray) Last() (v FileLocationClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *FileLocationClassArray) PopFirst() (v FileLocationClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero FileLocationClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *FileLocationClassArray) Pop() (v FileLocationClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsFileLocationUnavailable returns copy with only FileLocationUnavailable constructors.
func (s FileLocationClassArray) AsFileLocationUnavailable() (to FileLocationUnavailableArray) {
	for _, elem := range s {
		value, ok := elem.(*FileLocationUnavailable)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsFileLocation returns copy with only FileLocation constructors.
func (s FileLocationClassArray) AsFileLocation() (to FileLocationArray) {
	for _, elem := range s {
		value, ok := elem.(*FileLocation)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AppendOnlyAvailable appends only Available constructors to
// given slice.
func (s FileLocationClassArray) AppendOnlyAvailable(to []*FileLocation) []*FileLocation {
	for _, elem := range s {
		value, ok := elem.AsAvailable()
		if !ok {
			continue
		}
		to = append(to, value)
	}

	return to
}

// AsAvailable returns copy with only Available constructors.
func (s FileLocationClassArray) AsAvailable() (to []*FileLocation) {
	return s.AppendOnlyAvailable(to)
}

// FirstAsAvailable returns first element of slice (if exists).
func (s FileLocationClassArray) FirstAsAvailable() (v *FileLocation, ok bool) {
	value, ok := s.First()
	if !ok {
		return
	}
	return value.AsAvailable()
}

// LastAsAvailable returns last element of slice (if exists).
func (s FileLocationClassArray) LastAsAvailable() (v *FileLocation, ok bool) {
	value, ok := s.Last()
	if !ok {
		return
	}
	return value.AsAvailable()
}

// PopFirstAsAvailable returns element of slice (if exists).
func (s *FileLocationClassArray) PopFirstAsAvailable() (v *FileLocation, ok bool) {
	value, ok := s.PopFirst()
	if !ok {
		return
	}
	return value.AsAvailable()
}

// PopAsAvailable returns element of slice (if exists).
func (s *FileLocationClassArray) PopAsAvailable() (v *FileLocation, ok bool) {
	value, ok := s.Pop()
	if !ok {
		return
	}
	return value.AsAvailable()
}

// FileLocationUnavailableArray is adapter for slice of FileLocationUnavailable.
type FileLocationUnavailableArray []FileLocationUnavailable

// Sort sorts slice of FileLocationUnavailable.
func (s FileLocationUnavailableArray) Sort(less func(a, b FileLocationUnavailable) bool) FileLocationUnavailableArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of FileLocationUnavailable.
func (s FileLocationUnavailableArray) SortStable(less func(a, b FileLocationUnavailable) bool) FileLocationUnavailableArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of FileLocationUnavailable.
func (s FileLocationUnavailableArray) Retain(keep func(x FileLocationUnavailable) bool) FileLocationUnavailableArray {
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
func (s FileLocationUnavailableArray) First() (v FileLocationUnavailable, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s FileLocationUnavailableArray) Last() (v FileLocationUnavailable, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *FileLocationUnavailableArray) PopFirst() (v FileLocationUnavailable, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero FileLocationUnavailable
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *FileLocationUnavailableArray) Pop() (v FileLocationUnavailable, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// FileLocationArray is adapter for slice of FileLocation.
type FileLocationArray []FileLocation

// Sort sorts slice of FileLocation.
func (s FileLocationArray) Sort(less func(a, b FileLocation) bool) FileLocationArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of FileLocation.
func (s FileLocationArray) SortStable(less func(a, b FileLocation) bool) FileLocationArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of FileLocation.
func (s FileLocationArray) Retain(keep func(x FileLocation) bool) FileLocationArray {
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
func (s FileLocationArray) First() (v FileLocation, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s FileLocationArray) Last() (v FileLocation, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *FileLocationArray) PopFirst() (v FileLocation, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero FileLocation
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *FileLocationArray) Pop() (v FileLocation, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
