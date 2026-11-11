package tljson

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"

	"github.com/go-faster/errors"
)

// Field is a key-value pair of an Object.
type Field struct {
	Key   string
	Value any
}

// Object is a decoded TL object with ordered fields. The "_" field holds the
// constructor or function name.
//
// Values are: int32 (TL int), uint32 (flags), Long (TL long), float64,
// string, Bytes, bool, []any (vectors) and *Object.
type Object struct {
	Fields []Field
}

// NewObject creates an object with the "_" field set to name.
func NewObject(name string) *Object {
	return &Object{Fields: []Field{{Key: "_", Value: name}}}
}

// Name returns the "_" field.
func (o *Object) Name() string {
	if o == nil {
		return ""
	}
	s, _ := o.Get("_").(string)
	return s
}

// Get returns the value of a field, or nil.
func (o *Object) Get(key string) any {
	if o == nil {
		return nil
	}
	for _, f := range o.Fields {
		if f.Key == key {
			return f.Value
		}
	}
	return nil
}

// Set sets a field, replacing the existing value.
func (o *Object) Set(key string, value any) {
	for i, f := range o.Fields {
		if f.Key == key {
			o.Fields[i].Value = value
			return
		}
	}
	o.Fields = append(o.Fields, Field{Key: key, Value: value})
}

// MarshalJSON implements json.Marshaler, keeping the field order.
func (o *Object) MarshalJSON() ([]byte, error) {
	if o == nil {
		return []byte("null"), nil
	}
	var b bytes.Buffer
	b.WriteByte('{')
	for i, f := range o.Fields {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := json.Marshal(f.Key)
		if err != nil {
			return nil, errors.Wrap(err, "key")
		}
		b.Write(k)
		b.WriteByte(':')
		v, err := marshalValue(f.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "field %q", f.Key)
		}
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func marshalValue(v any) ([]byte, error) {
	if f, ok := v.(float64); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
		return json.Marshal(strconv.FormatFloat(f, 'g', -1, 64))
	}
	return json.Marshal(v)
}

// Long is a TL long, encoded as a decimal JSON string to avoid float loss.
type Long int64

// MarshalJSON implements json.Marshaler.
func (l Long) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(l), 10) + `"`), nil
}

// Bytes is a TL bytes value, encoded as base64.
type Bytes []byte
