package gen

import (
	"fmt"
	"strings"

	"github.com/ernado/tl"
	"golang.org/x/xerrors"
)

// structDef represents go structure definition.
type structDef struct {
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
	// RawType is type name from TL schema.
	RawType string

	// Interface refers to interface of generic type.
	Interface     string
	InterfaceFunc string

	// Fields of structure.
	Fields []fieldDef

	// Namespace for file structure generation.
	Namespace []string
	// BaseName for file structure generation.
	BaseName string
}

// makeStructures generates go structure definition representations.
func (g *Generator) makeStructures() error {
	for _, sd := range g.schema.Definitions {
		if sd.Category != tl.CategoryType {
			continue
		}
		var (
			d       = sd.Definition
			typeKey = definitionType(d)
		)
		t, ok := g.types[typeKey]
		if !ok {
			return xerrors.Errorf("failed to find type binding for %q", typeKey)
		}
		s := structDef{
			Namespace: t.Namespace,
			Name:      t.Name,
			BaseName:  d.Name,

			HexID:   fmt.Sprintf("%x", d.ID),
			BufArg:  "b",
			RawType: fmt.Sprintf("%s#%x", typeKey, d.ID),

			Interface:     t.Interface,
			InterfaceFunc: t.InterfaceFunc,
		}
		// Selecting receiver based on non-namespaced type.
		s.Receiver = strings.ToLower(d.Name[:1])
		if s.Receiver == "b" {
			// bin.Buffer argument collides with receiver.
			s.BufArg = "buf"
		}
		if s.Comment == "" {
			// TODO(ernado): multi-line comments.
			s.Comment = fmt.Sprintf("%s represents TL type `%s`.", s.Name, s.RawType)
		}
		for _, param := range d.Params {
			f, err := g.makeField(param, sd.Annotations)
			if err != nil {
				return xerrors.Errorf("failed to make field %s: %w", param.Name, err)
			}
			if f.Comment == "" {
				f.Comment = fmt.Sprintf("%s field of %s.", f.Name, s.Name)
			}
			s.Fields = append(s.Fields, f)
		}

		g.structs = append(g.structs, s)
	}

	return nil
}
