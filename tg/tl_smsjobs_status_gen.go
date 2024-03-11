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

// SMSJobsStatus represents TL type `smsjobs.status#2aee9191`.
//
// See https://core.telegram.org/constructor/smsjobs.status for reference.
type SMSJobsStatus struct {
	// Flags field of SMSJobsStatus.
	Flags bin.Fields
	// AllowInternational field of SMSJobsStatus.
	AllowInternational bool
	// RecentSent field of SMSJobsStatus.
	RecentSent int
	// RecentSince field of SMSJobsStatus.
	RecentSince int
	// RecentRemains field of SMSJobsStatus.
	RecentRemains int
	// TotalSent field of SMSJobsStatus.
	TotalSent int
	// TotalSince field of SMSJobsStatus.
	TotalSince int
	// LastGiftSlug field of SMSJobsStatus.
	//
	// Use SetLastGiftSlug and GetLastGiftSlug helpers.
	LastGiftSlug string
	// TermsURL field of SMSJobsStatus.
	TermsURL string
}

// SMSJobsStatusTypeID is TL type id of SMSJobsStatus.
const SMSJobsStatusTypeID = 0x2aee9191

// Ensuring interfaces in compile-time for SMSJobsStatus.
var (
	_ bin.Encoder     = &SMSJobsStatus{}
	_ bin.Decoder     = &SMSJobsStatus{}
	_ bin.BareEncoder = &SMSJobsStatus{}
	_ bin.BareDecoder = &SMSJobsStatus{}
)

func (s *SMSJobsStatus) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Flags.Zero()) {
		return false
	}
	if !(s.AllowInternational == false) {
		return false
	}
	if !(s.RecentSent == 0) {
		return false
	}
	if !(s.RecentSince == 0) {
		return false
	}
	if !(s.RecentRemains == 0) {
		return false
	}
	if !(s.TotalSent == 0) {
		return false
	}
	if !(s.TotalSince == 0) {
		return false
	}
	if !(s.LastGiftSlug == "") {
		return false
	}
	if !(s.TermsURL == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *SMSJobsStatus) String() string {
	if s == nil {
		return "SMSJobsStatus(nil)"
	}
	type Alias SMSJobsStatus
	return fmt.Sprintf("SMSJobsStatus%+v", Alias(*s))
}

// FillFrom fills SMSJobsStatus from given interface.
func (s *SMSJobsStatus) FillFrom(from interface {
	GetAllowInternational() (value bool)
	GetRecentSent() (value int)
	GetRecentSince() (value int)
	GetRecentRemains() (value int)
	GetTotalSent() (value int)
	GetTotalSince() (value int)
	GetLastGiftSlug() (value string, ok bool)
	GetTermsURL() (value string)
}) {
	s.AllowInternational = from.GetAllowInternational()
	s.RecentSent = from.GetRecentSent()
	s.RecentSince = from.GetRecentSince()
	s.RecentRemains = from.GetRecentRemains()
	s.TotalSent = from.GetTotalSent()
	s.TotalSince = from.GetTotalSince()
	if val, ok := from.GetLastGiftSlug(); ok {
		s.LastGiftSlug = val
	}

	s.TermsURL = from.GetTermsURL()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*SMSJobsStatus) TypeID() uint32 {
	return SMSJobsStatusTypeID
}

// TypeName returns name of type in TL schema.
func (*SMSJobsStatus) TypeName() string {
	return "smsjobs.status"
}

// TypeInfo returns info about TL type.
func (s *SMSJobsStatus) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "smsjobs.status",
		ID:   SMSJobsStatusTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "AllowInternational",
			SchemaName: "allow_international",
			Null:       !s.Flags.Has(0),
		},
		{
			Name:       "RecentSent",
			SchemaName: "recent_sent",
		},
		{
			Name:       "RecentSince",
			SchemaName: "recent_since",
		},
		{
			Name:       "RecentRemains",
			SchemaName: "recent_remains",
		},
		{
			Name:       "TotalSent",
			SchemaName: "total_sent",
		},
		{
			Name:       "TotalSince",
			SchemaName: "total_since",
		},
		{
			Name:       "LastGiftSlug",
			SchemaName: "last_gift_slug",
			Null:       !s.Flags.Has(1),
		},
		{
			Name:       "TermsURL",
			SchemaName: "terms_url",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (s *SMSJobsStatus) SetFlags() {
	if !(s.AllowInternational == false) {
		s.Flags.Set(0)
	}
	if !(s.LastGiftSlug == "") {
		s.Flags.Set(1)
	}
}

// Encode implements bin.Encoder.
func (s *SMSJobsStatus) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode smsjobs.status#2aee9191 as nil")
	}
	b.PutID(SMSJobsStatusTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *SMSJobsStatus) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode smsjobs.status#2aee9191 as nil")
	}
	s.SetFlags()
	if err := s.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode smsjobs.status#2aee9191: field flags: %w", err)
	}
	b.PutInt(s.RecentSent)
	b.PutInt(s.RecentSince)
	b.PutInt(s.RecentRemains)
	b.PutInt(s.TotalSent)
	b.PutInt(s.TotalSince)
	if s.Flags.Has(1) {
		b.PutString(s.LastGiftSlug)
	}
	b.PutString(s.TermsURL)
	return nil
}

// Decode implements bin.Decoder.
func (s *SMSJobsStatus) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode smsjobs.status#2aee9191 to nil")
	}
	if err := b.ConsumeID(SMSJobsStatusTypeID); err != nil {
		return fmt.Errorf("unable to decode smsjobs.status#2aee9191: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *SMSJobsStatus) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode smsjobs.status#2aee9191 to nil")
	}
	{
		if err := s.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field flags: %w", err)
		}
	}
	s.AllowInternational = s.Flags.Has(0)
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field recent_sent: %w", err)
		}
		s.RecentSent = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field recent_since: %w", err)
		}
		s.RecentSince = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field recent_remains: %w", err)
		}
		s.RecentRemains = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field total_sent: %w", err)
		}
		s.TotalSent = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field total_since: %w", err)
		}
		s.TotalSince = value
	}
	if s.Flags.Has(1) {
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field last_gift_slug: %w", err)
		}
		s.LastGiftSlug = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode smsjobs.status#2aee9191: field terms_url: %w", err)
		}
		s.TermsURL = value
	}
	return nil
}

// SetAllowInternational sets value of AllowInternational conditional field.
func (s *SMSJobsStatus) SetAllowInternational(value bool) {
	if value {
		s.Flags.Set(0)
		s.AllowInternational = true
	} else {
		s.Flags.Unset(0)
		s.AllowInternational = false
	}
}

// GetAllowInternational returns value of AllowInternational conditional field.
func (s *SMSJobsStatus) GetAllowInternational() (value bool) {
	if s == nil {
		return
	}
	return s.Flags.Has(0)
}

// GetRecentSent returns value of RecentSent field.
func (s *SMSJobsStatus) GetRecentSent() (value int) {
	if s == nil {
		return
	}
	return s.RecentSent
}

// GetRecentSince returns value of RecentSince field.
func (s *SMSJobsStatus) GetRecentSince() (value int) {
	if s == nil {
		return
	}
	return s.RecentSince
}

// GetRecentRemains returns value of RecentRemains field.
func (s *SMSJobsStatus) GetRecentRemains() (value int) {
	if s == nil {
		return
	}
	return s.RecentRemains
}

// GetTotalSent returns value of TotalSent field.
func (s *SMSJobsStatus) GetTotalSent() (value int) {
	if s == nil {
		return
	}
	return s.TotalSent
}

// GetTotalSince returns value of TotalSince field.
func (s *SMSJobsStatus) GetTotalSince() (value int) {
	if s == nil {
		return
	}
	return s.TotalSince
}

// SetLastGiftSlug sets value of LastGiftSlug conditional field.
func (s *SMSJobsStatus) SetLastGiftSlug(value string) {
	s.Flags.Set(1)
	s.LastGiftSlug = value
}

// GetLastGiftSlug returns value of LastGiftSlug conditional field and
// boolean which is true if field was set.
func (s *SMSJobsStatus) GetLastGiftSlug() (value string, ok bool) {
	if s == nil {
		return
	}
	if !s.Flags.Has(1) {
		return value, false
	}
	return s.LastGiftSlug, true
}

// GetTermsURL returns value of TermsURL field.
func (s *SMSJobsStatus) GetTermsURL() (value string) {
	if s == nil {
		return
	}
	return s.TermsURL
}