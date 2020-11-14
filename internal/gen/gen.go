package gen

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/ernado/td/internal/parser"
)

type Config struct {
	Package string
}

type Field struct {
	Name    string
	Type    string
	Comment string
}

type Type struct {
	HexID    string
	Comment  string
	Name     string
	Receiver string
	Fields   []Field
}

type Context struct {
	Header  string
	Config  Config
	Types   []Type
	Methods []Method
}

type Method struct {
	Comment   string
	Name      string
	Arguments []Field
	Output    string
}

func Generate(w io.Writer, t *template.Template, s *parser.Schema) error {
	cfg := Config{
		Package: "td",
	}
	renderCtx := Context{
		Config: cfg,
	}
	for _, typ := range s.Types {
		switch typ.Kind {
		case parser.KindFunction:
			m := Method{
				Name:   camel(typ.Definition.Name),
				Output: typ.Definition.Interface,
			}
			if m.Output == "Ok" {
				// Special case.
				m.Output = ""
			}
			argComments := make(map[string]string)
			for _, a := range typ.Annotations {
				if a.Key == "description" {
					m.Comment = a.Value
				} else {
					argComments[a.Key] = a.Value
				}
			}
			for _, f := range typ.Definition.Fields {
				m.Arguments = append(m.Arguments, Field{
					Name:    camel(f.Name),
					Type:    f.Type,
					Comment: argComments[f.Name],
				})

			}
			if m.Comment == "" {
				m.Comment = m.Name + " implements TL RPC for type " + typ.Definition.Name + "."
			}
			if !strings.HasSuffix(m.Comment, ".") {
				m.Comment = m.Comment + "."
			}
			renderCtx.Methods = append(renderCtx.Methods, m)
		case parser.KindType:
			t := Type{
				Name:  pascal(typ.Definition.Name),
				HexID: fmt.Sprintf("%x", typ.Definition.ID),
			}
			fieldComments := make(map[string]string)
			for _, a := range typ.Annotations {
				if a.Key == "description" {
					t.Comment = a.Value
				} else {
					fieldComments[a.Key] = a.Value
				}
			}
			for _, f := range typ.Definition.Fields {
				t.Fields = append(t.Fields, Field{
					Name:    pascal(f.Name),
					Type:    f.Type,
					Comment: fieldComments[f.Name],
				})
			}
			if t.Comment == "" {
				t.Comment = t.Name + " implements TL for type " + typ.Definition.Name + "."
			}
			if !strings.HasSuffix(t.Comment, ".") {
				t.Comment = t.Comment + "."
			}
			renderCtx.Types = append(renderCtx.Types, t)
		}
	}
	return t.ExecuteTemplate(w, "simple", renderCtx)
}
