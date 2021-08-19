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

// LangPackLanguage represents TL type `langPackLanguage#eeca5ce3`.
// Identifies a localization pack
//
// See https://core.telegram.org/constructor/langPackLanguage for reference.
type LangPackLanguage struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Whether the language pack is official
	Official bool
	// Is this a localization pack for an RTL language
	Rtl bool
	// Is this a beta localization pack?
	Beta bool
	// Language name
	Name string
	// Language name in the language itself
	NativeName string
	// Language code (pack identifier)
	LangCode string
	// Identifier of a base language pack; may be empty. If a string is missed in the
	// language pack, then it should be fetched from base language pack. Unsupported in
	// custom language packs
	//
	// Use SetBaseLangCode and GetBaseLangCode helpers.
	BaseLangCode string
	// A language code to be used to apply plural forms. See https://www.unicode
	// org/cldr/charts/latest/supplemental/language_plural_rules.html¹ for more info
	//
	// Links:
	//  1) https://www.unicode.org/cldr/charts/latest/supplemental/language_plural_rules.html
	PluralCode string
	// Total number of non-deleted strings from the language pack
	StringsCount int
	// Total number of translated strings from the language pack
	TranslatedCount int
	// Link to language translation interface; empty for custom local language packs
	TranslationsURL string
}

// LangPackLanguageTypeID is TL type id of LangPackLanguage.
const LangPackLanguageTypeID = 0xeeca5ce3

func (l *LangPackLanguage) Zero() bool {
	if l == nil {
		return true
	}
	if !(l.Flags.Zero()) {
		return false
	}
	if !(l.Official == false) {
		return false
	}
	if !(l.Rtl == false) {
		return false
	}
	if !(l.Beta == false) {
		return false
	}
	if !(l.Name == "") {
		return false
	}
	if !(l.NativeName == "") {
		return false
	}
	if !(l.LangCode == "") {
		return false
	}
	if !(l.BaseLangCode == "") {
		return false
	}
	if !(l.PluralCode == "") {
		return false
	}
	if !(l.StringsCount == 0) {
		return false
	}
	if !(l.TranslatedCount == 0) {
		return false
	}
	if !(l.TranslationsURL == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (l *LangPackLanguage) String() string {
	if l == nil {
		return "LangPackLanguage(nil)"
	}
	type Alias LangPackLanguage
	return fmt.Sprintf("LangPackLanguage%+v", Alias(*l))
}

// FillFrom fills LangPackLanguage from given interface.
func (l *LangPackLanguage) FillFrom(from interface {
	GetOfficial() (value bool)
	GetRtl() (value bool)
	GetBeta() (value bool)
	GetName() (value string)
	GetNativeName() (value string)
	GetLangCode() (value string)
	GetBaseLangCode() (value string, ok bool)
	GetPluralCode() (value string)
	GetStringsCount() (value int)
	GetTranslatedCount() (value int)
	GetTranslationsURL() (value string)
}) {
	l.Official = from.GetOfficial()
	l.Rtl = from.GetRtl()
	l.Beta = from.GetBeta()
	l.Name = from.GetName()
	l.NativeName = from.GetNativeName()
	l.LangCode = from.GetLangCode()
	if val, ok := from.GetBaseLangCode(); ok {
		l.BaseLangCode = val
	}

	l.PluralCode = from.GetPluralCode()
	l.StringsCount = from.GetStringsCount()
	l.TranslatedCount = from.GetTranslatedCount()
	l.TranslationsURL = from.GetTranslationsURL()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*LangPackLanguage) TypeID() uint32 {
	return LangPackLanguageTypeID
}

// TypeName returns name of type in TL schema.
func (*LangPackLanguage) TypeName() string {
	return "langPackLanguage"
}

// TypeInfo returns info about TL type.
func (l *LangPackLanguage) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "langPackLanguage",
		ID:   LangPackLanguageTypeID,
	}
	if l == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Official",
			SchemaName: "official",
			Null:       !l.Flags.Has(0),
		},
		{
			Name:       "Rtl",
			SchemaName: "rtl",
			Null:       !l.Flags.Has(2),
		},
		{
			Name:       "Beta",
			SchemaName: "beta",
			Null:       !l.Flags.Has(3),
		},
		{
			Name:       "Name",
			SchemaName: "name",
		},
		{
			Name:       "NativeName",
			SchemaName: "native_name",
		},
		{
			Name:       "LangCode",
			SchemaName: "lang_code",
		},
		{
			Name:       "BaseLangCode",
			SchemaName: "base_lang_code",
			Null:       !l.Flags.Has(1),
		},
		{
			Name:       "PluralCode",
			SchemaName: "plural_code",
		},
		{
			Name:       "StringsCount",
			SchemaName: "strings_count",
		},
		{
			Name:       "TranslatedCount",
			SchemaName: "translated_count",
		},
		{
			Name:       "TranslationsURL",
			SchemaName: "translations_url",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (l *LangPackLanguage) Encode(b *bin.Buffer) error {
	if l == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "langPackLanguage#eeca5ce3",
		}
	}
	b.PutID(LangPackLanguageTypeID)
	return l.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (l *LangPackLanguage) EncodeBare(b *bin.Buffer) error {
	if l == nil {
		return &bin.NilError{
			Action:   "encode",
			TypeName: "langPackLanguage#eeca5ce3",
		}
	}
	if !(l.Official == false) {
		l.Flags.Set(0)
	}
	if !(l.Rtl == false) {
		l.Flags.Set(2)
	}
	if !(l.Beta == false) {
		l.Flags.Set(3)
	}
	if !(l.BaseLangCode == "") {
		l.Flags.Set(1)
	}
	if err := l.Flags.Encode(b); err != nil {
		return &bin.FieldError{
			Action:     "encode",
			TypeName:   "langPackLanguage#eeca5ce3",
			FieldName:  "flags",
			Underlying: err,
		}
	}
	b.PutString(l.Name)
	b.PutString(l.NativeName)
	b.PutString(l.LangCode)
	if l.Flags.Has(1) {
		b.PutString(l.BaseLangCode)
	}
	b.PutString(l.PluralCode)
	b.PutInt(l.StringsCount)
	b.PutInt(l.TranslatedCount)
	b.PutString(l.TranslationsURL)
	return nil
}

// SetOfficial sets value of Official conditional field.
func (l *LangPackLanguage) SetOfficial(value bool) {
	if value {
		l.Flags.Set(0)
		l.Official = true
	} else {
		l.Flags.Unset(0)
		l.Official = false
	}
}

// GetOfficial returns value of Official conditional field.
func (l *LangPackLanguage) GetOfficial() (value bool) {
	return l.Flags.Has(0)
}

// SetRtl sets value of Rtl conditional field.
func (l *LangPackLanguage) SetRtl(value bool) {
	if value {
		l.Flags.Set(2)
		l.Rtl = true
	} else {
		l.Flags.Unset(2)
		l.Rtl = false
	}
}

// GetRtl returns value of Rtl conditional field.
func (l *LangPackLanguage) GetRtl() (value bool) {
	return l.Flags.Has(2)
}

// SetBeta sets value of Beta conditional field.
func (l *LangPackLanguage) SetBeta(value bool) {
	if value {
		l.Flags.Set(3)
		l.Beta = true
	} else {
		l.Flags.Unset(3)
		l.Beta = false
	}
}

// GetBeta returns value of Beta conditional field.
func (l *LangPackLanguage) GetBeta() (value bool) {
	return l.Flags.Has(3)
}

// GetName returns value of Name field.
func (l *LangPackLanguage) GetName() (value string) {
	return l.Name
}

// GetNativeName returns value of NativeName field.
func (l *LangPackLanguage) GetNativeName() (value string) {
	return l.NativeName
}

// GetLangCode returns value of LangCode field.
func (l *LangPackLanguage) GetLangCode() (value string) {
	return l.LangCode
}

// SetBaseLangCode sets value of BaseLangCode conditional field.
func (l *LangPackLanguage) SetBaseLangCode(value string) {
	l.Flags.Set(1)
	l.BaseLangCode = value
}

// GetBaseLangCode returns value of BaseLangCode conditional field and
// boolean which is true if field was set.
func (l *LangPackLanguage) GetBaseLangCode() (value string, ok bool) {
	if !l.Flags.Has(1) {
		return value, false
	}
	return l.BaseLangCode, true
}

// GetPluralCode returns value of PluralCode field.
func (l *LangPackLanguage) GetPluralCode() (value string) {
	return l.PluralCode
}

// GetStringsCount returns value of StringsCount field.
func (l *LangPackLanguage) GetStringsCount() (value int) {
	return l.StringsCount
}

// GetTranslatedCount returns value of TranslatedCount field.
func (l *LangPackLanguage) GetTranslatedCount() (value int) {
	return l.TranslatedCount
}

// GetTranslationsURL returns value of TranslationsURL field.
func (l *LangPackLanguage) GetTranslationsURL() (value string) {
	return l.TranslationsURL
}

// Decode implements bin.Decoder.
func (l *LangPackLanguage) Decode(b *bin.Buffer) error {
	if l == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "langPackLanguage#eeca5ce3",
		}
	}
	if err := b.ConsumeID(LangPackLanguageTypeID); err != nil {
		return &bin.DecodeError{
			TypeName:   "langPackLanguage#eeca5ce3",
			Underlying: err,
		}
	}
	return l.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (l *LangPackLanguage) DecodeBare(b *bin.Buffer) error {
	if l == nil {
		return &bin.NilError{
			Action:   "decode",
			TypeName: "langPackLanguage#eeca5ce3",
		}
	}
	{
		if err := l.Flags.Decode(b); err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "flags",
				Underlying: err,
			}
		}
	}
	l.Official = l.Flags.Has(0)
	l.Rtl = l.Flags.Has(2)
	l.Beta = l.Flags.Has(3)
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "name",
				Underlying: err,
			}
		}
		l.Name = value
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "native_name",
				Underlying: err,
			}
		}
		l.NativeName = value
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "lang_code",
				Underlying: err,
			}
		}
		l.LangCode = value
	}
	if l.Flags.Has(1) {
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "base_lang_code",
				Underlying: err,
			}
		}
		l.BaseLangCode = value
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "plural_code",
				Underlying: err,
			}
		}
		l.PluralCode = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "strings_count",
				Underlying: err,
			}
		}
		l.StringsCount = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "translated_count",
				Underlying: err,
			}
		}
		l.TranslatedCount = value
	}
	{
		value, err := b.String()
		if err != nil {
			return &bin.FieldError{
				Action:     "decode",
				TypeName:   "langPackLanguage#eeca5ce3",
				FieldName:  "translations_url",
				Underlying: err,
			}
		}
		l.TranslationsURL = value
	}
	return nil
}

// Ensuring interfaces in compile-time for LangPackLanguage.
var (
	_ bin.Encoder     = &LangPackLanguage{}
	_ bin.Decoder     = &LangPackLanguage{}
	_ bin.BareEncoder = &LangPackLanguage{}
	_ bin.BareDecoder = &LangPackLanguage{}
)
