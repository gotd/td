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

// WebAppOpenParameters represents TL type `webAppOpenParameters#51fa466f`.
type WebAppOpenParameters struct {
	// Preferred Web App theme; pass null to use the default theme
	Theme ThemeParameters
	// Short name of the current application; 0-64 English letters, digits, and underscores
	ApplicationName string
	// The mode in which the Web App is opened; pass null to open in webAppOpenModeFullSize
	Mode WebAppOpenModeClass
}

// WebAppOpenParametersTypeID is TL type id of WebAppOpenParameters.
const WebAppOpenParametersTypeID = 0x51fa466f

// Ensuring interfaces in compile-time for WebAppOpenParameters.
var (
	_ bin.Encoder     = &WebAppOpenParameters{}
	_ bin.Decoder     = &WebAppOpenParameters{}
	_ bin.BareEncoder = &WebAppOpenParameters{}
	_ bin.BareDecoder = &WebAppOpenParameters{}
)

func (w *WebAppOpenParameters) Zero() bool {
	if w == nil {
		return true
	}
	if !(w.Theme.Zero()) {
		return false
	}
	if !(w.ApplicationName == "") {
		return false
	}
	if !(w.Mode == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (w *WebAppOpenParameters) String() string {
	if w == nil {
		return "WebAppOpenParameters(nil)"
	}
	type Alias WebAppOpenParameters
	return fmt.Sprintf("WebAppOpenParameters%+v", Alias(*w))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*WebAppOpenParameters) TypeID() uint32 {
	return WebAppOpenParametersTypeID
}

// TypeName returns name of type in TL schema.
func (*WebAppOpenParameters) TypeName() string {
	return "webAppOpenParameters"
}

// TypeInfo returns info about TL type.
func (w *WebAppOpenParameters) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "webAppOpenParameters",
		ID:   WebAppOpenParametersTypeID,
	}
	if w == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Theme",
			SchemaName: "theme",
		},
		{
			Name:       "ApplicationName",
			SchemaName: "application_name",
		},
		{
			Name:       "Mode",
			SchemaName: "mode",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (w *WebAppOpenParameters) Encode(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't encode webAppOpenParameters#51fa466f as nil")
	}
	b.PutID(WebAppOpenParametersTypeID)
	return w.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (w *WebAppOpenParameters) EncodeBare(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't encode webAppOpenParameters#51fa466f as nil")
	}
	if err := w.Theme.Encode(b); err != nil {
		return fmt.Errorf("unable to encode webAppOpenParameters#51fa466f: field theme: %w", err)
	}
	b.PutString(w.ApplicationName)
	if w.Mode == nil {
		return fmt.Errorf("unable to encode webAppOpenParameters#51fa466f: field mode is nil")
	}
	if err := w.Mode.Encode(b); err != nil {
		return fmt.Errorf("unable to encode webAppOpenParameters#51fa466f: field mode: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (w *WebAppOpenParameters) Decode(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't decode webAppOpenParameters#51fa466f to nil")
	}
	if err := b.ConsumeID(WebAppOpenParametersTypeID); err != nil {
		return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: %w", err)
	}
	return w.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (w *WebAppOpenParameters) DecodeBare(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't decode webAppOpenParameters#51fa466f to nil")
	}
	{
		if err := w.Theme.Decode(b); err != nil {
			return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: field theme: %w", err)
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: field application_name: %w", err)
		}
		w.ApplicationName = value
	}
	{
		value, err := DecodeWebAppOpenMode(b)
		if err != nil {
			return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: field mode: %w", err)
		}
		w.Mode = value
	}
	return nil
}

// EncodeTDLibJSON implements tdjson.TDLibEncoder.
func (w *WebAppOpenParameters) EncodeTDLibJSON(b tdjson.Encoder) error {
	if w == nil {
		return fmt.Errorf("can't encode webAppOpenParameters#51fa466f as nil")
	}
	b.ObjStart()
	b.PutID("webAppOpenParameters")
	b.Comma()
	b.FieldStart("theme")
	if err := w.Theme.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode webAppOpenParameters#51fa466f: field theme: %w", err)
	}
	b.Comma()
	b.FieldStart("application_name")
	b.PutString(w.ApplicationName)
	b.Comma()
	b.FieldStart("mode")
	if w.Mode == nil {
		return fmt.Errorf("unable to encode webAppOpenParameters#51fa466f: field mode is nil")
	}
	if err := w.Mode.EncodeTDLibJSON(b); err != nil {
		return fmt.Errorf("unable to encode webAppOpenParameters#51fa466f: field mode: %w", err)
	}
	b.Comma()
	b.StripComma()
	b.ObjEnd()
	return nil
}

// DecodeTDLibJSON implements tdjson.TDLibDecoder.
func (w *WebAppOpenParameters) DecodeTDLibJSON(b tdjson.Decoder) error {
	if w == nil {
		return fmt.Errorf("can't decode webAppOpenParameters#51fa466f to nil")
	}

	return b.Obj(func(b tdjson.Decoder, key []byte) error {
		switch string(key) {
		case tdjson.TypeField:
			if err := b.ConsumeID("webAppOpenParameters"); err != nil {
				return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: %w", err)
			}
		case "theme":
			if err := w.Theme.DecodeTDLibJSON(b); err != nil {
				return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: field theme: %w", err)
			}
		case "application_name":
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: field application_name: %w", err)
			}
			w.ApplicationName = value
		case "mode":
			value, err := DecodeTDLibJSONWebAppOpenMode(b)
			if err != nil {
				return fmt.Errorf("unable to decode webAppOpenParameters#51fa466f: field mode: %w", err)
			}
			w.Mode = value
		default:
			return b.Skip()
		}
		return nil
	})
}

// GetTheme returns value of Theme field.
func (w *WebAppOpenParameters) GetTheme() (value ThemeParameters) {
	if w == nil {
		return
	}
	return w.Theme
}

// GetApplicationName returns value of ApplicationName field.
func (w *WebAppOpenParameters) GetApplicationName() (value string) {
	if w == nil {
		return
	}
	return w.ApplicationName
}

// GetMode returns value of Mode field.
func (w *WebAppOpenParameters) GetMode() (value WebAppOpenModeClass) {
	if w == nil {
		return
	}
	return w.Mode
}