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

// InputPersonalDocument represents TL type `inputPersonalDocument#bb343fae`.
type InputPersonalDocument struct {
	// List of files containing the pages of the document
	Files []InputFileClass
	// List of files containing a certified English translation of the document
	Translation []InputFileClass
}

// InputPersonalDocumentTypeID is TL type id of InputPersonalDocument.
const InputPersonalDocumentTypeID = 0xbb343fae

// Ensuring interfaces in compile-time for InputPersonalDocument.
var (
	_ bin.Encoder     = &InputPersonalDocument{}
	_ bin.Decoder     = &InputPersonalDocument{}
	_ bin.BareEncoder = &InputPersonalDocument{}
	_ bin.BareDecoder = &InputPersonalDocument{}
)

func (i *InputPersonalDocument) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Files == nil) {
		return false
	}
	if !(i.Translation == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InputPersonalDocument) String() string {
	if i == nil {
		return "InputPersonalDocument(nil)"
	}
	type Alias InputPersonalDocument
	return fmt.Sprintf("InputPersonalDocument%+v", Alias(*i))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InputPersonalDocument) TypeID() uint32 {
	return InputPersonalDocumentTypeID
}

// TypeName returns name of type in TL schema.
func (*InputPersonalDocument) TypeName() string {
	return "inputPersonalDocument"
}

// TypeInfo returns info about TL type.
func (i *InputPersonalDocument) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "inputPersonalDocument",
		ID:   InputPersonalDocumentTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Files",
			SchemaName: "files",
		},
		{
			Name:       "Translation",
			SchemaName: "translation",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (i *InputPersonalDocument) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputPersonalDocument#bb343fae as nil")
	}
	b.PutID(InputPersonalDocumentTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InputPersonalDocument) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode inputPersonalDocument#bb343fae as nil")
	}
	b.PutInt(len(i.Files))
	for idx, v := range i.Files {
		if v == nil {
			return fmt.Errorf("unable to encode inputPersonalDocument#bb343fae: field files element with index %d is nil", idx)
		}
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare inputPersonalDocument#bb343fae: field files element with index %d: %w", idx, err)
		}
	}
	b.PutInt(len(i.Translation))
	for idx, v := range i.Translation {
		if v == nil {
			return fmt.Errorf("unable to encode inputPersonalDocument#bb343fae: field translation element with index %d is nil", idx)
		}
		if err := v.EncodeBare(b); err != nil {
			return fmt.Errorf("unable to encode bare inputPersonalDocument#bb343fae: field translation element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (i *InputPersonalDocument) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputPersonalDocument#bb343fae to nil")
	}
	if err := b.ConsumeID(InputPersonalDocumentTypeID); err != nil {
		return fmt.Errorf("unable to decode inputPersonalDocument#bb343fae: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InputPersonalDocument) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode inputPersonalDocument#bb343fae to nil")
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode inputPersonalDocument#bb343fae: field files: %w", err)
		}

		if headerLen > 0 {
			i.Files = make([]InputFileClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputFile(b)
			if err != nil {
				return fmt.Errorf("unable to decode inputPersonalDocument#bb343fae: field files: %w", err)
			}
			i.Files = append(i.Files, value)
		}
	}
	{
		headerLen, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode inputPersonalDocument#bb343fae: field translation: %w", err)
		}

		if headerLen > 0 {
			i.Translation = make([]InputFileClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeInputFile(b)
			if err != nil {
				return fmt.Errorf("unable to decode inputPersonalDocument#bb343fae: field translation: %w", err)
			}
			i.Translation = append(i.Translation, value)
		}
	}
	return nil
}

// GetFiles returns value of Files field.
func (i *InputPersonalDocument) GetFiles() (value []InputFileClass) {
	return i.Files
}

// GetTranslation returns value of Translation field.
func (i *InputPersonalDocument) GetTranslation() (value []InputFileClass) {
	return i.Translation
}