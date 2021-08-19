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

// DialogPeer represents TL type `dialogPeer#e56dbf05`.
// Peer
//
// See https://core.telegram.org/constructor/dialogPeer for reference.
type DialogPeer struct {
	// Peer
	Peer PeerClass
}

// DialogPeerTypeID is TL type id of DialogPeer.
const DialogPeerTypeID = 0xe56dbf05

func (d *DialogPeer) Zero() bool {
	if d == nil {
		return true
	}
	if !(d.Peer == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (d *DialogPeer) String() string {
	if d == nil {
		return "DialogPeer(nil)"
	}
	type Alias DialogPeer
	return fmt.Sprintf("DialogPeer%+v", Alias(*d))
}

// FillFrom fills DialogPeer from given interface.
func (d *DialogPeer) FillFrom(from interface {
	GetPeer() (value PeerClass)
}) {
	d.Peer = from.GetPeer()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*DialogPeer) TypeID() uint32 {
	return DialogPeerTypeID
}

// TypeName returns name of type in TL schema.
func (*DialogPeer) TypeName() string {
	return "dialogPeer"
}

// TypeInfo returns info about TL type.
func (d *DialogPeer) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "dialogPeer",
		ID:   DialogPeerTypeID,
	}
	if d == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (d *DialogPeer) Encode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dialogPeer#e56dbf05",
		}
	}
	b.PutID(DialogPeerTypeID)
	return d.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (d *DialogPeer) EncodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dialogPeer#e56dbf05",
		}
	}
	if d.Peer == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "dialogPeer#e56dbf05",
			FieldName: "peer",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "Peer",
			},
		}
	}
	if err := d.Peer.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "dialogPeer#e56dbf05",
			FieldName:  "peer",
			Underlying: err,
		}
	}
	return nil
}

// GetPeer returns value of Peer field.
func (d *DialogPeer) GetPeer() (value PeerClass) {
	return d.Peer
}

// Decode implements bin.Decoder.
func (d *DialogPeer) Decode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dialogPeer#e56dbf05",
		}
	}
	if err := b.ConsumeID(DialogPeerTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "dialogPeer#e56dbf05",
			Underlying: err,
		}
	}
	return d.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (d *DialogPeer) DecodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dialogPeer#e56dbf05",
		}
	}
	{
		value, err := DecodePeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogPeer#e56dbf05",
				FieldName:  "peer",
				Underlying: err,
			}
		}
		d.Peer = value
	}
	return nil
}

// construct implements constructor of DialogPeerClass.
func (d DialogPeer) construct() DialogPeerClass { return &d }

// Ensuring interfaces in compile-time for DialogPeer.
var (
	_ bin.Encoder     = &DialogPeer{}
	_ bin.Decoder     = &DialogPeer{}
	_ bin.BareEncoder = &DialogPeer{}
	_ bin.BareDecoder = &DialogPeer{}

	_ DialogPeerClass = &DialogPeer{}
)

// DialogPeerFolder represents TL type `dialogPeerFolder#514519e2`.
// Peer folder¹
//
// Links:
//  1) https://core.telegram.org/api/folders#peer-folders
//
// See https://core.telegram.org/constructor/dialogPeerFolder for reference.
type DialogPeerFolder struct {
	// Peer folder ID, for more info click here¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders#peer-folders
	FolderID int
}

// DialogPeerFolderTypeID is TL type id of DialogPeerFolder.
const DialogPeerFolderTypeID = 0x514519e2

func (d *DialogPeerFolder) Zero() bool {
	if d == nil {
		return true
	}
	if !(d.FolderID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (d *DialogPeerFolder) String() string {
	if d == nil {
		return "DialogPeerFolder(nil)"
	}
	type Alias DialogPeerFolder
	return fmt.Sprintf("DialogPeerFolder%+v", Alias(*d))
}

// FillFrom fills DialogPeerFolder from given interface.
func (d *DialogPeerFolder) FillFrom(from interface {
	GetFolderID() (value int)
}) {
	d.FolderID = from.GetFolderID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*DialogPeerFolder) TypeID() uint32 {
	return DialogPeerFolderTypeID
}

// TypeName returns name of type in TL schema.
func (*DialogPeerFolder) TypeName() string {
	return "dialogPeerFolder"
}

// TypeInfo returns info about TL type.
func (d *DialogPeerFolder) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "dialogPeerFolder",
		ID:   DialogPeerFolderTypeID,
	}
	if d == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "FolderID",
			SchemaName: "folder_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (d *DialogPeerFolder) Encode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dialogPeerFolder#514519e2",
		}
	}
	b.PutID(DialogPeerFolderTypeID)
	return d.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (d *DialogPeerFolder) EncodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dialogPeerFolder#514519e2",
		}
	}
	b.PutInt(d.FolderID)
	return nil
}

// GetFolderID returns value of FolderID field.
func (d *DialogPeerFolder) GetFolderID() (value int) {
	return d.FolderID
}

// Decode implements bin.Decoder.
func (d *DialogPeerFolder) Decode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dialogPeerFolder#514519e2",
		}
	}
	if err := b.ConsumeID(DialogPeerFolderTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "dialogPeerFolder#514519e2",
			Underlying: err,
		}
	}
	return d.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (d *DialogPeerFolder) DecodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dialogPeerFolder#514519e2",
		}
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogPeerFolder#514519e2",
				FieldName:  "folder_id",
				Underlying: err,
			}
		}
		d.FolderID = value
	}
	return nil
}

// construct implements constructor of DialogPeerClass.
func (d DialogPeerFolder) construct() DialogPeerClass { return &d }

// Ensuring interfaces in compile-time for DialogPeerFolder.
var (
	_ bin.Encoder     = &DialogPeerFolder{}
	_ bin.Decoder     = &DialogPeerFolder{}
	_ bin.BareEncoder = &DialogPeerFolder{}
	_ bin.BareDecoder = &DialogPeerFolder{}

	_ DialogPeerClass = &DialogPeerFolder{}
)

// DialogPeerClass represents DialogPeer generic type.
//
// See https://core.telegram.org/type/DialogPeer for reference.
//
// Example:
//  g, err := tg.DecodeDialogPeer(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.DialogPeer: // dialogPeer#e56dbf05
//  case *tg.DialogPeerFolder: // dialogPeerFolder#514519e2
//  default: panic(v)
//  }
type DialogPeerClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() DialogPeerClass

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

// AsInput tries to map DialogPeerFolder to InputDialogPeerFolder.
func (d *DialogPeerFolder) AsInput() *InputDialogPeerFolder {
	value := new(InputDialogPeerFolder)
	value.FolderID = d.GetFolderID()

	return value
}

// DecodeDialogPeer implements binary de-serialization for DialogPeerClass.
func DecodeDialogPeer(buf *bin.Buffer) (DialogPeerClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case DialogPeerTypeID:
		// Decoding dialogPeer#e56dbf05.
		v := DialogPeer{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "DialogPeerClass",
				Underlying: err,
			}
		}
		return &v, nil
	case DialogPeerFolderTypeID:
		// Decoding dialogPeerFolder#514519e2.
		v := DialogPeerFolder{}
		if err := v.Decode(buf); err != nil {
			return nil, &bin.DecodeError{
				TypeName:   "DialogPeerClass",
				Underlying: err,
			}
		}
		return &v, nil
	default:
		return nil, &bin.DecodeError{
			TypeName:   "DialogPeerClass",
			Underlying: bin.NewUnexpectedID(id),
		}
	}
}

// DialogPeer boxes the DialogPeerClass providing a helper.
type DialogPeerBox struct {
	DialogPeer DialogPeerClass
}

// Decode implements bin.Decoder for DialogPeerBox.
func (b *DialogPeerBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "DialogPeerBox",
		}
	}
	v, err := DecodeDialogPeer(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.DialogPeer = v
	return nil
}

// Encode implements bin.Encode for DialogPeerBox.
func (b *DialogPeerBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.DialogPeer == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "DialogPeerBox",
		}
	}
	return b.DialogPeer.Encode(buf)
}

// DialogPeerClassArray is adapter for slice of DialogPeerClass.
type DialogPeerClassArray []DialogPeerClass

// Sort sorts slice of DialogPeerClass.
func (s DialogPeerClassArray) Sort(less func(a, b DialogPeerClass) bool) DialogPeerClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of DialogPeerClass.
func (s DialogPeerClassArray) SortStable(less func(a, b DialogPeerClass) bool) DialogPeerClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of DialogPeerClass.
func (s DialogPeerClassArray) Retain(keep func(x DialogPeerClass) bool) DialogPeerClassArray {
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
func (s DialogPeerClassArray) First() (v DialogPeerClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s DialogPeerClassArray) Last() (v DialogPeerClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *DialogPeerClassArray) PopFirst() (v DialogPeerClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero DialogPeerClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *DialogPeerClassArray) Pop() (v DialogPeerClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsDialogPeer returns copy with only DialogPeer constructors.
func (s DialogPeerClassArray) AsDialogPeer() (to DialogPeerArray) {
	for _, elem := range s {
		value, ok := elem.(*DialogPeer)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsDialogPeerFolder returns copy with only DialogPeerFolder constructors.
func (s DialogPeerClassArray) AsDialogPeerFolder() (to DialogPeerFolderArray) {
	for _, elem := range s {
		value, ok := elem.(*DialogPeerFolder)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// DialogPeerArray is adapter for slice of DialogPeer.
type DialogPeerArray []DialogPeer

// Sort sorts slice of DialogPeer.
func (s DialogPeerArray) Sort(less func(a, b DialogPeer) bool) DialogPeerArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of DialogPeer.
func (s DialogPeerArray) SortStable(less func(a, b DialogPeer) bool) DialogPeerArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of DialogPeer.
func (s DialogPeerArray) Retain(keep func(x DialogPeer) bool) DialogPeerArray {
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
func (s DialogPeerArray) First() (v DialogPeer, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s DialogPeerArray) Last() (v DialogPeer, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *DialogPeerArray) PopFirst() (v DialogPeer, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero DialogPeer
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *DialogPeerArray) Pop() (v DialogPeer, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// DialogPeerFolderArray is adapter for slice of DialogPeerFolder.
type DialogPeerFolderArray []DialogPeerFolder

// Sort sorts slice of DialogPeerFolder.
func (s DialogPeerFolderArray) Sort(less func(a, b DialogPeerFolder) bool) DialogPeerFolderArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of DialogPeerFolder.
func (s DialogPeerFolderArray) SortStable(less func(a, b DialogPeerFolder) bool) DialogPeerFolderArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of DialogPeerFolder.
func (s DialogPeerFolderArray) Retain(keep func(x DialogPeerFolder) bool) DialogPeerFolderArray {
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
func (s DialogPeerFolderArray) First() (v DialogPeerFolder, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s DialogPeerFolderArray) Last() (v DialogPeerFolder, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *DialogPeerFolderArray) PopFirst() (v DialogPeerFolder, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero DialogPeerFolder
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *DialogPeerFolderArray) Pop() (v DialogPeerFolder, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
