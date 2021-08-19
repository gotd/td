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

// MessagesUploadEncryptedFileRequest represents TL type `messages.uploadEncryptedFile#5057c497`.
// Upload encrypted file and associate it to a secret chat
//
// See https://core.telegram.org/method/messages.uploadEncryptedFile for reference.
type MessagesUploadEncryptedFileRequest struct {
	// The secret chat to associate the file to
	Peer InputEncryptedChat
	// The file
	File InputEncryptedFileClass
}

// MessagesUploadEncryptedFileRequestTypeID is TL type id of MessagesUploadEncryptedFileRequest.
const MessagesUploadEncryptedFileRequestTypeID = 0x5057c497

func (u *MessagesUploadEncryptedFileRequest) Zero() bool {
	if u == nil {
		return true
	}
	if !(u.Peer.Zero()) {
		return false
	}
	if !(u.File == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (u *MessagesUploadEncryptedFileRequest) String() string {
	if u == nil {
		return "MessagesUploadEncryptedFileRequest(nil)"
	}
	type Alias MessagesUploadEncryptedFileRequest
	return fmt.Sprintf("MessagesUploadEncryptedFileRequest%+v", Alias(*u))
}

// FillFrom fills MessagesUploadEncryptedFileRequest from given interface.
func (u *MessagesUploadEncryptedFileRequest) FillFrom(from interface {
	GetPeer() (value InputEncryptedChat)
	GetFile() (value InputEncryptedFileClass)
}) {
	u.Peer = from.GetPeer()
	u.File = from.GetFile()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesUploadEncryptedFileRequest) TypeID() uint32 {
	return MessagesUploadEncryptedFileRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesUploadEncryptedFileRequest) TypeName() string {
	return "messages.uploadEncryptedFile"
}

// TypeInfo returns info about TL type.
func (u *MessagesUploadEncryptedFileRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.uploadEncryptedFile",
		ID:   MessagesUploadEncryptedFileRequestTypeID,
	}
	if u == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
		{
			Name:       "File",
			SchemaName: "file",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (u *MessagesUploadEncryptedFileRequest) Encode(b *bin.Buffer) error {
	if u == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.uploadEncryptedFile#5057c497",
		}
	}
	b.PutID(MessagesUploadEncryptedFileRequestTypeID)
	return u.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (u *MessagesUploadEncryptedFileRequest) EncodeBare(b *bin.Buffer) error {
	if u == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "messages.uploadEncryptedFile#5057c497",
		}
	}
	if err := u.Peer.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.uploadEncryptedFile#5057c497",
			FieldName:  "peer",
			Underlying: err,
		}
	}
	if u.File == nil {
		return &bin.FieldError{
			Action:    "encode",
			TypeName:  "messages.uploadEncryptedFile#5057c497",
			FieldName: "file",
			Underlying: &bin.NilError{
				Action:   "encode",
				TypeName: "InputEncryptedFile",
			},
		}
	}
	if err := u.File.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "messages.uploadEncryptedFile#5057c497",
			FieldName:  "file",
			Underlying: err,
		}
	}
	return nil
}

// GetPeer returns value of Peer field.
func (u *MessagesUploadEncryptedFileRequest) GetPeer() (value InputEncryptedChat) {
	return u.Peer
}

// GetFile returns value of File field.
func (u *MessagesUploadEncryptedFileRequest) GetFile() (value InputEncryptedFileClass) {
	return u.File
}

// GetFileAsNotEmpty returns mapped value of File field.
func (u *MessagesUploadEncryptedFileRequest) GetFileAsNotEmpty() (NotEmptyInputEncryptedFile, bool) {
	return u.File.AsNotEmpty()
}

// Decode implements bin.Decoder.
func (u *MessagesUploadEncryptedFileRequest) Decode(b *bin.Buffer) error {
	if u == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.uploadEncryptedFile#5057c497",
		}
	}
	if err := b.ConsumeID(MessagesUploadEncryptedFileRequestTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "messages.uploadEncryptedFile#5057c497",
			Underlying: err,
		}
	}
	return u.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (u *MessagesUploadEncryptedFileRequest) DecodeBare(b *bin.Buffer) error {
	if u == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "messages.uploadEncryptedFile#5057c497",
		}
	}
	{
		if err := u.Peer.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.uploadEncryptedFile#5057c497",
				FieldName:  "peer",
				Underlying: err,
			}
		}
	}
	{
		value, err := DecodeInputEncryptedFile(b)
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "messages.uploadEncryptedFile#5057c497",
				FieldName:  "file",
				Underlying: err,
			}
		}
		u.File = value
	}
	return nil
}

// Ensuring interfaces in compile-time for MessagesUploadEncryptedFileRequest.
var (
	_ bin.Encoder     = &MessagesUploadEncryptedFileRequest{}
	_ bin.Decoder     = &MessagesUploadEncryptedFileRequest{}
	_ bin.BareEncoder = &MessagesUploadEncryptedFileRequest{}
	_ bin.BareDecoder = &MessagesUploadEncryptedFileRequest{}
)

// MessagesUploadEncryptedFile invokes method messages.uploadEncryptedFile#5057c497 returning error if any.
// Upload encrypted file and associate it to a secret chat
//
// See https://core.telegram.org/method/messages.uploadEncryptedFile for reference.
func (c *Client) MessagesUploadEncryptedFile(ctx context.Context, request *MessagesUploadEncryptedFileRequest) (EncryptedFileClass, error) {
	var result EncryptedFileBox

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.EncryptedFile, nil
}
