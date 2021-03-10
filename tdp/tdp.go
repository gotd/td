// Package tdp is td pretty-printing and formatting facilities for types from
// MTProto.
package tdp

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// options for formatting.
type options struct {
	writeTypeID bool
}

// Option of formatting.
type Option func(o *options)

// WithTypeID adds type id tp type name.
func WithTypeID(o *options) {
	o.writeTypeID = true
}

const (
	defaultIdent = "  "
	noIdent      = ""
)

func formatValue(b *strings.Builder, prefix, fieldName string, opt options, v reflect.Value) {
	switch v.Kind() {
	case reflect.Struct, reflect.Ptr, reflect.Interface:
		i, ok := v.Interface().(Object)
		if ok {
			format(b, prefix+defaultIdent, opt, i)
		} else if v.CanAddr() {
			formatValue(b, prefix, fieldName, opt, v.Addr())
		}
	case reflect.Slice:
		if buf, ok := v.Interface().([]byte); ok {
			b.WriteString(base64.RawURLEncoding.EncodeToString(buf))
			return
		}

		b.WriteRune('\n')
		for i := 0; i < v.Len(); i++ {
			vi := v.Index(i)
			b.WriteString(prefix)
			b.WriteString(defaultIdent)
			b.WriteString("- ")
			formatValue(b, prefix+defaultIdent, fieldName, opt, vi)
			b.WriteRune('\n')
		}
	case reflect.Int:
		// Special case for date.
		var (
			now = time.Now()
			max = now.AddDate(0, 0, 7).Unix()
			min = now.AddDate(-2, 0, 0).Unix()
		)
		i := v.Int()
		if i > min && i < max && strings.Contains(fieldName, "date") {
			b.WriteString(time.Unix(i, 0).UTC().Format(time.RFC3339))
		} else {
			b.WriteString(strconv.FormatInt(i, 10))
		}
	default:
		b.WriteString(fmt.Sprint(v.Interface()))
	}
}

func format(b *strings.Builder, prefix string, opt options, obj Object) {
	if obj == nil {
		// No type information is available. it is like Format(nil).
		b.WriteString("<nil>")
		return
	}

	info := obj.TypeInfo()
	b.WriteString(info.Name)
	if opt.writeTypeID {
		b.WriteRune('#')
		b.WriteString(strconv.FormatInt(int64(info.ID), 16))
	}
	if info.Null {
		b.WriteString("(nil)")
		return
	}

	v := reflect.ValueOf(obj).Elem()
	for i, f := range info.Fields {
		if i == 0 && f.SchemaName == "flags" {
			// Flag field, skipping.
			continue
		}
		if f.Null {
			// Optional field not set, skipping.
			continue
		}
		b.WriteRune('\n')
		b.WriteString(prefix)
		b.WriteString(defaultIdent)
		b.WriteString(f.SchemaName)
		b.WriteString(": ")

		formatValue(b, prefix, f.SchemaName, opt, v.FieldByName(f.Name))
	}
}

// Format pretty-prints v into string.
func Format(object Object, opts ...Option) string {
	var opt options
	for _, o := range opts {
		o(&opt)
	}

	var b strings.Builder
	format(&b, noIdent, opt, object)

	return b.String()
}
