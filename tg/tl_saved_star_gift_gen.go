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
	"github.com/gotd/td/tdjson"
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
	_ = tdjson.Encoder{}
)

// SavedStarGift represents TL type `savedStarGift#6056dba5`.
//
// See https://core.telegram.org/constructor/savedStarGift for reference.
type SavedStarGift struct {
	// Flags field of SavedStarGift.
	Flags bin.Fields
	// NameHidden field of SavedStarGift.
	NameHidden bool
	// Unsaved field of SavedStarGift.
	Unsaved bool
	// Refunded field of SavedStarGift.
	Refunded bool
	// CanUpgrade field of SavedStarGift.
	CanUpgrade bool
	// FromID field of SavedStarGift.
	//
	// Use SetFromID and GetFromID helpers.
	FromID PeerClass
	// Date field of SavedStarGift.
	Date int
	// Gift field of SavedStarGift.
	Gift StarGiftClass
	// Message field of SavedStarGift.
	//
	// Use SetMessage and GetMessage helpers.
	Message TextWithEntities
	// MsgID field of SavedStarGift.
	//
	// Use SetMsgID and GetMsgID helpers.
	MsgID int
	// SavedID field of SavedStarGift.
	//
	// Use SetSavedID and GetSavedID helpers.
	SavedID int64
	// ConvertStars field of SavedStarGift.
	//
	// Use SetConvertStars and GetConvertStars helpers.
	ConvertStars int64
	// UpgradeStars field of SavedStarGift.
	//
	// Use SetUpgradeStars and GetUpgradeStars helpers.
	UpgradeStars int64
	// CanExportAt field of SavedStarGift.
	//
	// Use SetCanExportAt and GetCanExportAt helpers.
	CanExportAt int
	// TransferStars field of SavedStarGift.
	//
	// Use SetTransferStars and GetTransferStars helpers.
	TransferStars int64
}

// SavedStarGiftTypeID is TL type id of SavedStarGift.
const SavedStarGiftTypeID = 0x6056dba5

// Ensuring interfaces in compile-time for SavedStarGift.
var (
	_ bin.Encoder     = &SavedStarGift{}
	_ bin.Decoder     = &SavedStarGift{}
	_ bin.BareEncoder = &SavedStarGift{}
	_ bin.BareDecoder = &SavedStarGift{}
)

func (s *SavedStarGift) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Flags.Zero()) {
		return false
	}
	if !(s.NameHidden == false) {
		return false
	}
	if !(s.Unsaved == false) {
		return false
	}
	if !(s.Refunded == false) {
		return false
	}
	if !(s.CanUpgrade == false) {
		return false
	}
	if !(s.FromID == nil) {
		return false
	}
	if !(s.Date == 0) {
		return false
	}
	if !(s.Gift == nil) {
		return false
	}
	if !(s.Message.Zero()) {
		return false
	}
	if !(s.MsgID == 0) {
		return false
	}
	if !(s.SavedID == 0) {
		return false
	}
	if !(s.ConvertStars == 0) {
		return false
	}
	if !(s.UpgradeStars == 0) {
		return false
	}
	if !(s.CanExportAt == 0) {
		return false
	}
	if !(s.TransferStars == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SavedStarGift) String() string {
	if s == nil {
		return "SavedStarGift(nil)"
	}
	type Alias SavedStarGift
	return fmt.Sprintf("SavedStarGift%+v", Alias(*s))
}

// FillFrom fills SavedStarGift from given interface.
func (s *SavedStarGift) FillFrom(from interface {
	GetNameHidden() (value bool)
	GetUnsaved() (value bool)
	GetRefunded() (value bool)
	GetCanUpgrade() (value bool)
	GetFromID() (value PeerClass, ok bool)
	GetDate() (value int)
	GetGift() (value StarGiftClass)
	GetMessage() (value TextWithEntities, ok bool)
	GetMsgID() (value int, ok bool)
	GetSavedID() (value int64, ok bool)
	GetConvertStars() (value int64, ok bool)
	GetUpgradeStars() (value int64, ok bool)
	GetCanExportAt() (value int, ok bool)
	GetTransferStars() (value int64, ok bool)
}) {
	s.NameHidden = from.GetNameHidden()
	s.Unsaved = from.GetUnsaved()
	s.Refunded = from.GetRefunded()
	s.CanUpgrade = from.GetCanUpgrade()
	if val, ok := from.GetFromID(); ok {
		s.FromID = val
	}

	s.Date = from.GetDate()
	s.Gift = from.GetGift()
	if val, ok := from.GetMessage(); ok {
		s.Message = val
	}

	if val, ok := from.GetMsgID(); ok {
		s.MsgID = val
	}

	if val, ok := from.GetSavedID(); ok {
		s.SavedID = val
	}

	if val, ok := from.GetConvertStars(); ok {
		s.ConvertStars = val
	}

	if val, ok := from.GetUpgradeStars(); ok {
		s.UpgradeStars = val
	}

	if val, ok := from.GetCanExportAt(); ok {
		s.CanExportAt = val
	}

	if val, ok := from.GetTransferStars(); ok {
		s.TransferStars = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SavedStarGift) TypeID() uint32 {
	return SavedStarGiftTypeID
}

// TypeName returns name of type in TL schema.
func (*SavedStarGift) TypeName() string {
	return "savedStarGift"
}

// TypeInfo returns info about TL type.
func (s *SavedStarGift) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "savedStarGift",
		ID:   SavedStarGiftTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "NameHidden",
			SchemaName: "name_hidden",
			Null:       !s.Flags.Has(0),
		},
		{
			Name:       "Unsaved",
			SchemaName: "unsaved",
			Null:       !s.Flags.Has(5),
		},
		{
			Name:       "Refunded",
			SchemaName: "refunded",
			Null:       !s.Flags.Has(9),
		},
		{
			Name:       "CanUpgrade",
			SchemaName: "can_upgrade",
			Null:       !s.Flags.Has(10),
		},
		{
			Name:       "FromID",
			SchemaName: "from_id",
			Null:       !s.Flags.Has(1),
		},
		{
			Name:       "Date",
			SchemaName: "date",
		},
		{
			Name:       "Gift",
			SchemaName: "gift",
		},
		{
			Name:       "Message",
			SchemaName: "message",
			Null:       !s.Flags.Has(2),
		},
		{
			Name:       "MsgID",
			SchemaName: "msg_id",
			Null:       !s.Flags.Has(3),
		},
		{
			Name:       "SavedID",
			SchemaName: "saved_id",
			Null:       !s.Flags.Has(11),
		},
		{
			Name:       "ConvertStars",
			SchemaName: "convert_stars",
			Null:       !s.Flags.Has(4),
		},
		{
			Name:       "UpgradeStars",
			SchemaName: "upgrade_stars",
			Null:       !s.Flags.Has(6),
		},
		{
			Name:       "CanExportAt",
			SchemaName: "can_export_at",
			Null:       !s.Flags.Has(7),
		},
		{
			Name:       "TransferStars",
			SchemaName: "transfer_stars",
			Null:       !s.Flags.Has(8),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (s *SavedStarGift) SetFlags() {
	if !(s.NameHidden == false) {
		s.Flags.Set(0)
	}
	if !(s.Unsaved == false) {
		s.Flags.Set(5)
	}
	if !(s.Refunded == false) {
		s.Flags.Set(9)
	}
	if !(s.CanUpgrade == false) {
		s.Flags.Set(10)
	}
	if !(s.FromID == nil) {
		s.Flags.Set(1)
	}
	if !(s.Message.Zero()) {
		s.Flags.Set(2)
	}
	if !(s.MsgID == 0) {
		s.Flags.Set(3)
	}
	if !(s.SavedID == 0) {
		s.Flags.Set(11)
	}
	if !(s.ConvertStars == 0) {
		s.Flags.Set(4)
	}
	if !(s.UpgradeStars == 0) {
		s.Flags.Set(6)
	}
	if !(s.CanExportAt == 0) {
		s.Flags.Set(7)
	}
	if !(s.TransferStars == 0) {
		s.Flags.Set(8)
	}
}

// Encode implements bin.Encoder.
func (s *SavedStarGift) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode savedStarGift#6056dba5 as nil")
	}
	b.PutID(SavedStarGiftTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SavedStarGift) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode savedStarGift#6056dba5 as nil")
	}
	s.SetFlags()
	if err := s.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode savedStarGift#6056dba5: field flags: %w", err)
	}
	if s.Flags.Has(1) {
		if s.FromID == nil {
			return fmt.Errorf("unable to encode savedStarGift#6056dba5: field from_id is nil")
		}
		if err := s.FromID.Encode(b); err != nil {
			return fmt.Errorf("unable to encode savedStarGift#6056dba5: field from_id: %w", err)
		}
	}
	b.PutInt(s.Date)
	if s.Gift == nil {
		return fmt.Errorf("unable to encode savedStarGift#6056dba5: field gift is nil")
	}
	if err := s.Gift.Encode(b); err != nil {
		return fmt.Errorf("unable to encode savedStarGift#6056dba5: field gift: %w", err)
	}
	if s.Flags.Has(2) {
		if err := s.Message.Encode(b); err != nil {
			return fmt.Errorf("unable to encode savedStarGift#6056dba5: field message: %w", err)
		}
	}
	if s.Flags.Has(3) {
		b.PutInt(s.MsgID)
	}
	if s.Flags.Has(11) {
		b.PutLong(s.SavedID)
	}
	if s.Flags.Has(4) {
		b.PutLong(s.ConvertStars)
	}
	if s.Flags.Has(6) {
		b.PutLong(s.UpgradeStars)
	}
	if s.Flags.Has(7) {
		b.PutInt(s.CanExportAt)
	}
	if s.Flags.Has(8) {
		b.PutLong(s.TransferStars)
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *SavedStarGift) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode savedStarGift#6056dba5 to nil")
	}
	if err := b.ConsumeID(SavedStarGiftTypeID); err != nil {
		return fmt.Errorf("unable to decode savedStarGift#6056dba5: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SavedStarGift) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode savedStarGift#6056dba5 to nil")
	}
	{
		if err := s.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field flags: %w", err)
		}
	}
	s.NameHidden = s.Flags.Has(0)
	s.Unsaved = s.Flags.Has(5)
	s.Refunded = s.Flags.Has(9)
	s.CanUpgrade = s.Flags.Has(10)
	if s.Flags.Has(1) {
		value, err := DecodePeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field from_id: %w", err)
		}
		s.FromID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field date: %w", err)
		}
		s.Date = value
	}
	{
		value, err := DecodeStarGift(b)
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field gift: %w", err)
		}
		s.Gift = value
	}
	if s.Flags.Has(2) {
		if err := s.Message.Decode(b); err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field message: %w", err)
		}
	}
	if s.Flags.Has(3) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field msg_id: %w", err)
		}
		s.MsgID = value
	}
	if s.Flags.Has(11) {
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field saved_id: %w", err)
		}
		s.SavedID = value
	}
	if s.Flags.Has(4) {
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field convert_stars: %w", err)
		}
		s.ConvertStars = value
	}
	if s.Flags.Has(6) {
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field upgrade_stars: %w", err)
		}
		s.UpgradeStars = value
	}
	if s.Flags.Has(7) {
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field can_export_at: %w", err)
		}
		s.CanExportAt = value
	}
	if s.Flags.Has(8) {
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode savedStarGift#6056dba5: field transfer_stars: %w", err)
		}
		s.TransferStars = value
	}
	return nil
}

// SetNameHidden sets value of NameHidden conditional field.
func (s *SavedStarGift) SetNameHidden(value bool) {
	if value {
		s.Flags.Set(0)
		s.NameHidden = true
	} else {
		s.Flags.Unset(0)
		s.NameHidden = false
	}
}

// GetNameHidden returns value of NameHidden conditional field.
func (s *SavedStarGift) GetNameHidden() (value bool) {
	if s == nil {
		return
	}
	return s.Flags.Has(0)
}

// SetUnsaved sets value of Unsaved conditional field.
func (s *SavedStarGift) SetUnsaved(value bool) {
	if value {
		s.Flags.Set(5)
		s.Unsaved = true
	} else {
		s.Flags.Unset(5)
		s.Unsaved = false
	}
}

// GetUnsaved returns value of Unsaved conditional field.
func (s *SavedStarGift) GetUnsaved() (value bool) {
	if s == nil {
		return
	}
	return s.Flags.Has(5)
}

// SetRefunded sets value of Refunded conditional field.
func (s *SavedStarGift) SetRefunded(value bool) {
	if value {
		s.Flags.Set(9)
		s.Refunded = true
	} else {
		s.Flags.Unset(9)
		s.Refunded = false
	}
}

// GetRefunded returns value of Refunded conditional field.
func (s *SavedStarGift) GetRefunded() (value bool) {
	if s == nil {
		return
	}
	return s.Flags.Has(9)
}

// SetCanUpgrade sets value of CanUpgrade conditional field.
func (s *SavedStarGift) SetCanUpgrade(value bool) {
	if value {
		s.Flags.Set(10)
		s.CanUpgrade = true
	} else {
		s.Flags.Unset(10)
		s.CanUpgrade = false
	}
}

// GetCanUpgrade returns value of CanUpgrade conditional field.
func (s *SavedStarGift) GetCanUpgrade() (value bool) {
	if s == nil {
		return
	}
	return s.Flags.Has(10)
}

// SetFromID sets value of FromID conditional field.
func (s *SavedStarGift) SetFromID(value PeerClass) {
	s.Flags.Set(1)
	s.FromID = value
}

// GetFromID returns value of FromID conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetFromID() (value PeerClass, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(1) {
		return value, false
	}
	return s.FromID, true
}

// GetDate returns value of Date field.
func (s *SavedStarGift) GetDate() (value int) {
	if s == nil {
		return
	}
	return s.Date
}

// GetGift returns value of Gift field.
func (s *SavedStarGift) GetGift() (value StarGiftClass) {
	if s == nil {
		return
	}
	return s.Gift
}

// SetMessage sets value of Message conditional field.
func (s *SavedStarGift) SetMessage(value TextWithEntities) {
	s.Flags.Set(2)
	s.Message = value
}

// GetMessage returns value of Message conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetMessage() (value TextWithEntities, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(2) {
		return value, false
	}
	return s.Message, true
}

// SetMsgID sets value of MsgID conditional field.
func (s *SavedStarGift) SetMsgID(value int) {
	s.Flags.Set(3)
	s.MsgID = value
}

// GetMsgID returns value of MsgID conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetMsgID() (value int, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(3) {
		return value, false
	}
	return s.MsgID, true
}

// SetSavedID sets value of SavedID conditional field.
func (s *SavedStarGift) SetSavedID(value int64) {
	s.Flags.Set(11)
	s.SavedID = value
}

// GetSavedID returns value of SavedID conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetSavedID() (value int64, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(11) {
		return value, false
	}
	return s.SavedID, true
}

// SetConvertStars sets value of ConvertStars conditional field.
func (s *SavedStarGift) SetConvertStars(value int64) {
	s.Flags.Set(4)
	s.ConvertStars = value
}

// GetConvertStars returns value of ConvertStars conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetConvertStars() (value int64, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(4) {
		return value, false
	}
	return s.ConvertStars, true
}

// SetUpgradeStars sets value of UpgradeStars conditional field.
func (s *SavedStarGift) SetUpgradeStars(value int64) {
	s.Flags.Set(6)
	s.UpgradeStars = value
}

// GetUpgradeStars returns value of UpgradeStars conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetUpgradeStars() (value int64, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(6) {
		return value, false
	}
	return s.UpgradeStars, true
}

// SetCanExportAt sets value of CanExportAt conditional field.
func (s *SavedStarGift) SetCanExportAt(value int) {
	s.Flags.Set(7)
	s.CanExportAt = value
}

// GetCanExportAt returns value of CanExportAt conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetCanExportAt() (value int, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(7) {
		return value, false
	}
	return s.CanExportAt, true
}

// SetTransferStars sets value of TransferStars conditional field.
func (s *SavedStarGift) SetTransferStars(value int64) {
	s.Flags.Set(8)
	s.TransferStars = value
}

// GetTransferStars returns value of TransferStars conditional field and
// boolean which is true if field was set.
func (s *SavedStarGift) GetTransferStars() (value int64, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(8) {
		return value, false
	}
	return s.TransferStars, true
}