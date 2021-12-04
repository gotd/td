package tljson

import (
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/gotd/td/tg"
)

// Decode decodes JSON and converts it to tg.JSONValueClass.
func Decode(d *jx.Decoder) (tg.JSONValueClass, error) {
	switch tt := d.Next(); tt {
	case jx.String:
		s, err := d.Str()
		if err != nil {
			return nil, err
		}
		return &tg.JSONString{Value: s}, nil
	case jx.Number:
		f, err := d.Float64()
		if err != nil {
			return nil, err
		}
		return &tg.JSONNumber{Value: f}, nil
	case jx.Null:
		if err := d.Null(); err != nil {
			return nil, err
		}
		return &tg.JSONNull{}, nil
	case jx.Bool:
		b, err := d.Bool()
		if err != nil {
			return nil, err
		}
		return &tg.JSONBool{Value: b}, nil
	case jx.Array:
		var r []tg.JSONValueClass
		if err := d.Arr(func(d *jx.Decoder) error {
			obj, err := Decode(d)
			if err != nil {
				return errors.Wrapf(err, "decode %d element", len(r))
			}

			r = append(r, obj)
			return nil
		}); err != nil {
			return nil, err
		}
		return &tg.JSONArray{Value: r}, nil
	case jx.Object:
		var r []tg.JSONObjectValue
		if err := d.Obj(func(d *jx.Decoder, key string) error {
			obj, err := Decode(d)
			if err != nil {
				return errors.Wrapf(err, "decode %q", key)
			}

			r = append(r, tg.JSONObjectValue{
				Key:   key,
				Value: obj,
			})
			return nil
		}); err != nil {
			return nil, err
		}
		return &tg.JSONObject{Value: r}, nil
	default:
		return nil, errors.Errorf("unexpected type %v", tt)
	}
}

// Encode writes given value to Encoder.
func Encode(obj tg.JSONValueClass, e *jx.Encoder) {
	switch obj := obj.(type) {
	case *tg.JSONNull:
		e.Null()
	case *tg.JSONBool:
		e.Bool(obj.Value)
	case *tg.JSONNumber:
		if v := int64(obj.Value); float64(v) == obj.Value {
			e.Int64(v)
		} else {
			e.Float64(obj.Value)
		}
	case *tg.JSONString:
		e.Str(obj.Value)
	case *tg.JSONArray:
		e.ArrStart()
		for _, v := range obj.Value {
			Encode(v, e)
		}
		e.ArrEnd()
	case *tg.JSONObject:
		e.ObjStart()
		for _, pair := range obj.Value {
			e.FieldStart(pair.Key)
			Encode(pair.Value, e)
		}
		e.ObjEnd()
	}
}
