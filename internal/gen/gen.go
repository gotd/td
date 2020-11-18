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
}

// Argument of interface method.
type Argument struct {
	Name string
	Type string
}

// Result of Method.
type Result struct {
	Blank bool
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
		constructors[d.Definition.Type.Name] += 1
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
			s := Struct{
				Name:     pascal(d.Definition.Name),
				Receiver: strings.ToLower(d.Definition.Name[0:1]),
				HexID:    fmt.Sprintf("%x", d.Definition.ID),
				BufArg:   "b",
				TLType:   fmt.Sprintf("%s#%x", d.Definition.Name, d.Definition.ID),
			}
			if s.Receiver == "b" {
				// bin.Buffer argument collides with reciever.
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
				switch param.Type.Name {
				case "int":
					f.Func = "Int"
				case "int32":
					f.Func = "Int32"
				case "string":
					f.Func = "String"
				case "Bool":
					f.Func = "Bool"
					f.Type = "bool"
				default:
					f.Encoder = true
				}
				if f.Comment == "" {
					f.Comment = fmt.Sprintf("%s field of %s.", f.Name, s.Name)
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
