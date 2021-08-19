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

// InputFolderPeer represents TL type `inputFolderPeer#fbd2c296`.
// Peer in a folder
//
// See https://core.telegram.org/constructor/inputFolderPeer for reference.
type InputFolderPeer struct {
	// Peer
	Peer InputPeerClass
	// Peer folder ID, for more info click here¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders#peer-folders
	FolderID int
}

// InputFolderPeerTypeID is TL type id of InputFolderPeer.
const InputFolderPeerTypeID = 0xfbd2c296

func (i *InputFolderPeer) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Peer == nil) {
		return false
	}
	if !(i.FolderID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputFolderPeer) String() string {
	if i == nil {
		return "InputFolderPeer(nil)"
	}
	type Alias InputFolderPeer
	return fmt.Sprintf("InputFolderPeer%+v", Alias(*i))
}

// FillFrom fills InputFolderPeer from given interface.
func (i *InputFolderPeer) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
	GetFolderID() (value int)
}) {
	i.Peer = from.GetPeer()
	i.FolderID = from.GetFolderID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputFolderPeer) TypeID() uint32 {
	return InputFolderPeerTypeID
}

// TypeName returns name of type in TL schema.
func (*InputFolderPeer) TypeName() string {
	return "inputFolderPeer"
}

// TypeInfo returns info about TL type.
func (i *InputFolderPeer) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputFolderPeer",
		ID:   InputFolderPeerTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "FolderID",
			SchemaName: "folder_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (i *InputFolderPeer) Encode(b *bin.Buffer) error {
	if i == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "inputFolderPeer#fbd2c296",
		}
	}
	b.PutID(InputFolderPeerTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputFolderPeer) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "inputFolderPeer#fbd2c296",
		}
	}
	if i.Peer == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "inputFolderPeer#fbd2c296",
			FieldName: "peer",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputPeer",
			},
		}
	}
	if err := i.Peer.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "inputFolderPeer#fbd2c296",
			FieldName:  "peer",
			Underlying: err,
		}
	}
	b.PutInt(i.FolderID)
	return nil
}

// GetPeer returns value of Peer field.
func (i *InputFolderPeer) GetPeer() (value InputPeerClass) {
	return i.Peer
}

// GetFolderID returns value of FolderID field.
func (i *InputFolderPeer) GetFolderID() (value int) {
	return i.FolderID
}

// Decode implements bin.Decoder.
func (i *InputFolderPeer) Decode(b *bin.Buffer) error {
	if i == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "inputFolderPeer#fbd2c296",
		}
	}
	if err := b.ConsumeID(InputFolderPeerTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "inputFolderPeer#fbd2c296",
			Underlying: err,
		}
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputFolderPeer) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "inputFolderPeer#fbd2c296",
		}
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "inputFolderPeer#fbd2c296",
				FieldName:  "peer",
				Underlying: err,
			}
		}
		i.Peer = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "inputFolderPeer#fbd2c296",
				FieldName:  "folder_id",
				Underlying: err,
			}
		}
		i.FolderID = value
	}
	return nil
}

// Ensuring interfaces in compile-time for InputFolderPeer.
var (
	_ bin.Encoder     = &InputFolderPeer{}
	_ bin.Decoder     = &InputFolderPeer{}
	_ bin.BareEncoder = &InputFolderPeer{}
	_ bin.BareDecoder = &InputFolderPeer{}
)
