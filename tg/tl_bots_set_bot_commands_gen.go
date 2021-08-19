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

// BotsSetBotCommandsRequest represents TL type `bots.setBotCommands#517165a`.
// Set bot command list
//
// See https://core.telegram.org/method/bots.setBotCommands for reference.
type BotsSetBotCommandsRequest struct {
	// Scope field of BotsSetBotCommandsRequest.
	Scope BotCommandScopeClass
	// LangCode field of BotsSetBotCommandsRequest.
	LangCode string
	// Bot commands
	Commands []BotCommand
}

// BotsSetBotCommandsRequestTypeID is TL type id of BotsSetBotCommandsRequest.
const BotsSetBotCommandsRequestTypeID = 0x517165a

func (s *BotsSetBotCommandsRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Scope == nil) {
		return false
	}
	if !(s.LangCode == "") {
		return false
	}
	if !(s.Commands == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *BotsSetBotCommandsRequest) String() string {
	if s == nil {
		return "BotsSetBotCommandsRequest(nil)"
	}
	type Alias BotsSetBotCommandsRequest
	return fmt.Sprintf("BotsSetBotCommandsRequest%+v", Alias(*s))
}

// FillFrom fills BotsSetBotCommandsRequest from given interface.
func (s *BotsSetBotCommandsRequest) FillFrom(from interface {
	GetScope() (value BotCommandScopeClass)
	GetLangCode() (value string)
	GetCommands() (value []BotCommand)
}) {
	s.Scope = from.GetScope()
	s.LangCode = from.GetLangCode()
	s.Commands = from.GetCommands()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*BotsSetBotCommandsRequest) TypeID() uint32 {
	return BotsSetBotCommandsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*BotsSetBotCommandsRequest) TypeName() string {
	return "bots.setBotCommands"
}

// TypeInfo returns info about TL type.
func (s *BotsSetBotCommandsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "bots.setBotCommands",
		ID:   BotsSetBotCommandsRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Scope",
			SchemaName: "scope",
		},
		{
			Name:       "LangCode",
			SchemaName: "lang_code",
		},
		{
			Name:       "Commands",
			SchemaName: "commands",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *BotsSetBotCommandsRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "bots.setBotCommands#517165a",
		}
	}
	b.PutID(BotsSetBotCommandsRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *BotsSetBotCommandsRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "bots.setBotCommands#517165a",
		}
	}
	if s.Scope == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "bots.setBotCommands#517165a",
			FieldName: "scope",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "BotCommandScope",
			},
		}
	}
	if err := s.Scope.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "bots.setBotCommands#517165a",
			FieldName:  "scope",
			Underlying: err,
		}
	}
	b.PutString(s.LangCode)
	b.PutVectorHeader(len(s.Commands))
	for idx, v := range s.Commands {
		if err := v.Encode(b); err != nil {
			return &bin.FieldError{
				Action:    "encode",
				TypeName:  "bots.setBotCommands#517165a",
				FieldName: "commands",
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

// GetScope returns value of Scope field.
func (s *BotsSetBotCommandsRequest) GetScope() (value BotCommandScopeClass) {
	return s.Scope
}

// GetLangCode returns value of LangCode field.
func (s *BotsSetBotCommandsRequest) GetLangCode() (value string) {
	return s.LangCode
}

// GetCommands returns value of Commands field.
func (s *BotsSetBotCommandsRequest) GetCommands() (value []BotCommand) {
	return s.Commands
}

// Decode implements bin.Decoder.
func (s *BotsSetBotCommandsRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "bots.setBotCommands#517165a",
		}
	}
	if err := b.ConsumeID(BotsSetBotCommandsRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "bots.setBotCommands#517165a",
			Underlying: err,
		}
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *BotsSetBotCommandsRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "bots.setBotCommands#517165a",
		}
	}
	{
		value, err := DecodeBotCommandScope(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "bots.setBotCommands#517165a",
				FieldName:  "scope",
				Underlying: err,
			}
		}
		s.Scope = value
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "bots.setBotCommands#517165a",
				FieldName:  "lang_code",
				Underlying: err,
			}
		}
		s.LangCode = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "bots.setBotCommands#517165a",
				FieldName:  "commands",
				Underlying: err,
			}
		}

		if headerLen > 0 {
			s.Commands = make([]BotCommand, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value BotCommand
			if err := value.Decode(b); err != nil {
				return &bin.FieldError{
					Action:     "decode",
					BareField:  false,
					TypeName:   "bots.setBotCommands#517165a",
					FieldName:  "commands",
					Underlying: err,
				}
			}
			s.Commands = append(s.Commands, value)
		}
	}
	return nil
}

// Ensuring interfaces in compile-time for BotsSetBotCommandsRequest.
var (
	_ bin.Encoder     = &BotsSetBotCommandsRequest{}
	_ bin.Decoder     = &BotsSetBotCommandsRequest{}
	_ bin.BareEncoder = &BotsSetBotCommandsRequest{}
	_ bin.BareDecoder = &BotsSetBotCommandsRequest{}
)

// BotsSetBotCommands invokes method bots.setBotCommands#517165a returning error if any.
// Set bot command list
//
// See https://core.telegram.org/method/bots.setBotCommands for reference.
// Can be used by bots.
func (c *Client) BotsSetBotCommands(ctx context.Context, request *BotsSetBotCommandsRequest) (bool, error) {
	var result BoolBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return false, err
	}
	_, ok := result.Bool.(*BoolTrue)
	return ok, nil
}
