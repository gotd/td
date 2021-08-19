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

// UploadCDNFileReuploadNeeded represents TL type `upload.cdnFileReuploadNeeded#eea8e46e`.
// The file was cleared from the temporary RAM cache of the CDN¹ and has to be
// reuploaded.
//
// Links:
//  1) https://core.telegram.org/cdn
//
// See https://core.telegram.org/constructor/upload.cdnFileReuploadNeeded for reference.
type UploadCDNFileReuploadNeeded struct {
	// Request token (see CDN¹)
	//
	// Links:
	//  1) https://core.telegram.org/cdn
	RequestToken []byte
}

// UploadCDNFileReuploadNeededTypeID is TL type id of UploadCDNFileReuploadNeeded.
const UploadCDNFileReuploadNeededTypeID = 0xeea8e46e

func (c *UploadCDNFileReuploadNeeded) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.RequestToken == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *UploadCDNFileReuploadNeeded) String() string {
	if c == nil {
		return "UploadCDNFileReuploadNeeded(nil)"
	}
	type Alias UploadCDNFileReuploadNeeded
	return fmt.Sprintf("UploadCDNFileReuploadNeeded%+v", Alias(*c))
}

// FillFrom fills UploadCDNFileReuploadNeeded from given interface.
func (c *UploadCDNFileReuploadNeeded) FillFrom(from interface {
	GetRequestToken() (value []byte)
}) {
	c.RequestToken = from.GetRequestToken()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*UploadCDNFileReuploadNeeded) TypeID() uint32 {
	return UploadCDNFileReuploadNeededTypeID
}

// TypeName returns name of type in TL schema.
func (*UploadCDNFileReuploadNeeded) TypeName() string {
	return "upload.cdnFileReuploadNeeded"
}

// TypeInfo returns info about TL type.
func (c *UploadCDNFileReuploadNeeded) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "upload.cdnFileReuploadNeeded",
		ID:   UploadCDNFileReuploadNeededTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "RequestToken",
			SchemaName: "request_token",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *UploadCDNFileReuploadNeeded) Encode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "upload.cdnFileReuploadNeeded#eea8e46e",
		}
	}
	b.PutID(UploadCDNFileReuploadNeededTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *UploadCDNFileReuploadNeeded) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "upload.cdnFileReuploadNeeded#eea8e46e",
		}
	}
	b.PutBytes(c.RequestToken)
	return nil
}

// GetRequestToken returns value of RequestToken field.
func (c *UploadCDNFileReuploadNeeded) GetRequestToken() (value []byte) {
	return c.RequestToken
}

// Decode implements bin.Decoder.
func (c *UploadCDNFileReuploadNeeded) Decode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "upload.cdnFileReuploadNeeded#eea8e46e",
		}
	}
	if err := b.ConsumeID(UploadCDNFileReuploadNeededTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "upload.cdnFileReuploadNeeded#eea8e46e",
			Underlying: err,
		}
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *UploadCDNFileReuploadNeeded) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "upload.cdnFileReuploadNeeded#eea8e46e",
		}
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "upload.cdnFileReuploadNeeded#eea8e46e",
				FieldName:  "request_token",
				Underlying: err,
			}
		}
		c.RequestToken = value
	}
	return nil
}

// construct implements constructor of UploadCDNFileClass.
func (c UploadCDNFileReuploadNeeded) construct() UploadCDNFileClass { return &c }

// Ensuring interfaces in compile-time for UploadCDNFileReuploadNeeded.
var (
	_ bin.Encoder     = &UploadCDNFileReuploadNeeded{}
	_ bin.Decoder     = &UploadCDNFileReuploadNeeded{}
	_ bin.BareEncoder = &UploadCDNFileReuploadNeeded{}
	_ bin.BareDecoder = &UploadCDNFileReuploadNeeded{}

	_ UploadCDNFileClass = &UploadCDNFileReuploadNeeded{}
)

// UploadCDNFile represents TL type `upload.cdnFile#a99fca4f`.
// Represent a chunk of a CDN¹ file.
//
// Links:
//  1) https://core.telegram.org/cdn
//
// See https://core.telegram.org/constructor/upload.cdnFile for reference.
type UploadCDNFile struct {
	// The data
	Bytes []byte
}

// UploadCDNFileTypeID is TL type id of UploadCDNFile.
const UploadCDNFileTypeID = 0xa99fca4f

func (c *UploadCDNFile) Zero() bool {
	if c == nil {
		return true
	}
	if !(c.Bytes == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (c *UploadCDNFile) String() string {
	if c == nil {
		return "UploadCDNFile(nil)"
	}
	type Alias UploadCDNFile
	return fmt.Sprintf("UploadCDNFile%+v", Alias(*c))
}

// FillFrom fills UploadCDNFile from given interface.
func (c *UploadCDNFile) FillFrom(from interface {
	GetBytes() (value []byte)
}) {
	c.Bytes = from.GetBytes()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*UploadCDNFile) TypeID() uint32 {
	return UploadCDNFileTypeID
}

// TypeName returns name of type in TL schema.
func (*UploadCDNFile) TypeName() string {
	return "upload.cdnFile"
}

// TypeInfo returns info about TL type.
func (c *UploadCDNFile) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "upload.cdnFile",
		ID:   UploadCDNFileTypeID,
	}
	if c == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Bytes",
			SchemaName: "bytes",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (c *UploadCDNFile) Encode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "upload.cdnFile#a99fca4f",
		}
	}
	b.PutID(UploadCDNFileTypeID)
	return c.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (c *UploadCDNFile) EncodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "upload.cdnFile#a99fca4f",
		}
	}
	b.PutBytes(c.Bytes)
	return nil
}

// GetBytes returns value of Bytes field.
func (c *UploadCDNFile) GetBytes() (value []byte) {
	return c.Bytes
}

// Decode implements bin.Decoder.
func (c *UploadCDNFile) Decode(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "upload.cdnFile#a99fca4f",
		}
	}
	if err := b.ConsumeID(UploadCDNFileTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "upload.cdnFile#a99fca4f",
			Underlying: err,
		}
	}
	return c.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (c *UploadCDNFile) DecodeBare(b *bin.Buffer) error {
	if c == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "upload.cdnFile#a99fca4f",
		}
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "upload.cdnFile#a99fca4f",
				FieldName:  "bytes",
				Underlying: err,
			}
		}
		c.Bytes = value
	}
	return nil
}

// construct implements constructor of UploadCDNFileClass.
func (c UploadCDNFile) construct() UploadCDNFileClass { return &c }

// Ensuring interfaces in compile-time for UploadCDNFile.
var (
	_ bin.Encoder     = &UploadCDNFile{}
	_ bin.Decoder     = &UploadCDNFile{}
	_ bin.BareEncoder = &UploadCDNFile{}
	_ bin.BareDecoder = &UploadCDNFile{}

	_ UploadCDNFileClass = &UploadCDNFile{}
)

// UploadCDNFileClass represents upload.CdnFile generic type.
//
// See https://core.telegram.org/type/upload.CdnFile for reference.
//
// Example:
//  g, err := tg.DecodeUploadCDNFile(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.UploadCDNFileReuploadNeeded: // upload.cdnFileReuploadNeeded#eea8e46e
//  case *tg.UploadCDNFile: // upload.cdnFile#a99fca4f
//  default: panic(v)
//  }
type UploadCDNFileClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() UploadCDNFileClass

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
}

// DecodeUploadCDNFile implements binary de-serialization for UploadCDNFileClass.
func DecodeUploadCDNFile(buf *bin.Buffer) (UploadCDNFileClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case UploadCDNFileReuploadNeededTypeID:
		// Decoding upload.cdnFileReuploadNeeded#eea8e46e.
		v := UploadCDNFileReuploadNeeded{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "UploadCDNFileClass",
				Underlying: err,
			}
		}
		return &v, nil
	case UploadCDNFileTypeID:
		// Decoding upload.cdnFile#a99fca4f.
		v := UploadCDNFile{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "UploadCDNFileClass",
				Underlying: err,
			}
		}
		return &v, nil
	default:
		return nil, &bin.DecodeError{
			TypeName:   "UploadCDNFileClass",
			Underlying: bin.NewUnexpectedID(id),
		}
	}
}

// UploadCDNFile boxes the UploadCDNFileClass providing a helper.
type UploadCDNFileBox struct {
	CdnFile UploadCDNFileClass
}

// Decode implements bin.Decoder for UploadCDNFileBox.
func (b *UploadCDNFileBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "UploadCDNFileBox",
		}
	}
	v, err := DecodeUploadCDNFile(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.CdnFile = v
	return nil
}

// Encode implements bin.Encode for UploadCDNFileBox.
func (b *UploadCDNFileBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.CdnFile == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "UploadCDNFileBox",
		}
	}
	return b.CdnFile.Encode(buf)
}

// UploadCDNFileClassArray is adapter for slice of UploadCDNFileClass.
type UploadCDNFileClassArray []UploadCDNFileClass

// Sort sorts slice of UploadCDNFileClass.
func (s UploadCDNFileClassArray) Sort(less func(a, b UploadCDNFileClass) bool) UploadCDNFileClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UploadCDNFileClass.
func (s UploadCDNFileClassArray) SortStable(less func(a, b UploadCDNFileClass) bool) UploadCDNFileClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UploadCDNFileClass.
func (s UploadCDNFileClassArray) Retain(keep func(x UploadCDNFileClass) bool) UploadCDNFileClassArray {
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
func (s UploadCDNFileClassArray) First() (v UploadCDNFileClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UploadCDNFileClassArray) Last() (v UploadCDNFileClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UploadCDNFileClassArray) PopFirst() (v UploadCDNFileClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UploadCDNFileClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UploadCDNFileClassArray) Pop() (v UploadCDNFileClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsUploadCDNFileReuploadNeeded returns copy with only UploadCDNFileReuploadNeeded constructors.
func (s UploadCDNFileClassArray) AsUploadCDNFileReuploadNeeded() (to UploadCDNFileReuploadNeededArray) {
	for _, elem := range s {
		value, ok := elem.(*UploadCDNFileReuploadNeeded)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsUploadCDNFile returns copy with only UploadCDNFile constructors.
func (s UploadCDNFileClassArray) AsUploadCDNFile() (to UploadCDNFileArray) {
	for _, elem := range s {
		value, ok := elem.(*UploadCDNFile)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// UploadCDNFileReuploadNeededArray is adapter for slice of UploadCDNFileReuploadNeeded.
type UploadCDNFileReuploadNeededArray []UploadCDNFileReuploadNeeded

// Sort sorts slice of UploadCDNFileReuploadNeeded.
func (s UploadCDNFileReuploadNeededArray) Sort(less func(a, b UploadCDNFileReuploadNeeded) bool) UploadCDNFileReuploadNeededArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UploadCDNFileReuploadNeeded.
func (s UploadCDNFileReuploadNeededArray) SortStable(less func(a, b UploadCDNFileReuploadNeeded) bool) UploadCDNFileReuploadNeededArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UploadCDNFileReuploadNeeded.
func (s UploadCDNFileReuploadNeededArray) Retain(keep func(x UploadCDNFileReuploadNeeded) bool) UploadCDNFileReuploadNeededArray {
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
func (s UploadCDNFileReuploadNeededArray) First() (v UploadCDNFileReuploadNeeded, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UploadCDNFileReuploadNeededArray) Last() (v UploadCDNFileReuploadNeeded, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UploadCDNFileReuploadNeededArray) PopFirst() (v UploadCDNFileReuploadNeeded, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UploadCDNFileReuploadNeeded
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UploadCDNFileReuploadNeededArray) Pop() (v UploadCDNFileReuploadNeeded, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// UploadCDNFileArray is adapter for slice of UploadCDNFile.
type UploadCDNFileArray []UploadCDNFile

// Sort sorts slice of UploadCDNFile.
func (s UploadCDNFileArray) Sort(less func(a, b UploadCDNFile) bool) UploadCDNFileArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UploadCDNFile.
func (s UploadCDNFileArray) SortStable(less func(a, b UploadCDNFile) bool) UploadCDNFileArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UploadCDNFile.
func (s UploadCDNFileArray) Retain(keep func(x UploadCDNFile) bool) UploadCDNFileArray {
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
func (s UploadCDNFileArray) First() (v UploadCDNFile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UploadCDNFileArray) Last() (v UploadCDNFile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UploadCDNFileArray) PopFirst() (v UploadCDNFile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UploadCDNFile
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UploadCDNFileArray) Pop() (v UploadCDNFile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
