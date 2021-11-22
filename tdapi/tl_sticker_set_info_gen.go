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

// StickerSetInfo represents TL type `stickerSetInfo#aba733ac`.
type StickerSetInfo struct {
	// Identifier of the sticker set
	ID Int64
	// Title of the sticker set
	Title string
	// Name of the sticker set
	Name string
	// Sticker set thumbnail in WEBP or TGS format with width and height 100; may be null
	Thumbnail Thumbnail
	// Sticker set thumbnail's outline represented as a list of closed vector paths; may be
	// empty. The coordinate system origin is in the upper-left corner
	ThumbnailOutline []ClosedVectorPath
	// True, if the sticker set has been installed by the current user
	IsInstalled bool
	// True, if the sticker set has been archived. A sticker set can't be installed and
	// archived simultaneously
	IsArchived bool
	// True, if the sticker set is official
	IsOfficial bool
	// True, is the stickers in the set are animated
	IsAnimated bool
	// True, if the stickers in the set are masks
	IsMasks bool
	// True for already viewed trending sticker sets
	IsViewed bool
	// Total number of stickers in the set
	Size int32
	// Contains up to the first 5 stickers from the set, depending on the context. If the
	// application needs more stickers the full set should be requested
	Covers []Sticker
}

// StickerSetInfoTypeID is TL type id of StickerSetInfo.
const StickerSetInfoTypeID = 0xaba733ac

// Ensuring interfaces in compile-time for StickerSetInfo.
var (
	_ bin.Encoder     = &StickerSetInfo{}
	_ bin.Decoder     = &StickerSetInfo{}
	_ bin.BareEncoder = &StickerSetInfo{}
	_ bin.BareDecoder = &StickerSetInfo{}
)

func (s *StickerSetInfo) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.ID.Zero()) {
		return false
	}
	if !(s.Title == "") {
		return false
	}
	if !(s.Name == "") {
		return false
	}
	if !(s.Thumbnail.Zero()) {
		return false
	}
	if !(s.ThumbnailOutline == nil) {
		return false
	}
	if !(s.IsInstalled == false) {
		return false
	}
	if !(s.IsArchived == false) {
		return false
	}
	if !(s.IsOfficial == false) {
		return false
	}
	if !(s.IsAnimated == false) {
		return false
	}
	if !(s.IsMasks == false) {
		return false
	}
	if !(s.IsViewed == false) {
		return false
	}
	if !(s.Size == 0) {
		return false
	}
	if !(s.Covers == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *StickerSetInfo) String() string {
	if s == nil {
		return "StickerSetInfo(nil)"
	}
	type Alias StickerSetInfo
	return fmt.Sprintf("StickerSetInfo%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*StickerSetInfo) TypeID() uint32 {
	return StickerSetInfoTypeID
}

// TypeName returns name of type in TL schema.
func (*StickerSetInfo) TypeName() string {
	return "stickerSetInfo"
}

// TypeInfo returns info about TL type.
func (s *StickerSetInfo) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "stickerSetInfo",
		ID:   StickerSetInfoTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ID",
			SchemaName: "id",
		},
		{
			Name:       "Title",
			SchemaName: "title",
		},
		{
			Name:       "Name",
			SchemaName: "name",
		},
		{
			Name:       "Thumbnail",
			SchemaName: "thumbnail",
		},
		{
			Name:       "ThumbnailOutline",
			SchemaName: "thumbnail_outline",
		},
		{
			Name:       "IsInstalled",
			SchemaName: "is_installed",
		},
		{
			Name:       "IsArchived",
			SchemaName: "is_archived",
		},
		{
			Name:       "IsOfficial",
			SchemaName: "is_official",
		},
		{
			Name:       "IsAnimated",
			SchemaName: "is_animated",
		},
		{
			Name:       "IsMasks",
			SchemaName: "is_masks",
		},
		{
			Name:       "IsViewed",
			SchemaName: "is_viewed",
		},
		{
			Name:       "Size",
			SchemaName: "size",
		},
		{
			Name:       "Covers",
			SchemaName: "covers",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *StickerSetInfo) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode stickerSetInfo#aba733ac as nil")
	}
	b.PutID(StickerSetInfoTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *StickerSetInfo) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode stickerSetInfo#aba733ac as nil")
	}
	if err := s.ID.Encode(b); err != nil {
		return fmt.Errorf("unable to encode stickerSetInfo#aba733ac: field id: %w", err)
	}
	b.PutString(s.Title)
	b.PutString(s.Name)
	if err := s.Thumbnail.Encode(b); err != nil {
		return fmt.Errorf("unable to encode stickerSetInfo#aba733ac: field thumbnail: %w", err)
	}
	b.PutInt(len(s.ThumbnailOutline))
	for idx, v := range s.ThumbnailOutline {
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare stickerSetInfo#aba733ac: field thumbnail_outline element with index %d: %w", idx, err)
		}
	}
	b.PutBool(s.IsInstalled)
	b.PutBool(s.IsArchived)
	b.PutBool(s.IsOfficial)
	b.PutBool(s.IsAnimated)
	b.PutBool(s.IsMasks)
	b.PutBool(s.IsViewed)
	b.PutInt32(s.Size)
	b.PutInt(len(s.Covers))
	for idx, v := range s.Covers {
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare stickerSetInfo#aba733ac: field covers element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (s *StickerSetInfo) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode stickerSetInfo#aba733ac to nil")
	}
	if err := b.ConsumeID(StickerSetInfoTypeID); err != nil {
		return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *StickerSetInfo) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode stickerSetInfo#aba733ac to nil")
	}
	{
		if err := s.ID.Decode(b); err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field id: %w", err)
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field title: %w", err)
		}
		s.Title = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field name: %w", err)
		}
		s.Name = value
	}
	{
		if err := s.Thumbnail.Decode(b); err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field thumbnail: %w", err)
		}
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field thumbnail_outline: %w", err)
		}

		if headerLen > 0 {
			s.ThumbnailOutline = make([]ClosedVectorPath, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value ClosedVectorPath
			if err := value.DecodeBare(b); err != nil {
				return fmt.Errorf("unable to decode bare stickerSetInfo#aba733ac: field thumbnail_outline: %w", err)
			}
			s.ThumbnailOutline = append(s.ThumbnailOutline, value)
		}
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field is_installed: %w", err)
		}
		s.IsInstalled = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field is_archived: %w", err)
		}
		s.IsArchived = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field is_official: %w", err)
		}
		s.IsOfficial = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field is_animated: %w", err)
		}
		s.IsAnimated = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field is_masks: %w", err)
		}
		s.IsMasks = value
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field is_viewed: %w", err)
		}
		s.IsViewed = value
	}
	{
		value, err := b.Int32()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field size: %w", err)
		}
		s.Size = value
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode stickerSetInfo#aba733ac: field covers: %w", err)
		}

		if headerLen > 0 {
			s.Covers = make([]Sticker, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value Sticker
			if err := value.DecodeBare(b); err != nil {
				return fmt.Errorf("unable to decode bare stickerSetInfo#aba733ac: field covers: %w", err)
			}
			s.Covers = append(s.Covers, value)
		}
	}
	return nil
}

// GetID returns value of ID field.
func (s *StickerSetInfo) GetID() (value Int64) {
	return s.ID
}

// GetTitle returns value of Title field.
func (s *StickerSetInfo) GetTitle() (value string) {
	return s.Title
}

// GetName returns value of Name field.
func (s *StickerSetInfo) GetName() (value string) {
	return s.Name
}

// GetThumbnail returns value of Thumbnail field.
func (s *StickerSetInfo) GetThumbnail() (value Thumbnail) {
	return s.Thumbnail
}

// GetThumbnailOutline returns value of ThumbnailOutline field.
func (s *StickerSetInfo) GetThumbnailOutline() (value []ClosedVectorPath) {
	return s.ThumbnailOutline
}

// GetIsInstalled returns value of IsInstalled field.
func (s *StickerSetInfo) GetIsInstalled() (value bool) {
	return s.IsInstalled
}

// GetIsArchived returns value of IsArchived field.
func (s *StickerSetInfo) GetIsArchived() (value bool) {
	return s.IsArchived
}

// GetIsOfficial returns value of IsOfficial field.
func (s *StickerSetInfo) GetIsOfficial() (value bool) {
	return s.IsOfficial
}

// GetIsAnimated returns value of IsAnimated field.
func (s *StickerSetInfo) GetIsAnimated() (value bool) {
	return s.IsAnimated
}

// GetIsMasks returns value of IsMasks field.
func (s *StickerSetInfo) GetIsMasks() (value bool) {
	return s.IsMasks
}

// GetIsViewed returns value of IsViewed field.
func (s *StickerSetInfo) GetIsViewed() (value bool) {
	return s.IsViewed
}

// GetSize returns value of Size field.
func (s *StickerSetInfo) GetSize() (value int32) {
	return s.Size
}

// GetCovers returns value of Covers field.
func (s *StickerSetInfo) GetCovers() (value []Sticker) {
	return s.Covers
}