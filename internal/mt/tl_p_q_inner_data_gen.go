// Code generated by gotdgen, DO NOT EDIT.

package mt

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

// PQInnerData represents TL type `p_q_inner_data#83c95aec`.
type PQInnerData struct {
	// Pq field of PQInnerData.
	Pq []byte
	// P field of PQInnerData.
	P []byte
	// Q field of PQInnerData.
	Q []byte
	// Nonce field of PQInnerData.
	Nonce bin.Int128
	// ServerNonce field of PQInnerData.
	ServerNonce bin.Int128
	// NewNonce field of PQInnerData.
	NewNonce bin.Int256
}

// PQInnerDataTypeID is TL type id of PQInnerData.
const PQInnerDataTypeID = 0x83c95aec

func (p *PQInnerData) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.Pq == nil) {
		return false
	}
	if !(p.P == nil) {
		return false
	}
	if !(p.Q == nil) {
		return false
	}
	if !(p.Nonce == bin.Int128{}) {
		return false
	}
	if !(p.ServerNonce == bin.Int128{}) {
		return false
	}
	if !(p.NewNonce == bin.Int256{}) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *PQInnerData) String() string {
	if p == nil {
		return "PQInnerData(nil)"
	}
	type Alias PQInnerData
	return fmt.Sprintf("PQInnerData%+v", Alias(*p))
}

// FillFrom fills PQInnerData from given interface.
func (p *PQInnerData) FillFrom(from interface {
	GetPq() (value []byte)
	GetP() (value []byte)
	GetQ() (value []byte)
	GetNonce() (value bin.Int128)
	GetServerNonce() (value bin.Int128)
	GetNewNonce() (value bin.Int256)
}) {
	p.Pq = from.GetPq()
	p.P = from.GetP()
	p.Q = from.GetQ()
	p.Nonce = from.GetNonce()
	p.ServerNonce = from.GetServerNonce()
	p.NewNonce = from.GetNewNonce()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PQInnerData) TypeID() uint32 {
	return PQInnerDataTypeID
}

// TypeName returns name of type in TL schema.
func (*PQInnerData) TypeName() string {
	return "p_q_inner_data"
}

// TypeInfo returns info about TL type.
func (p *PQInnerData) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "p_q_inner_data",
		ID:   PQInnerDataTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Pq",
			SchemaName: "pq",
		},
		{
			Name:       "P",
			SchemaName: "p",
		},
		{
			Name:       "Q",
			SchemaName: "q",
		},
		{
			Name:       "Nonce",
			SchemaName: "nonce",
		},
		{
			Name:       "ServerNonce",
			SchemaName: "server_nonce",
		},
		{
			Name:       "NewNonce",
			SchemaName: "new_nonce",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *PQInnerData) Encode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "p_q_inner_data#83c95aec",
		}
	}
	b.PutID(PQInnerDataTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *PQInnerData) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "p_q_inner_data#83c95aec",
		}
	}
	b.PutBytes(p.Pq)
	b.PutBytes(p.P)
	b.PutBytes(p.Q)
	b.PutInt128(p.Nonce)
	b.PutInt128(p.ServerNonce)
	b.PutInt256(p.NewNonce)
	return nil
}

// GetPq returns value of Pq field.
func (p *PQInnerData) GetPq() (value []byte) {
	return p.Pq
}

// GetP returns value of P field.
func (p *PQInnerData) GetP() (value []byte) {
	return p.P
}

// GetQ returns value of Q field.
func (p *PQInnerData) GetQ() (value []byte) {
	return p.Q
}

// GetNonce returns value of Nonce field.
func (p *PQInnerData) GetNonce() (value bin.Int128) {
	return p.Nonce
}

// GetServerNonce returns value of ServerNonce field.
func (p *PQInnerData) GetServerNonce() (value bin.Int128) {
	return p.ServerNonce
}

// GetNewNonce returns value of NewNonce field.
func (p *PQInnerData) GetNewNonce() (value bin.Int256) {
	return p.NewNonce
}

// Decode implements bin.Decoder.
func (p *PQInnerData) Decode(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "p_q_inner_data#83c95aec",
		}
	}
	if err := b.ConsumeID(PQInnerDataTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "p_q_inner_data#83c95aec",
			Underlying: err,
		}
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *PQInnerData) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "p_q_inner_data#83c95aec",
		}
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "p_q_inner_data#83c95aec",
				FieldName:  "pq",
				Underlying: err,
			}
		}
		p.Pq = value
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "p_q_inner_data#83c95aec",
				FieldName:  "p",
				Underlying: err,
			}
		}
		p.P = value
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "p_q_inner_data#83c95aec",
				FieldName:  "q",
				Underlying: err,
			}
		}
		p.Q = value
	}
	{
		value, err := b.Int128()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "p_q_inner_data#83c95aec",
				FieldName:  "nonce",
				Underlying: err,
			}
		}
		p.Nonce = value
	}
	{
		value, err := b.Int128()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "p_q_inner_data#83c95aec",
				FieldName:  "server_nonce",
				Underlying: err,
			}
		}
		p.ServerNonce = value
	}
	{
		value, err := b.Int256()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "p_q_inner_data#83c95aec",
				FieldName:  "new_nonce",
				Underlying: err,
			}
		}
		p.NewNonce = value
	}
	return nil
}

// Ensuring interfaces in compile-time for PQInnerData.
var (
	_ bin.Encoder     = &PQInnerData{}
	_ bin.Decoder     = &PQInnerData{}
	_ bin.BareEncoder = &PQInnerData{}
	_ bin.BareDecoder = &PQInnerData{}
)
