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

// DialogFilter represents TL type `dialogFilter#7438f7e8`.
// Dialog filter AKA folder¹
//
// Links:
//  1) https://core.telegram.org/api/folders
//
// See https://core.telegram.org/constructor/dialogFilter for reference.
type DialogFilter struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Whether to include all contacts in this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	Contacts bool
	// Whether to include all non-contacts in this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	NonContacts bool
	// Whether to include all groups in this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	Groups bool
	// Whether to include all channels in this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	Broadcasts bool
	// Whether to include all bots in this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	Bots bool
	// Whether to exclude muted chats from this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	ExcludeMuted bool
	// Whether to exclude read chats from this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	ExcludeRead bool
	// Whether to exclude archived chats from this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	ExcludeArchived bool
	// Folder¹ ID
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	ID int
	// Folder¹ name
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	Title string
	// Folder¹ emoticon
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	//
	// Use SetEmoticon and GetEmoticon helpers.
	Emoticon string
	// Pinned chats, folders¹ can have unlimited pinned chats
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	PinnedPeers []InputPeerClass
	// Include the following chats in this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	IncludePeers []InputPeerClass
	// Exclude the following chats from this folder¹
	//
	// Links:
	//  1) https://core.telegram.org/api/folders
	ExcludePeers []InputPeerClass
}

// DialogFilterTypeID is TL type id of DialogFilter.
const DialogFilterTypeID = 0x7438f7e8

func (d *DialogFilter) Zero() bool {
	if d == nil {
		return true
	}
	if !(d.Flags.Zero()) {
		return false
	}
	if !(d.Contacts == false) {
		return false
	}
	if !(d.NonContacts == false) {
		return false
	}
	if !(d.Groups == false) {
		return false
	}
	if !(d.Broadcasts == false) {
		return false
	}
	if !(d.Bots == false) {
		return false
	}
	if !(d.ExcludeMuted == false) {
		return false
	}
	if !(d.ExcludeRead == false) {
		return false
	}
	if !(d.ExcludeArchived == false) {
		return false
	}
	if !(d.ID == 0) {
		return false
	}
	if !(d.Title == "") {
		return false
	}
	if !(d.Emoticon == "") {
		return false
	}
	if !(d.PinnedPeers == nil) {
		return false
	}
	if !(d.IncludePeers == nil) {
		return false
	}
	if !(d.ExcludePeers == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (d *DialogFilter) String() string {
	if d == nil {
		return "DialogFilter(nil)"
	}
	type Alias DialogFilter
	return fmt.Sprintf("DialogFilter%+v", Alias(*d))
}

// FillFrom fills DialogFilter from given interface.
func (d *DialogFilter) FillFrom(from interface {
	GetContacts() (value bool)
	GetNonContacts() (value bool)
	GetGroups() (value bool)
	GetBroadcasts() (value bool)
	GetBots() (value bool)
	GetExcludeMuted() (value bool)
	GetExcludeRead() (value bool)
	GetExcludeArchived() (value bool)
	GetID() (value int)
	GetTitle() (value string)
	GetEmoticon() (value string, ok bool)
	GetPinnedPeers() (value []InputPeerClass)
	GetIncludePeers() (value []InputPeerClass)
	GetExcludePeers() (value []InputPeerClass)
}) {
	d.Contacts = from.GetContacts()
	d.NonContacts = from.GetNonContacts()
	d.Groups = from.GetGroups()
	d.Broadcasts = from.GetBroadcasts()
	d.Bots = from.GetBots()
	d.ExcludeMuted = from.GetExcludeMuted()
	d.ExcludeRead = from.GetExcludeRead()
	d.ExcludeArchived = from.GetExcludeArchived()
	d.ID = from.GetID()
	d.Title = from.GetTitle()
	if val, ok := from.GetEmoticon(); ok {
		d.Emoticon = val
	}

	d.PinnedPeers = from.GetPinnedPeers()
	d.IncludePeers = from.GetIncludePeers()
	d.ExcludePeers = from.GetExcludePeers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*DialogFilter) TypeID() uint32 {
	return DialogFilterTypeID
}

// TypeName returns name of type in TL schema.
func (*DialogFilter) TypeName() string {
	return "dialogFilter"
}

// TypeInfo returns info about TL type.
func (d *DialogFilter) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "dialogFilter",
		ID:   DialogFilterTypeID,
	}
	if d == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Contacts",
			SchemaName: "contacts",
			Null:       !d.Flags.Has(0),
		},
		{
			Name:       "NonContacts",
			SchemaName: "non_contacts",
			Null:       !d.Flags.Has(1),
		},
		{
			Name:       "Groups",
			SchemaName: "groups",
			Null:       !d.Flags.Has(2),
		},
		{
			Name:       "Broadcasts",
			SchemaName: "broadcasts",
			Null:       !d.Flags.Has(3),
		},
		{
			Name:       "Bots",
			SchemaName: "bots",
			Null:       !d.Flags.Has(4),
		},
		{
			Name:       "ExcludeMuted",
			SchemaName: "exclude_muted",
			Null:       !d.Flags.Has(11),
		},
		{
			Name:       "ExcludeRead",
			SchemaName: "exclude_read",
			Null:       !d.Flags.Has(12),
		},
		{
			Name:       "ExcludeArchived",
			SchemaName: "exclude_archived",
			Null:       !d.Flags.Has(13),
		},
		{
			Name:       "ID",
			SchemaName: "id",
		},
		{
			Name:       "Title",
			SchemaName: "title",
		},
		{
			Name:       "Emoticon",
			SchemaName: "emoticon",
			Null:       !d.Flags.Has(25),
		},
		{
			Name:       "PinnedPeers",
			SchemaName: "pinned_peers",
		},
		{
			Name:       "IncludePeers",
			SchemaName: "include_peers",
		},
		{
			Name:       "ExcludePeers",
			SchemaName: "exclude_peers",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (d *DialogFilter) Encode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dialogFilter#7438f7e8",
		}
	}
	b.PutID(DialogFilterTypeID)
	return d.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (d *DialogFilter) EncodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "dialogFilter#7438f7e8",
		}
	}
	if !(d.Contacts == false) {
		d.Flags.Set(0)
	}
	if !(d.NonContacts == false) {
		d.Flags.Set(1)
	}
	if !(d.Groups == false) {
		d.Flags.Set(2)
	}
	if !(d.Broadcasts == false) {
		d.Flags.Set(3)
	}
	if !(d.Bots == false) {
		d.Flags.Set(4)
	}
	if !(d.ExcludeMuted == false) {
		d.Flags.Set(11)
	}
	if !(d.ExcludeRead == false) {
		d.Flags.Set(12)
	}
	if !(d.ExcludeArchived == false) {
		d.Flags.Set(13)
	}
	if !(d.Emoticon == "") {
		d.Flags.Set(25)
	}
	if err := d.Flags.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "dialogFilter#7438f7e8",
			FieldName:  "flags",
			Underlying: err,
		}
	}
	b.PutInt(d.ID)
	b.PutString(d.Title)
	if d.Flags.Has(25) {
		b.PutString(d.Emoticon)
	}
	b.PutVectorHeader(len(d.PinnedPeers))
	for idx, v := range d.PinnedPeers {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "dialogFilter#7438f7e8",
				FieldName: "pinned_peers",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<InputPeer>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "dialogFilter#7438f7e8",
				FieldName: "pinned_peers",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	b.PutVectorHeader(len(d.IncludePeers))
	for idx, v := range d.IncludePeers {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "dialogFilter#7438f7e8",
				FieldName: "include_peers",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<InputPeer>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "dialogFilter#7438f7e8",
				FieldName: "include_peers",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	b.PutVectorHeader(len(d.ExcludePeers))
	for idx, v := range d.ExcludePeers {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "dialogFilter#7438f7e8",
				FieldName: "exclude_peers",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<InputPeer>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "dialogFilter#7438f7e8",
				FieldName: "exclude_peers",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	return nil
}

// SetContacts sets value of Contacts conditional field.
func (d *DialogFilter) SetContacts(value bool) {
	if value {
		d.Flags.Set(0)
		d.Contacts = true
	} else {
		d.Flags.Unset(0)
		d.Contacts = false
	}
}

// GetContacts returns value of Contacts conditional field.
func (d *DialogFilter) GetContacts() (value bool) {
	return d.Flags.Has(0)
}

// SetNonContacts sets value of NonContacts conditional field.
func (d *DialogFilter) SetNonContacts(value bool) {
	if value {
		d.Flags.Set(1)
		d.NonContacts = true
	} else {
		d.Flags.Unset(1)
		d.NonContacts = false
	}
}

// GetNonContacts returns value of NonContacts conditional field.
func (d *DialogFilter) GetNonContacts() (value bool) {
	return d.Flags.Has(1)
}

// SetGroups sets value of Groups conditional field.
func (d *DialogFilter) SetGroups(value bool) {
	if value {
		d.Flags.Set(2)
		d.Groups = true
	} else {
		d.Flags.Unset(2)
		d.Groups = false
	}
}

// GetGroups returns value of Groups conditional field.
func (d *DialogFilter) GetGroups() (value bool) {
	return d.Flags.Has(2)
}

// SetBroadcasts sets value of Broadcasts conditional field.
func (d *DialogFilter) SetBroadcasts(value bool) {
	if value {
		d.Flags.Set(3)
		d.Broadcasts = true
	} else {
		d.Flags.Unset(3)
		d.Broadcasts = false
	}
}

// GetBroadcasts returns value of Broadcasts conditional field.
func (d *DialogFilter) GetBroadcasts() (value bool) {
	return d.Flags.Has(3)
}

// SetBots sets value of Bots conditional field.
func (d *DialogFilter) SetBots(value bool) {
	if value {
		d.Flags.Set(4)
		d.Bots = true
	} else {
		d.Flags.Unset(4)
		d.Bots = false
	}
}

// GetBots returns value of Bots conditional field.
func (d *DialogFilter) GetBots() (value bool) {
	return d.Flags.Has(4)
}

// SetExcludeMuted sets value of ExcludeMuted conditional field.
func (d *DialogFilter) SetExcludeMuted(value bool) {
	if value {
		d.Flags.Set(11)
		d.ExcludeMuted = true
	} else {
		d.Flags.Unset(11)
		d.ExcludeMuted = false
	}
}

// GetExcludeMuted returns value of ExcludeMuted conditional field.
func (d *DialogFilter) GetExcludeMuted() (value bool) {
	return d.Flags.Has(11)
}

// SetExcludeRead sets value of ExcludeRead conditional field.
func (d *DialogFilter) SetExcludeRead(value bool) {
	if value {
		d.Flags.Set(12)
		d.ExcludeRead = true
	} else {
		d.Flags.Unset(12)
		d.ExcludeRead = false
	}
}

// GetExcludeRead returns value of ExcludeRead conditional field.
func (d *DialogFilter) GetExcludeRead() (value bool) {
	return d.Flags.Has(12)
}

// SetExcludeArchived sets value of ExcludeArchived conditional field.
func (d *DialogFilter) SetExcludeArchived(value bool) {
	if value {
		d.Flags.Set(13)
		d.ExcludeArchived = true
	} else {
		d.Flags.Unset(13)
		d.ExcludeArchived = false
	}
}

// GetExcludeArchived returns value of ExcludeArchived conditional field.
func (d *DialogFilter) GetExcludeArchived() (value bool) {
	return d.Flags.Has(13)
}

// GetID returns value of ID field.
func (d *DialogFilter) GetID() (value int) {
	return d.ID
}

// GetTitle returns value of Title field.
func (d *DialogFilter) GetTitle() (value string) {
	return d.Title
}

// SetEmoticon sets value of Emoticon conditional field.
func (d *DialogFilter) SetEmoticon(value string) {
	d.Flags.Set(25)
	d.Emoticon = value
}

// GetEmoticon returns value of Emoticon conditional field and
// boolean which is true if field was set.
func (d *DialogFilter) GetEmoticon() (value string, ok bool) {
	if !d.Flags.Has(25) {
		return value, false
	}
	return d.Emoticon, true
}

// GetPinnedPeers returns value of PinnedPeers field.
func (d *DialogFilter) GetPinnedPeers() (value []InputPeerClass) {
	return d.PinnedPeers
}

// MapPinnedPeers returns field PinnedPeers wrapped in InputPeerClassArray helper.
func (d *DialogFilter) MapPinnedPeers() (value InputPeerClassArray) {
	return InputPeerClassArray(d.PinnedPeers)
}

// GetIncludePeers returns value of IncludePeers field.
func (d *DialogFilter) GetIncludePeers() (value []InputPeerClass) {
	return d.IncludePeers
}

// MapIncludePeers returns field IncludePeers wrapped in InputPeerClassArray helper.
func (d *DialogFilter) MapIncludePeers() (value InputPeerClassArray) {
	return InputPeerClassArray(d.IncludePeers)
}

// GetExcludePeers returns value of ExcludePeers field.
func (d *DialogFilter) GetExcludePeers() (value []InputPeerClass) {
	return d.ExcludePeers
}

// MapExcludePeers returns field ExcludePeers wrapped in InputPeerClassArray helper.
func (d *DialogFilter) MapExcludePeers() (value InputPeerClassArray) {
	return InputPeerClassArray(d.ExcludePeers)
}

// Decode implements bin.Decoder.
func (d *DialogFilter) Decode(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dialogFilter#7438f7e8",
		}
	}
	if err := b.ConsumeID(DialogFilterTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "dialogFilter#7438f7e8",
			Underlying: err,
		}
	}
	return d.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (d *DialogFilter) DecodeBare(b *bin.Buffer) error {
	if d == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "dialogFilter#7438f7e8",
		}
	}
	{
		if err := d.Flags.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogFilter#7438f7e8",
				FieldName:  "flags",
				Underlying: err,
			}
		}
	}
	d.Contacts = d.Flags.Has(0)
	d.NonContacts = d.Flags.Has(1)
	d.Groups = d.Flags.Has(2)
	d.Broadcasts = d.Flags.Has(3)
	d.Bots = d.Flags.Has(4)
	d.ExcludeMuted = d.Flags.Has(11)
	d.ExcludeRead = d.Flags.Has(12)
	d.ExcludeArchived = d.Flags.Has(13)
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogFilter#7438f7e8",
				FieldName:  "id",
				Underlying: err,
			}
		}
		d.ID = value
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogFilter#7438f7e8",
				FieldName:  "title",
				Underlying: err,
			}
		}
		d.Title = value
	}
	if d.Flags.Has(25) {
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogFilter#7438f7e8",
				FieldName:  "emoticon",
				Underlying: err,
			}
		}
		d.Emoticon = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogFilter#7438f7e8",
				FieldName:  "pinned_peers",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			d.PinnedPeers = make([]InputPeerClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputPeer(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "dialogFilter#7438f7e8",
					FieldName:  "pinned_peers",
					Underlying: err,
				}
			}
			d.PinnedPeers = append(d.PinnedPeers, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogFilter#7438f7e8",
				FieldName:  "include_peers",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			d.IncludePeers = make([]InputPeerClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputPeer(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "dialogFilter#7438f7e8",
					FieldName:  "include_peers",
					Underlying: err,
				}
			}
			d.IncludePeers = append(d.IncludePeers, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "dialogFilter#7438f7e8",
				FieldName:  "exclude_peers",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			d.ExcludePeers = make([]InputPeerClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputPeer(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "dialogFilter#7438f7e8",
					FieldName:  "exclude_peers",
					Underlying: err,
				}
			}
			d.ExcludePeers = append(d.ExcludePeers, value)
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for DialogFilter.
var (
	_ bin.Encoder     = &DialogFilter{}
	_ bin.Decoder     = &DialogFilter{}
	_ bin.BareEncoder = &DialogFilter{}
	_ bin.BareDecoder = &DialogFilter{}
)
