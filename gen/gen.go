package gen

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/ernado/tl"
)

type Config struct {
	Package    string
	Structs    []Struct
	Interfaces []Class
	Methods    []Method
}

// Struct represents go structure definition.
type Struct struct {
	// Name of struct, just like that: `type Name struct {`.
	Name string
	// Comment for struct, in one line.
	Comment string
	// Receiver name. E.g. "m" for Message.
	Receiver string
	// HexID is hex-encoded id, like 1ef134.
	HexID string
	// BufArg is name of Encode and Decode argument of bin.Buffer type
	// that is used in those functions.
	//
	// Should not equal to Name.
	BufArg string
	// TLType is type name from TL schema.
	TLType string

	// Interface refers to interface of generic type.
	Interface string
	// Constructor denotes whether Struct is constructor for some
	// generic type. If false, Interface is blank.
	Constructor bool

	// Fields of structure.
	Fields []Field
}

// Field represents go Struct field.
type Field struct {
	// Name of field. Should be in camel case.
	Name string
	// Comment for field. Currently only one-line.
	Comment string
	// Type is go type for field.
	Type string
	// Func is name for bin.* functions, e.g. String will render
	// to bin.Buffer.String and bin.Buffer.PutString.
	Func string
	// Encoder denotes whether Field implements bin.Encoder and bin.Decoder.
	Encoder bool
	// TLName is raw name from TL Schema.
	TLName string
	// Vector denotes whether Field TL type is vector.
	Vector bool
	// DoubleVector denotes whether Field TL type is vector of vectors.
	DoubleVector bool
	// Slice denotes whether slice should be used for this field.
	Slice bool
	// DoubleSlice denotes whether double slicing should be used, e.g. [][]bytes.
	DoubleSlice bool
	// Generic denotes whether Field Type has generic constructors.
	Generic bool
	// Conditional denotes whether Field is conditional.
	Conditional bool
	// ConditionalField if name of bitset param.
	ConditionalField string
	// ConditionalIndex is Field bit in ConditionalField.
	ConditionalIndex int
	// ConditionalBool denotes whether value is fully encoded in ConditionalField as bit.
	ConditionalBool bool
}

// Argument of interface method.
type Argument struct {
	Name  string
	Type  string
	Slice bool
}

// Result of Method.
type Result struct {
	Blank     bool
	Type      string
	Interface bool
	Slice     bool
}

// Method represents RPC method with Name, Arguments and Result.
type Method struct {
	Name      string
	Arguments []Argument
	Result    Result
}

// Class represents generic interface, type which has multiple constructors.
type Class struct {
	// Name of interface.
	Name string
	// Constructors of interface.
	Constructors []Struct
}

// Generate generates go code based on provided TL schema.
//
// nolint:goconst,gocognit,gocyclo
func Generate(w io.Writer, t *template.Template, s *tl.Schema) error {
	cfg := Config{
		Package: "td",
	}

	// Searching for all types with single constructor.
	// This can be used to reduce interfaces.
	constructors := map[string]int{}
	for _, d := range s.Definitions {
		if d.Category != tl.CategoryType {
			continue
		}
		// TODO: Namespaces?
		constructors[d.Definition.Type.Name]++
	}
	singular := map[string]struct{}{}
	for k, v := range constructors {
		if v == 1 {
			singular[k] = struct{}{}
		}
	}

	classes := map[string][]Struct{}
	var classNames []string
	for _, d := range s.Definitions {
		switch d.Category {
		case tl.CategoryType:
			// Type definition.
			s := Struct{
				Name:     pascal(d.Definition.Name),
				Receiver: strings.ToLower(d.Definition.Name[0:1]),
				HexID:    fmt.Sprintf("%x", d.Definition.ID),
				BufArg:   "b",
				TLType:   fmt.Sprintf("%s#%x", d.Definition.Name, d.Definition.ID),
			}
			if s.Receiver == "b" {
				// bin.Buffer argument collides with receiver.
				s.BufArg = "buf"
			}
			for _, a := range d.Annotations {
				if a.Name == tl.AnnotationDescription {
					s.Comment = a.Value
				}
			}
			if s.Comment == "" {
				s.Comment = fmt.Sprintf("%s represents TL type %s#%x.",
					s.Name,
					d.Definition.Name,
					d.Definition.ID,
				)
			}
			for _, param := range d.Definition.Params {
				f := Field{
					Name:   pascal(param.Name),
					Type:   param.Type.Name,
					TLName: param.Name,
				}
				for _, a := range d.Annotations {
					if a.Name == param.Name {
						f.Comment = a.Value
					}
				}
				if f.Type == "vector" {
					f.Type = param.Type.GenericArg.Name
					f.Vector = true
					f.Slice = true
				}
				if f.Type == "vector" {
					f.Type = param.Type.GenericArg.GenericArg.Name
					f.DoubleSlice = true
					f.DoubleVector = true
				}
				if param.Flags {
					f.Type = "bin.Fields"
				}
				switch f.Type {
				case "int":
					f.Func = "Int"
				case "int32":
					f.Func = "Int32"
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
					f.Slice = true
					if param.Type.Name == "vector" {
						f.DoubleSlice = true
						f.Vector = true
					}
				default:
					f.Encoder = true
					if _, ok := singular[param.Type.Name]; !ok && !param.Flags {
						f.Generic = true
					}
					if param.Type.Name == "vector" {
						if param.Type.GenericArg.Bare {
							f.Type = pascal(f.Type)
							f.Generic = false
						}
					} else if param.Type.Bare {
						f.Type = pascal(f.Type)
						f.Generic = false
					}
				}
				if f.Comment == "" {
					f.Comment = fmt.Sprintf("%s field of %s.", f.Name, s.Name)
				}
				if flag := param.Flag; flag != nil {
					f.Conditional = true
					f.ConditionalIndex = flag.Index
					f.ConditionalField = pascal(flag.Name)
					if f.Type == "bool" && !f.Vector {
						f.ConditionalBool = true
					}
				}
				s.Fields = append(s.Fields, f)
			}
			if _, ok := singular[d.Definition.Type.Name]; !ok {
				name := d.Definition.Type.Name
				if _, ok := classes[name]; !ok {
					classNames = append(classNames, name)
				}
				classes[name] = append(classes[name], s)
				s.Constructor = true
				s.Interface = pascal(name)
			}
			cfg.Structs = append(cfg.Structs, s)
		case tl.CategoryFunction:
			// RPC call definition.
			m := Method{
				Name:   pascal(d.Definition.Name),
				Result: Result{},
			}
			if d.Definition.Type.Name == "Ok" {
				m.Result.Blank = true
			} else {
				m.Result.Type = pascal(d.Definition.Type.Name)
				if m.Result.Type == "vector" {
					m.Result.Type = pascal(d.Definition.Type.GenericArg.Name)
					m.Result.Slice = true
				}
				if _, ok := singular[d.Definition.Type.Name]; !ok {
					m.Result.Interface = true
				}
			}
			// Arguments of definition.
			for _, param := range d.Definition.Params {
				arg := Argument{
					Name: camel(param.Name),
					Type: param.Type.Name,
				}
				switch arg.Name {
				case "type":
					arg.Name = "typ"
				}
				if arg.Type == "vector" {
					arg.Type = param.Type.GenericArg.Name
					arg.Slice = true
				}
				switch arg.Type {
				case "int":
					arg.Type = "int"
				case "int32":
					arg.Type = "int32"
				case "string":
					arg.Type = "string"
				case "long", "int53":
					arg.Type = "int64"
				case "double":
					arg.Type = "float64"
				case "Bool", "bool", "true", "false":
					arg.Type = "bool"
				default:
					arg.Type = pascal(param.Type.Name)
				}
				m.Arguments = append(m.Arguments, arg)
			}
			cfg.Methods = append(cfg.Methods, m)
		}
	}
	for _, name := range classNames {
		cfg.Interfaces = append(cfg.Interfaces, Class{
			Name:         pascal(name),
			Constructors: classes[name],
		})
	}
	return t.ExecuteTemplate(w, "simple", cfg)
}
