package gen

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/tl"
)

const flagsType = "bin.Fields"

// fieldDef represents structDef field.
type fieldDef struct {
	// Name of field. Should be in camel case.
	Name string
	// Comment for field. Currently only one-line.
	Comment []string
	// Type is go type for field.
	Type string
	// Func is name for bin.* functions, e.g. String will render
	// to bin.Buffer.String and bin.Buffer.PutString.
	Func string
	// Encoder denotes whether fieldDef implements bin.Encoder and bin.Decoder.
	Encoder bool
	// RawName is raw name from TL Schema.
	RawName string
	// RawType is type from TL Schema.
	RawType string
	// BareVector denotes whether fieldDef TL type is a bare vector.
	BareVector bool
	// Vector denotes whether fieldDef TL type is vector.
	Vector bool
	// DoubleVector denotes whether fieldDef TL type is vector of vectors.
	DoubleVector bool
	// Slice denotes whether slice should be used for this field.
	Slice bool
	// DoubleSlice denotes whether double slicing should be used, e.g. [][]bytes.
	DoubleSlice bool
	// BareEncoder denotes whether field type should use bare encoder.
	BareEncoder bool
	// Interface is name of interface type if field type is constructor.
	Interface string
	// InterfaceFunc is encoding func postfix if Interface is set.
	InterfaceFunc string
	// Conditional denotes whether fieldDef is conditional.
	Conditional bool
	// ConditionalField if name of bitset param.
	ConditionalField string
	// ConditionalIndex is fieldDef bit in ConditionalField.
	ConditionalIndex int
	// ConditionalBool denotes whether value is fully encoded in ConditionalField as bit.
	ConditionalBool bool
	// Links from documentation
	Links []string
}

type fieldPair struct {
	L, R fieldDef
}

func (f fieldDef) String() string {
	b := strings.Builder{}
	b.Grow(len(f.Name) + len(f.Type) + 16)
	b.WriteString(f.Name)
	b.WriteByte(' ')

	switch {
	case f.Slice || f.Vector:
		b.WriteString("[]")
	case f.DoubleSlice || f.DoubleVector:
		b.WriteString("[][]")
	}
	b.WriteString(f.Type)
	switch {
	case f.Conditional:
		b.WriteString("?")
	case f.ConditionalBool:
		b.WriteString("?true")
	}

	return b.String()
}

func (f fieldDef) SameType(b fieldDef) bool {
	return f.Type == b.Type &&
		f.Func == b.Func &&
		f.Vector == b.Vector &&
		f.DoubleVector == b.DoubleVector &&
		f.Slice == b.Slice &&
		f.DoubleSlice == b.DoubleSlice
}

func (f fieldDef) EqualAsField(b fieldDef) bool {
	return f.Name == b.Name &&
		f.SameType(b) &&
		f.Interface == b.Interface &&
		f.InterfaceFunc == b.InterfaceFunc &&
		f.Conditional == b.Conditional &&
		f.ConditionalBool == b.ConditionalBool
}

// nolint:gocognit,gocyclo
//
// TODO(ernado) Split into multiple sections: base type, encoder and conditional.
func (g *Generator) makeField(param tl.Parameter, annotations []tl.Annotation) (fieldDef, error) {
	const bareVectorName = "vector"

	f := fieldDef{
		Name:    pascal(param.Name),
		RawName: param.Name,
		RawType: param.Type.String(),
	}
	// Unwrapping up to 2 levels of vectors.
	baseType := param.Type
	for _, a := range annotations {
		if a.Name == param.Name {
			if a.Value != "" {
				f.Comment = []string{a.Value}
			}
		}
	}
	if baseType.Name == bareVectorName || baseType.Name == "Vector" {
		f.BareVector = baseType.Name == bareVectorName
		baseType = *baseType.GenericArg
		f.Vector = true
		f.Slice = true
		f.BareVector = f.BareVector || baseType.Percent
	}
	if baseType.Name == bareVectorName || baseType.Name == "Vector" {
		baseType = *baseType.GenericArg
		f.DoubleSlice = true
		f.DoubleVector = true
	}
	if param.Flags {
		f.Type = flagsType
	}
	f.Type = baseType.Name
	switch baseType.Name {
	case "int":
		f.Func = "Int"
	case "int32":
		f.Func = "Int32"
	case "int128":
		f.Func = "Int128"
		f.Type = "bin.Int128"
	case "int256":
		f.Func = "Int256"
		f.Type = "bin.Int256"
	case "double":
		f.Func = "Double"
		f.Type = "float64"
	case "int53":
		f.Func = "Int53"
		f.Type = "int64"
	case "long", "int64":
		f.Func = "Long"
		f.Type = "int64"
	case "string":
		f.Func = "String"
	case "Bool", "bool", "true", "false":
		f.Func = "Bool"
		f.Type = "bool"
	case "bytes":
		f.Func = "Bytes"
		f.Type = "byte"
		if f.Slice {
			f.DoubleSlice = true
		}
		f.Slice = true
	default:
		f.Encoder = true
		f.BareEncoder = f.BareVector

		if param.Flags {
			f.Type = flagsType
			break
		}

		if baseType.Bare {
			// Using exact go type for bare types.
			tn := strings.TrimPrefix(baseType.String(), "%")
			t, ok := g.types[tn]
			if !ok {
				return fieldDef{}, errors.Errorf("types[%s] not found", baseType)
			}
			f.Type = t.Name
		} else {
			// Type is generic.
			t, ok := g.classes[baseType.String()]
			if param.Type.GenericRef {
				ok = true
				t.RawType = param.Type.Name
				t.Name = "bin.Object"
			}
			if !ok {
				return fieldDef{}, errors.Errorf("classes[%s] not found", baseType)
			}
			f.Type = t.Name
			if !baseType.Percent && t.Singular && !param.Type.GenericRef {
				f.BareEncoder = false
			}
			if !t.Singular && !param.Type.GenericRef {
				f.Interface = t.Name
				f.InterfaceFunc = t.Func
			}
		}
	}
	if flag := param.Flag; flag != nil {
		f.Conditional = true
		f.ConditionalIndex = flag.Index
		f.ConditionalField = pascal(flag.Name)
		if f.Type == "bool" && !f.Vector && baseType.Name == "true" {
			f.ConditionalBool = true
		}
	}
	return f, nil
}
