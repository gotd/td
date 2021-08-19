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

// MessagesHighScores represents TL type `messages.highScores#9a3bfd99`.
// Highscores in a game
//
// See https://core.telegram.org/constructor/messages.highScores for reference.
type MessagesHighScores struct {
	// Highscores
	Scores []HighScore
	// Users, associated to the highscores
	Users []UserClass
}

// MessagesHighScoresTypeID is TL type id of MessagesHighScores.
const MessagesHighScoresTypeID = 0x9a3bfd99

func (h *MessagesHighScores) Zero() bool {
	if h == nil {
		return true
	}
	if !(h.Scores == nil) {
		return false
	}
	if !(h.Users == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (h *MessagesHighScores) String() string {
	if h == nil {
		return "MessagesHighScores(nil)"
	}
	type Alias MessagesHighScores
	return fmt.Sprintf("MessagesHighScores%+v", Alias(*h))
}

// FillFrom fills MessagesHighScores from given interface.
func (h *MessagesHighScores) FillFrom(from interface {
	GetScores() (value []HighScore)
	GetUsers() (value []UserClass)
}) {
	h.Scores = from.GetScores()
	h.Users = from.GetUsers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesHighScores) TypeID() uint32 {
	return MessagesHighScoresTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesHighScores) TypeName() string {
	return "messages.highScores"
}

// TypeInfo returns info about TL type.
func (h *MessagesHighScores) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.highScores",
		ID:   MessagesHighScoresTypeID,
	}
	if h == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Scores",
			SchemaName: "scores",
		},
		{
			Name:       "Users",
			SchemaName: "users",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (h *MessagesHighScores) Encode(b *bin.Buffer) error {
	if h == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.highScores#9a3bfd99",
		}
	}
	b.PutID(MessagesHighScoresTypeID)
	return h.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (h *MessagesHighScores) EncodeBare(b *bin.Buffer) error {
	if h == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.highScores#9a3bfd99",
		}
	}
	b.PutVectorHeader(len(h.Scores))
	for idx, v := range h.Scores {
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.highScores#9a3bfd99",
				FieldName: "scores",
				BareField: false,
				Underlying: &bin.IndexError{
					Index:      idx,
					Underlying: err,
				},
			}
		}
	}
	b.PutVectorHeader(len(h.Users))
	for idx, v := range h.Users {
		if v == nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.highScores#9a3bfd99",
				FieldName: "users",
				Underlying: &bin.IndexError{
					Index: idx,
					Underlying: &bin.NilError{
						Action:   "encode",
						TypeName: "Vector<User>",
					},
				},
			}
		}
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "messages.highScores#9a3bfd99",
				FieldName: "users",
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

// GetScores returns value of Scores field.
func (h *MessagesHighScores) GetScores() (value []HighScore) {
	return h.Scores
}

// GetUsers returns value of Users field.
func (h *MessagesHighScores) GetUsers() (value []UserClass) {
	return h.Users
}

// MapUsers returns field Users wrapped in UserClassArray helper.
func (h *MessagesHighScores) MapUsers() (value UserClassArray) {
	return UserClassArray(h.Users)
}

// Decode implements bin.Decoder.
func (h *MessagesHighScores) Decode(b *bin.Buffer) error {
	if h == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.highScores#9a3bfd99",
		}
	}
	if err := b.ConsumeID(MessagesHighScoresTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.highScores#9a3bfd99",
			Underlying: err,
		}
	}
	return h.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (h *MessagesHighScores) DecodeBare(b *bin.Buffer) error {
	if h == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.highScores#9a3bfd99",
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.highScores#9a3bfd99",
				FieldName:  "scores",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			h.Scores = make([]HighScore, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value HighScore
			if err := value.Decode(b); err != nil {
				return &bin.FieldError{
					Action:     "decode",
					BareField:  false,
					TypeName:   "messages.highScores#9a3bfd99",
					FieldName:  "scores",
					Underlying: err,
				}
			}
			h.Scores = append(h.Scores, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.highScores#9a3bfd99",
				FieldName:  "users",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			h.Users = make([]UserClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeUser(b)
			if err != nil {
				return &bin.FieldError{
					Action:     "decode",
					TypeName:   "messages.highScores#9a3bfd99",
					FieldName:  "users",
					Underlying: err,
				}
			}
			h.Users = append(h.Users, value)
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesHighScores.
var (
	_ bin.Encoder     = &MessagesHighScores{}
	_ bin.Decoder     = &MessagesHighScores{}
	_ bin.BareEncoder = &MessagesHighScores{}
	_ bin.BareDecoder = &MessagesHighScores{}
)
