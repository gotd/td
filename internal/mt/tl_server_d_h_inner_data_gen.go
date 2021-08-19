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

// ServerDHInnerData represents TL type `server_DH_inner_data#b5890dba`.
type ServerDHInnerData struct {
	// Nonce field of ServerDHInnerData.
	Nonce bin.Int128
	// ServerNonce field of ServerDHInnerData.
	ServerNonce bin.Int128
	// G field of ServerDHInnerData.
	G int
	// DhPrime field of ServerDHInnerData.
	DhPrime []byte
	// GA field of ServerDHInnerData.
	GA []byte
	// ServerTime field of ServerDHInnerData.
	ServerTime int
}

// ServerDHInnerDataTypeID is TL type id of ServerDHInnerData.
const ServerDHInnerDataTypeID = 0xb5890dba

func (s *ServerDHInnerData) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Nonce == bin.Int128{}) {
		return false
	}
	if !(s.ServerNonce == bin.Int128{}) {
		return false
	}
	if !(s.G == 0) {
		return false
	}
	if !(s.DhPrime == nil) {
		return false
	}
	if !(s.GA == nil) {
		return false
	}
	if !(s.ServerTime == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *ServerDHInnerData) String() string {
	if s == nil {
		return "ServerDHInnerData(nil)"
	}
	type Alias ServerDHInnerData
	return fmt.Sprintf("ServerDHInnerData%+v", Alias(*s))
}

// FillFrom fills ServerDHInnerData from given interface.
func (s *ServerDHInnerData) FillFrom(from interface {
	GetNonce() (value bin.Int128)
	GetServerNonce() (value bin.Int128)
	GetG() (value int)
	GetDhPrime() (value []byte)
	GetGA() (value []byte)
	GetServerTime() (value int)
}) {
	s.Nonce = from.GetNonce()
	s.ServerNonce = from.GetServerNonce()
	s.G = from.GetG()
	s.DhPrime = from.GetDhPrime()
	s.GA = from.GetGA()
	s.ServerTime = from.GetServerTime()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ServerDHInnerData) TypeID() uint32 {
	return ServerDHInnerDataTypeID
}

// TypeName returns name of type in TL schema.
func (*ServerDHInnerData) TypeName() string {
	return "server_DH_inner_data"
}

// TypeInfo returns info about TL type.
func (s *ServerDHInnerData) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "server_DH_inner_data",
		ID:   ServerDHInnerDataTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Nonce",
			SchemaName: "nonce",
		},
		{
			Name:       "ServerNonce",
			SchemaName: "server_nonce",
		},
		{
			Name:       "G",
			SchemaName: "g",
		},
		{
			Name:       "DhPrime",
			SchemaName: "dh_prime",
		},
		{
			Name:       "GA",
			SchemaName: "g_a",
		},
		{
			Name:       "ServerTime",
			SchemaName: "server_time",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *ServerDHInnerData) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "server_DH_inner_data#b5890dba",
		}
	}
	b.PutID(ServerDHInnerDataTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *ServerDHInnerData) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "server_DH_inner_data#b5890dba",
		}
	}
	b.PutInt128(s.Nonce)
	b.PutInt128(s.ServerNonce)
	b.PutInt(s.G)
	b.PutBytes(s.DhPrime)
	b.PutBytes(s.GA)
	b.PutInt(s.ServerTime)
	return nil
}

// GetNonce returns value of Nonce field.
func (s *ServerDHInnerData) GetNonce() (value bin.Int128) {
	return s.Nonce
}

// GetServerNonce returns value of ServerNonce field.
func (s *ServerDHInnerData) GetServerNonce() (value bin.Int128) {
	return s.ServerNonce
}

// GetG returns value of G field.
func (s *ServerDHInnerData) GetG() (value int) {
	return s.G
}

// GetDhPrime returns value of DhPrime field.
func (s *ServerDHInnerData) GetDhPrime() (value []byte) {
	return s.DhPrime
}

// GetGA returns value of GA field.
func (s *ServerDHInnerData) GetGA() (value []byte) {
	return s.GA
}

// GetServerTime returns value of ServerTime field.
func (s *ServerDHInnerData) GetServerTime() (value int) {
	return s.ServerTime
}

// Decode implements bin.Decoder.
func (s *ServerDHInnerData) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "server_DH_inner_data#b5890dba",
		}
	}
	if err := b.ConsumeID(ServerDHInnerDataTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "server_DH_inner_data#b5890dba",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *ServerDHInnerData) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "server_DH_inner_data#b5890dba",
		}
	}
	{
		value, err := b.Int128()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "server_DH_inner_data#b5890dba",
				FieldName:  "nonce",
				Underlying: err,
			}
		}
		s.Nonce = value
	}
	{
		value, err := b.Int128()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "server_DH_inner_data#b5890dba",
				FieldName:  "server_nonce",
				Underlying: err,
			}
		}
		s.ServerNonce = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "server_DH_inner_data#b5890dba",
				FieldName:  "g",
				Underlying: err,
			}
		}
		s.G = value
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "server_DH_inner_data#b5890dba",
				FieldName:  "dh_prime",
				Underlying: err,
			}
		}
		s.DhPrime = value
	}
	{
		value, err := b.Bytes()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "server_DH_inner_data#b5890dba",
				FieldName:  "g_a",
				Underlying: err,
			}
		}
		s.GA = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "server_DH_inner_data#b5890dba",
				FieldName:  "server_time",
				Underlying: err,
			}
		}
		s.ServerTime = value
	}
	return nil
}

// Ensuring interfaces in compile-time for ServerDHInnerData.
var (
	_ bin.Encoder     = &ServerDHInnerData{}
	_ bin.Decoder     = &ServerDHInnerData{}
	_ bin.BareEncoder = &ServerDHInnerData{}
	_ bin.BareDecoder = &ServerDHInnerData{}
)
