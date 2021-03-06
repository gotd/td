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

// JSONNull represents TL type `jsonNull#3f6d7b68`.
// null JSON value
//
// See https://core.telegram.org/constructor/jsonNull for reference.
type JSONNull struct {
}

// JSONNullTypeID is TL type id of JSONNull.
const JSONNullTypeID = 0x3f6d7b68

func (j *JSONNull) Zero() bool {
	if j == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (j *JSONNull) String() string {
	if j == nil {
		return "JSONNull(nil)"
	}
	type Alias JSONNull
	return fmt.Sprintf("JSONNull%+v", Alias(*j))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*JSONNull) TypeID() uint32 {
	return JSONNullTypeID
}

// TypeName returns name of type in TL schema.
func (*JSONNull) TypeName() string {
	return "jsonNull"
}

// TypeInfo returns info about TL type.
func (j *JSONNull) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "jsonNull",
		ID:   JSONNullTypeID,
	}
	if j == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (j *JSONNull) Encode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonNull#3f6d7b68 as nil")
	}
	b.PutID(JSONNullTypeID)
	return j.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (j *JSONNull) EncodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonNull#3f6d7b68 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (j *JSONNull) Decode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonNull#3f6d7b68 to nil")
	}
	if err := b.ConsumeID(JSONNullTypeID); err != nil {
		return fmt.Errorf("unable to decode jsonNull#3f6d7b68: %w", err)
	}
	return j.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (j *JSONNull) DecodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonNull#3f6d7b68 to nil")
	}
	return nil
}

// construct implements constructor of JSONValueClass.
func (j JSONNull) construct() JSONValueClass { return &j }

// Ensuring interfaces in compile-time for JSONNull.
var (
	_ bin.Encoder     = &JSONNull{}
	_ bin.Decoder     = &JSONNull{}
	_ bin.BareEncoder = &JSONNull{}
	_ bin.BareDecoder = &JSONNull{}

	_ JSONValueClass = &JSONNull{}
)

// JSONBool represents TL type `jsonBool#c7345e6a`.
// JSON boolean value
//
// See https://core.telegram.org/constructor/jsonBool for reference.
type JSONBool struct {
	// Value
	Value bool
}

// JSONBoolTypeID is TL type id of JSONBool.
const JSONBoolTypeID = 0xc7345e6a

func (j *JSONBool) Zero() bool {
	if j == nil {
		return true
	}
	if !(j.Value == false) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (j *JSONBool) String() string {
	if j == nil {
		return "JSONBool(nil)"
	}
	type Alias JSONBool
	return fmt.Sprintf("JSONBool%+v", Alias(*j))
}

// FillFrom fills JSONBool from given interface.
func (j *JSONBool) FillFrom(from interface {
	GetValue() (value bool)
}) {
	j.Value = from.GetValue()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*JSONBool) TypeID() uint32 {
	return JSONBoolTypeID
}

// TypeName returns name of type in TL schema.
func (*JSONBool) TypeName() string {
	return "jsonBool"
}

// TypeInfo returns info about TL type.
func (j *JSONBool) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "jsonBool",
		ID:   JSONBoolTypeID,
	}
	if j == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Value",
			SchemaName: "value",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (j *JSONBool) Encode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonBool#c7345e6a as nil")
	}
	b.PutID(JSONBoolTypeID)
	return j.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (j *JSONBool) EncodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonBool#c7345e6a as nil")
	}
	b.PutBool(j.Value)
	return nil
}

// GetValue returns value of Value field.
func (j *JSONBool) GetValue() (value bool) {
	return j.Value
}

// Decode implements bin.Decoder.
func (j *JSONBool) Decode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonBool#c7345e6a to nil")
	}
	if err := b.ConsumeID(JSONBoolTypeID); err != nil {
		return fmt.Errorf("unable to decode jsonBool#c7345e6a: %w", err)
	}
	return j.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (j *JSONBool) DecodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonBool#c7345e6a to nil")
	}
	{
		value, err := b.Bool()
		if err != nil {
			return fmt.Errorf("unable to decode jsonBool#c7345e6a: field value: %w", err)
		}
		j.Value = value
	}
	return nil
}

// construct implements constructor of JSONValueClass.
func (j JSONBool) construct() JSONValueClass { return &j }

// Ensuring interfaces in compile-time for JSONBool.
var (
	_ bin.Encoder     = &JSONBool{}
	_ bin.Decoder     = &JSONBool{}
	_ bin.BareEncoder = &JSONBool{}
	_ bin.BareDecoder = &JSONBool{}

	_ JSONValueClass = &JSONBool{}
)

// JSONNumber represents TL type `jsonNumber#2be0dfa4`.
// JSON numeric value
//
// See https://core.telegram.org/constructor/jsonNumber for reference.
type JSONNumber struct {
	// Value
	Value float64
}

// JSONNumberTypeID is TL type id of JSONNumber.
const JSONNumberTypeID = 0x2be0dfa4

func (j *JSONNumber) Zero() bool {
	if j == nil {
		return true
	}
	if !(j.Value == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (j *JSONNumber) String() string {
	if j == nil {
		return "JSONNumber(nil)"
	}
	type Alias JSONNumber
	return fmt.Sprintf("JSONNumber%+v", Alias(*j))
}

// FillFrom fills JSONNumber from given interface.
func (j *JSONNumber) FillFrom(from interface {
	GetValue() (value float64)
}) {
	j.Value = from.GetValue()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*JSONNumber) TypeID() uint32 {
	return JSONNumberTypeID
}

// TypeName returns name of type in TL schema.
func (*JSONNumber) TypeName() string {
	return "jsonNumber"
}

// TypeInfo returns info about TL type.
func (j *JSONNumber) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "jsonNumber",
		ID:   JSONNumberTypeID,
	}
	if j == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Value",
			SchemaName: "value",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (j *JSONNumber) Encode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonNumber#2be0dfa4 as nil")
	}
	b.PutID(JSONNumberTypeID)
	return j.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (j *JSONNumber) EncodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonNumber#2be0dfa4 as nil")
	}
	b.PutDouble(j.Value)
	return nil
}

// GetValue returns value of Value field.
func (j *JSONNumber) GetValue() (value float64) {
	return j.Value
}

// Decode implements bin.Decoder.
func (j *JSONNumber) Decode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonNumber#2be0dfa4 to nil")
	}
	if err := b.ConsumeID(JSONNumberTypeID); err != nil {
		return fmt.Errorf("unable to decode jsonNumber#2be0dfa4: %w", err)
	}
	return j.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (j *JSONNumber) DecodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonNumber#2be0dfa4 to nil")
	}
	{
		value, err := b.Double()
		if err != nil {
			return fmt.Errorf("unable to decode jsonNumber#2be0dfa4: field value: %w", err)
		}
		j.Value = value
	}
	return nil
}

// construct implements constructor of JSONValueClass.
func (j JSONNumber) construct() JSONValueClass { return &j }

// Ensuring interfaces in compile-time for JSONNumber.
var (
	_ bin.Encoder     = &JSONNumber{}
	_ bin.Decoder     = &JSONNumber{}
	_ bin.BareEncoder = &JSONNumber{}
	_ bin.BareDecoder = &JSONNumber{}

	_ JSONValueClass = &JSONNumber{}
)

// JSONString represents TL type `jsonString#b71e767a`.
// JSON string
//
// See https://core.telegram.org/constructor/jsonString for reference.
type JSONString struct {
	// Value
	Value string
}

// JSONStringTypeID is TL type id of JSONString.
const JSONStringTypeID = 0xb71e767a

func (j *JSONString) Zero() bool {
	if j == nil {
		return true
	}
	if !(j.Value == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (j *JSONString) String() string {
	if j == nil {
		return "JSONString(nil)"
	}
	type Alias JSONString
	return fmt.Sprintf("JSONString%+v", Alias(*j))
}

// FillFrom fills JSONString from given interface.
func (j *JSONString) FillFrom(from interface {
	GetValue() (value string)
}) {
	j.Value = from.GetValue()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*JSONString) TypeID() uint32 {
	return JSONStringTypeID
}

// TypeName returns name of type in TL schema.
func (*JSONString) TypeName() string {
	return "jsonString"
}

// TypeInfo returns info about TL type.
func (j *JSONString) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "jsonString",
		ID:   JSONStringTypeID,
	}
	if j == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Value",
			SchemaName: "value",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (j *JSONString) Encode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonString#b71e767a as nil")
	}
	b.PutID(JSONStringTypeID)
	return j.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (j *JSONString) EncodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonString#b71e767a as nil")
	}
	b.PutString(j.Value)
	return nil
}

// GetValue returns value of Value field.
func (j *JSONString) GetValue() (value string) {
	return j.Value
}

// Decode implements bin.Decoder.
func (j *JSONString) Decode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonString#b71e767a to nil")
	}
	if err := b.ConsumeID(JSONStringTypeID); err != nil {
		return fmt.Errorf("unable to decode jsonString#b71e767a: %w", err)
	}
	return j.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (j *JSONString) DecodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonString#b71e767a to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode jsonString#b71e767a: field value: %w", err)
		}
		j.Value = value
	}
	return nil
}

// construct implements constructor of JSONValueClass.
func (j JSONString) construct() JSONValueClass { return &j }

// Ensuring interfaces in compile-time for JSONString.
var (
	_ bin.Encoder     = &JSONString{}
	_ bin.Decoder     = &JSONString{}
	_ bin.BareEncoder = &JSONString{}
	_ bin.BareDecoder = &JSONString{}

	_ JSONValueClass = &JSONString{}
)

// JSONArray represents TL type `jsonArray#f7444763`.
// JSON array
//
// See https://core.telegram.org/constructor/jsonArray for reference.
type JSONArray struct {
	// JSON values
	Value []JSONValueClass
}

// JSONArrayTypeID is TL type id of JSONArray.
const JSONArrayTypeID = 0xf7444763

func (j *JSONArray) Zero() bool {
	if j == nil {
		return true
	}
	if !(j.Value == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (j *JSONArray) String() string {
	if j == nil {
		return "JSONArray(nil)"
	}
	type Alias JSONArray
	return fmt.Sprintf("JSONArray%+v", Alias(*j))
}

// FillFrom fills JSONArray from given interface.
func (j *JSONArray) FillFrom(from interface {
	GetValue() (value []JSONValueClass)
}) {
	j.Value = from.GetValue()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*JSONArray) TypeID() uint32 {
	return JSONArrayTypeID
}

// TypeName returns name of type in TL schema.
func (*JSONArray) TypeName() string {
	return "jsonArray"
}

// TypeInfo returns info about TL type.
func (j *JSONArray) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "jsonArray",
		ID:   JSONArrayTypeID,
	}
	if j == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Value",
			SchemaName: "value",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (j *JSONArray) Encode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonArray#f7444763 as nil")
	}
	b.PutID(JSONArrayTypeID)
	return j.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (j *JSONArray) EncodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonArray#f7444763 as nil")
	}
	b.PutVectorHeader(len(j.Value))
	for idx, v := range j.Value {
		if v == nil {
			return fmt.Errorf("unable to encode jsonArray#f7444763: field value element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode jsonArray#f7444763: field value element with index %d: %w", idx, err)
		}
	}
	return nil
}

// GetValue returns value of Value field.
func (j *JSONArray) GetValue() (value []JSONValueClass) {
	return j.Value
}

// MapValue returns field Value wrapped in JSONValueClassArray helper.
func (j *JSONArray) MapValue() (value JSONValueClassArray) {
	return JSONValueClassArray(j.Value)
}

// Decode implements bin.Decoder.
func (j *JSONArray) Decode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonArray#f7444763 to nil")
	}
	if err := b.ConsumeID(JSONArrayTypeID); err != nil {
		return fmt.Errorf("unable to decode jsonArray#f7444763: %w", err)
	}
	return j.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (j *JSONArray) DecodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonArray#f7444763 to nil")
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode jsonArray#f7444763: field value: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeJSONValue(b)
			if err != nil {
				return fmt.Errorf("unable to decode jsonArray#f7444763: field value: %w", err)
			}
			j.Value = append(j.Value, value)
		}
	}
	return nil
}

// construct implements constructor of JSONValueClass.
func (j JSONArray) construct() JSONValueClass { return &j }

// Ensuring interfaces in compile-time for JSONArray.
var (
	_ bin.Encoder     = &JSONArray{}
	_ bin.Decoder     = &JSONArray{}
	_ bin.BareEncoder = &JSONArray{}
	_ bin.BareDecoder = &JSONArray{}

	_ JSONValueClass = &JSONArray{}
)

// JSONObject represents TL type `jsonObject#99c1d49d`.
// JSON object value
//
// See https://core.telegram.org/constructor/jsonObject for reference.
type JSONObject struct {
	// Values
	Value []JSONObjectValue
}

// JSONObjectTypeID is TL type id of JSONObject.
const JSONObjectTypeID = 0x99c1d49d

func (j *JSONObject) Zero() bool {
	if j == nil {
		return true
	}
	if !(j.Value == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (j *JSONObject) String() string {
	if j == nil {
		return "JSONObject(nil)"
	}
	type Alias JSONObject
	return fmt.Sprintf("JSONObject%+v", Alias(*j))
}

// FillFrom fills JSONObject from given interface.
func (j *JSONObject) FillFrom(from interface {
	GetValue() (value []JSONObjectValue)
}) {
	j.Value = from.GetValue()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*JSONObject) TypeID() uint32 {
	return JSONObjectTypeID
}

// TypeName returns name of type in TL schema.
func (*JSONObject) TypeName() string {
	return "jsonObject"
}

// TypeInfo returns info about TL type.
func (j *JSONObject) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "jsonObject",
		ID:   JSONObjectTypeID,
	}
	if j == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Value",
			SchemaName: "value",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (j *JSONObject) Encode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonObject#99c1d49d as nil")
	}
	b.PutID(JSONObjectTypeID)
	return j.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (j *JSONObject) EncodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't encode jsonObject#99c1d49d as nil")
	}
	b.PutVectorHeader(len(j.Value))
	for idx, v := range j.Value {
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode jsonObject#99c1d49d: field value element with index %d: %w", idx, err)
		}
	}
	return nil
}

// GetValue returns value of Value field.
func (j *JSONObject) GetValue() (value []JSONObjectValue) {
	return j.Value
}

// Decode implements bin.Decoder.
func (j *JSONObject) Decode(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonObject#99c1d49d to nil")
	}
	if err := b.ConsumeID(JSONObjectTypeID); err != nil {
		return fmt.Errorf("unable to decode jsonObject#99c1d49d: %w", err)
	}
	return j.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (j *JSONObject) DecodeBare(b *bin.Buffer) error {
	if j == nil {
		return fmt.Errorf("can't decode jsonObject#99c1d49d to nil")
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode jsonObject#99c1d49d: field value: %w", err)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value JSONObjectValue
			if err := value.Decode(b); err != nil {
				return fmt.Errorf("unable to decode jsonObject#99c1d49d: field value: %w", err)
			}
			j.Value = append(j.Value, value)
		}
	}
	return nil
}

// construct implements constructor of JSONValueClass.
func (j JSONObject) construct() JSONValueClass { return &j }

// Ensuring interfaces in compile-time for JSONObject.
var (
	_ bin.Encoder     = &JSONObject{}
	_ bin.Decoder     = &JSONObject{}
	_ bin.BareEncoder = &JSONObject{}
	_ bin.BareDecoder = &JSONObject{}

	_ JSONValueClass = &JSONObject{}
)

// JSONValueClass represents JSONValue generic type.
//
// See https://core.telegram.org/type/JSONValue for reference.
//
// Example:
//  g, err := tg.DecodeJSONValue(buf)
//  if err != nil {
//      panic(err)
//  }
//  switch v := g.(type) {
//  case *tg.JSONNull: // jsonNull#3f6d7b68
//  case *tg.JSONBool: // jsonBool#c7345e6a
//  case *tg.JSONNumber: // jsonNumber#2be0dfa4
//  case *tg.JSONString: // jsonString#b71e767a
//  case *tg.JSONArray: // jsonArray#f7444763
//  case *tg.JSONObject: // jsonObject#99c1d49d
//  default: panic(v)
//  }
type JSONValueClass interface {
	bin.Encoder
	bin.Decoder
	bin.BareEncoder
	bin.BareDecoder
	construct() JSONValueClass

	// TypeID returns type id in TL schema.
	//
	// See https://core.telegram.org/mtproto/TL-tl#remarks.
	TypeID() uint32
	// TypeName returns name of type in TL schema.
	TypeName() string
	// String implements fmt.Stringer.
	String() string
	// Zero returns true if current object has a zero value.
	Zero() bool
}

// DecodeJSONValue implements binary de-serialization for JSONValueClass.
func DecodeJSONValue(buf *bin.Buffer) (JSONValueClass, error) {
	id, err := buf.PeekID()
	if err != nil {
		return nil, err
	}
	switch id {
	case JSONNullTypeID:
		// Decoding jsonNull#3f6d7b68.
		v := JSONNull{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode JSONValueClass: %w", err)
		}
		return &v, nil
	case JSONBoolTypeID:
		// Decoding jsonBool#c7345e6a.
		v := JSONBool{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode JSONValueClass: %w", err)
		}
		return &v, nil
	case JSONNumberTypeID:
		// Decoding jsonNumber#2be0dfa4.
		v := JSONNumber{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode JSONValueClass: %w", err)
		}
		return &v, nil
	case JSONStringTypeID:
		// Decoding jsonString#b71e767a.
		v := JSONString{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode JSONValueClass: %w", err)
		}
		return &v, nil
	case JSONArrayTypeID:
		// Decoding jsonArray#f7444763.
		v := JSONArray{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode JSONValueClass: %w", err)
		}
		return &v, nil
	case JSONObjectTypeID:
		// Decoding jsonObject#99c1d49d.
		v := JSONObject{}
		if err := v.Decode(buf); err != nil {
			return nil, fmt.Errorf("unable to decode JSONValueClass: %w", err)
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("unable to decode JSONValueClass: %w", bin.NewUnexpectedID(id))
	}
}

// JSONValue boxes the JSONValueClass providing a helper.
type JSONValueBox struct {
	JSONValue JSONValueClass
}

// Decode implements bin.Decoder for JSONValueBox.
func (b *JSONValueBox) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("unable to decode JSONValueBox to nil")
	}
	v, err := DecodeJSONValue(buf)
	if err != nil {
		return fmt.Errorf("unable to decode boxed value: %w", err)
	}
	b.JSONValue = v
	return nil
}

// Encode implements bin.Encode for JSONValueBox.
func (b *JSONValueBox) Encode(buf *bin.Buffer) error {
	if b == nil || b.JSONValue == nil {
		return fmt.Errorf("unable to encode JSONValueClass as nil")
	}
	return b.JSONValue.Encode(buf)
}

// JSONValueClassArray is adapter for slice of JSONValueClass.
type JSONValueClassArray []JSONValueClass

// Sort sorts slice of JSONValueClass.
func (s JSONValueClassArray) Sort(less func(a, b JSONValueClass) bool) JSONValueClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of JSONValueClass.
func (s JSONValueClassArray) SortStable(less func(a, b JSONValueClass) bool) JSONValueClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of JSONValueClass.
func (s JSONValueClassArray) Retain(keep func(x JSONValueClass) bool) JSONValueClassArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s JSONValueClassArray) First() (v JSONValueClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s JSONValueClassArray) Last() (v JSONValueClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *JSONValueClassArray) PopFirst() (v JSONValueClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero JSONValueClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *JSONValueClassArray) Pop() (v JSONValueClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsJSONBool returns copy with only JSONBool constructors.
func (s JSONValueClassArray) AsJSONBool() (to JSONBoolArray) {
	for _, elem := range s {
		value, ok := elem.(*JSONBool)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsJSONNumber returns copy with only JSONNumber constructors.
func (s JSONValueClassArray) AsJSONNumber() (to JSONNumberArray) {
	for _, elem := range s {
		value, ok := elem.(*JSONNumber)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsJSONString returns copy with only JSONString constructors.
func (s JSONValueClassArray) AsJSONString() (to JSONStringArray) {
	for _, elem := range s {
		value, ok := elem.(*JSONString)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsJSONArray returns copy with only JSONArray constructors.
func (s JSONValueClassArray) AsJSONArray() (to JSONArrayArray) {
	for _, elem := range s {
		value, ok := elem.(*JSONArray)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsJSONObject returns copy with only JSONObject constructors.
func (s JSONValueClassArray) AsJSONObject() (to JSONObjectArray) {
	for _, elem := range s {
		value, ok := elem.(*JSONObject)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// JSONBoolArray is adapter for slice of JSONBool.
type JSONBoolArray []JSONBool

// Sort sorts slice of JSONBool.
func (s JSONBoolArray) Sort(less func(a, b JSONBool) bool) JSONBoolArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of JSONBool.
func (s JSONBoolArray) SortStable(less func(a, b JSONBool) bool) JSONBoolArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of JSONBool.
func (s JSONBoolArray) Retain(keep func(x JSONBool) bool) JSONBoolArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s JSONBoolArray) First() (v JSONBool, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s JSONBoolArray) Last() (v JSONBool, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *JSONBoolArray) PopFirst() (v JSONBool, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero JSONBool
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *JSONBoolArray) Pop() (v JSONBool, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// JSONNumberArray is adapter for slice of JSONNumber.
type JSONNumberArray []JSONNumber

// Sort sorts slice of JSONNumber.
func (s JSONNumberArray) Sort(less func(a, b JSONNumber) bool) JSONNumberArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of JSONNumber.
func (s JSONNumberArray) SortStable(less func(a, b JSONNumber) bool) JSONNumberArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of JSONNumber.
func (s JSONNumberArray) Retain(keep func(x JSONNumber) bool) JSONNumberArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s JSONNumberArray) First() (v JSONNumber, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s JSONNumberArray) Last() (v JSONNumber, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *JSONNumberArray) PopFirst() (v JSONNumber, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero JSONNumber
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *JSONNumberArray) Pop() (v JSONNumber, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// JSONStringArray is adapter for slice of JSONString.
type JSONStringArray []JSONString

// Sort sorts slice of JSONString.
func (s JSONStringArray) Sort(less func(a, b JSONString) bool) JSONStringArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of JSONString.
func (s JSONStringArray) SortStable(less func(a, b JSONString) bool) JSONStringArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of JSONString.
func (s JSONStringArray) Retain(keep func(x JSONString) bool) JSONStringArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s JSONStringArray) First() (v JSONString, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s JSONStringArray) Last() (v JSONString, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *JSONStringArray) PopFirst() (v JSONString, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero JSONString
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *JSONStringArray) Pop() (v JSONString, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// JSONArrayArray is adapter for slice of JSONArray.
type JSONArrayArray []JSONArray

// Sort sorts slice of JSONArray.
func (s JSONArrayArray) Sort(less func(a, b JSONArray) bool) JSONArrayArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of JSONArray.
func (s JSONArrayArray) SortStable(less func(a, b JSONArray) bool) JSONArrayArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of JSONArray.
func (s JSONArrayArray) Retain(keep func(x JSONArray) bool) JSONArrayArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s JSONArrayArray) First() (v JSONArray, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s JSONArrayArray) Last() (v JSONArray, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *JSONArrayArray) PopFirst() (v JSONArray, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero JSONArray
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *JSONArrayArray) Pop() (v JSONArray, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// JSONObjectArray is adapter for slice of JSONObject.
type JSONObjectArray []JSONObject

// Sort sorts slice of JSONObject.
func (s JSONObjectArray) Sort(less func(a, b JSONObject) bool) JSONObjectArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of JSONObject.
func (s JSONObjectArray) SortStable(less func(a, b JSONObject) bool) JSONObjectArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of JSONObject.
func (s JSONObjectArray) Retain(keep func(x JSONObject) bool) JSONObjectArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s JSONObjectArray) First() (v JSONObject, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s JSONObjectArray) Last() (v JSONObject, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *JSONObjectArray) PopFirst() (v JSONObject, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero JSONObject
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *JSONObjectArray) Pop() (v JSONObject, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
