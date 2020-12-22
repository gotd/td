package gen

import (
	"golang.org/x/xerrors"

	"github.com/gotd/tl"
)

// fieldDef represents structDef field.
type fieldDef struct {
	// Name of field. Should be in camel case.
	Name string
	// Comment for field. Currently only one-line.
	Comment string
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
	// Vector denotes whether fieldDef TL type is vector.
	Vector bool
	// DoubleVector denotes whether fieldDef TL type is vector of vectors.
	DoubleVector bool
	// Slice denotes whether slice should be used for this field.
	Slice bool
	// DoubleSlice denotes whether double slicing should be used, e.g. [][]bytes.
	DoubleSlice bool
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

// nolint:gocognit,gocyclo
//
// TODO(ernado) Split into multiple sections: base type, encoder and conditional.
func (g *Generator) makeField(param tl.Parameter, annotations []tl.Annotation) (fieldDef, error) {
	f := fieldDef{
		Name:    pascal(param.Name),
		RawName: param.Name,
		RawType: param.Type.String(),
	}
	// Unwrapping up to 2 levels of vectors.
	baseType := param.Type
	for _, a := range annotations {
		if a.Name == param.Name {
			f.Comment = a.Value
		}
	}
	if baseType.Name == "vector" || baseType.Name == "Vector" {
		baseType = *baseType.GenericArg
		f.Vector = true
		f.Slice = true
	}
	if baseType.Name == "vector" || baseType.Name == "Vector" {
		baseType = *baseType.GenericArg
		f.DoubleSlice = true
		f.DoubleVector = true
	}
	if param.Flags {
		f.Type = "bin.Fields"
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
	case "long", "int53":
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
		if param.Flags {
			f.Type = "bin.Fields"
			break
		}

		if baseType.Bare {
			// Using exact go type for bare types.
			t, ok := g.types[baseType.String()]
			if !ok {
				return fieldDef{}, xerrors.Errorf("types[%s] not found", baseType)
			}
			f.Type = t.Name
		} else {
			// Type is generic.
			t, ok := g.classes[baseType.String()]
			if !ok {
				return fieldDef{}, xerrors.Errorf("classes[%s] not found", baseType)
			}
			f.Type = t.Name
			if !t.Singular {
				f.Interface = t.Name
				f.InterfaceFunc = t.Func
			}
		}
	}
	if flag := param.Flag; flag != nil {
		f.Conditional = true
		f.ConditionalIndex = flag.Index
		f.ConditionalField = pascal(flag.Name)
		if f.Type == "bool" && !f.Vector {
			f.ConditionalBool = true
		}
	}
	return f, nil
}
