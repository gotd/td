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

// StarTransactions represents TL type `starTransactions#b5f5820b`.
type StarTransactions struct {
	// The amount of owned Telegram Stars
	StarAmount StarAmount
	// List of transactions with Telegram Stars
	Transactions []StarTransaction
	// The offset for the next request. If empty, then there are no more results
	NextOffset string
}

// StarTransactionsTypeID is TL type id of StarTransactions.
const StarTransactionsTypeID = 0xb5f5820b

// Ensuring interfaces in compile-time for StarTransactions.
var (
	_ bin.Encoder     = &StarTransactions{}
	_ bin.Decoder     = &StarTransactions{}
	_ bin.BareEncoder = &StarTransactions{}
	_ bin.BareDecoder = &StarTransactions{}
)

func (s *StarTransactions) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.StarAmount.Zero()) {
		return false
	}
	if !(s.Transactions == nil) {
		return false
	}
	if !(s.NextOffset == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *StarTransactions) String() string {
	if s == nil {
		return "StarTransactions(nil)"
	}
	type Alias StarTransactions
	return fmt.Sprintf("StarTransactions%+v", Alias(*s))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*StarTransactions) TypeID() uint32 {
	return StarTransactionsTypeID
}

// TypeName returns name of type in TL schema.
func (*StarTransactions) TypeName() string {
	return "starTransactions"
}

// TypeInfo returns info about TL type.
func (s *StarTransactions) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "starTransactions",
		ID:   StarTransactionsTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "StarAmount",
			SchemaName: "star_amount",
		},
		{
			Name:       "Transactions",
			SchemaName: "transactions",
		},
		{
			Name:       "NextOffset",
			SchemaName: "next_offset",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *StarTransactions) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode starTransactions#b5f5820b as nil")
	}
	b.PutID(StarTransactionsTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *StarTransactions) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode starTransactions#b5f5820b as nil")
	}
	if err := s.StarAmount.Encode(b); err != nil {
		return fmt.Errorf("unable to encode starTransactions#b5f5820b: field star_amount: %w", err)
	}
	b.PutInt(len(s.Transactions))
	for idx, v := range s.Transactions {
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare starTransactions#b5f5820b: field transactions element with index %d: %w", idx, err)
		}
	}
	b.PutString(s.NextOffset)
	return nil
}

// Decode implements bin.Decoder.
func (s *StarTransactions) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode starTransactions#b5f5820b to nil")
	}
	if err := b.ConsumeID(StarTransactionsTypeID); err != nil {
		return fmt.Errorf("unable to decode starTransactions#b5f5820b: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *StarTransactions) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode starTransactions#b5f5820b to nil")
	}
	{
		if err := s.StarAmount.Decode(b); err != nil {
			return fmt.Errorf("unable to decode starTransactions#b5f5820b: field star_amount: %w", err)
		}
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode starTransactions#b5f5820b: field transactions: %w", err)
		}

		if headerLen > 0 {
			s.Transactions = make([]StarTransaction, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value StarTransaction
			if err := value.DecodeBare(b); err != nil {
				return fmt.Errorf("unable to decode bare starTransactions#b5f5820b: field transactions: %w", err)
			}
			s.Transactions = append(s.Transactions, value)
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode starTransactions#b5f5820b: field next_offset: %w", err)
		}
		s.NextOffset = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (s *StarTransactions) EncodeTDLibJSON(b tdjson.Encoder) error {
	if s == nil {
		return fmt.Errorf("can't encode starTransactions#b5f5820b as nil")
	}
	b.ObjStart()
	b.PutID("starTransactions")
	b.Comma()
	b.FieldStart("star_amount")
	if err := s.StarAmount.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode starTransactions#b5f5820b: field star_amount: %w", err)
	}
	b.Comma()
	b.FieldStart("transactions")
	b.ArrStart()
	for idx, v := range s.Transactions {
		if err := v.EncodeTDLibJSON(b); err != nil {
			return fmt.Errorf("unable to encode starTransactions#b5f5820b: field transactions element with index %d: %w", idx, err)
		}
		b.Comma()
	}
	b.StripComma()
	b.ArrEnd()
	b.Comma()
	b.FieldStart("next_offset")
	b.PutString(s.NextOffset)
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (s *StarTransactions) DecodeTDLibJSON(b tdjson.Decoder) error {
	if s == nil {
		return fmt.Errorf("can't decode starTransactions#b5f5820b to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("starTransactions"); err != nil {
				return fmt.Errorf("unable to decode starTransactions#b5f5820b: %w", err)
			}
		case "star_amount":
			if err := s.StarAmount.DecodeTDLibJSON(b); err != nil {
				return fmt.Errorf("unable to decode starTransactions#b5f5820b: field star_amount: %w", err)
			}
		case "transactions":
			if err := b.Arr(func(b tdjson.Decoder) error {
				var value StarTransaction
				if err := value.DecodeTDLibJSON(b); err != nil {
					return fmt.Errorf("unable to decode starTransactions#b5f5820b: field transactions: %w", err)
				}
				s.Transactions = append(s.Transactions, value)
				return nil
			}); err != nil {
				return fmt.Errorf("unable to decode starTransactions#b5f5820b: field transactions: %w", err)
			}
		case "next_offset":
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode starTransactions#b5f5820b: field next_offset: %w", err)
			}
			s.NextOffset = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetStarAmount returns value of StarAmount field.
func (s *StarTransactions) GetStarAmount() (value StarAmount) {
	if s == nil {
		return
	}
	return s.StarAmount
}

// GetTransactions returns value of Transactions field.
func (s *StarTransactions) GetTransactions() (value []StarTransaction) {
	if s == nil {
		return
	}
	return s.Transactions
}

// GetNextOffset returns value of NextOffset field.
func (s *StarTransactions) GetNextOffset() (value string) {
	if s == nil {
		return
	}
	return s.NextOffset
}
