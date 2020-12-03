package gen

import (
	"golang.org/x/xerrors"
)

// interfaceDef represents generic interface, type which has multiple constructors.
type interfaceDef struct {
	// Name of interface.
	Name string `validate:"required"`
	// RawType is raw type from TL schema.
	RawType string `validate:"required"`

	// Constructors of interface.
	Constructors []structDef `validate:"required"`
	Func         string      `validate:"required"`
	Namespace    []string    `validate:"required"`
	BaseName     string      `validate:"required"`
	URL          string      `validate:"omitempty"`
}

func (g *Generator) makeInterfaces() error {
	// Make interfaces for classes.
	for _, c := range g.classes {
		if c.Singular {
			continue
		}
		if err := g.validator.Struct(c); err != nil {
			return xerrors.Errorf("invalid class: %w", err)
		}
		def := interfaceDef{
			Name:      c.Name,
			Namespace: c.Namespace,
			Func:      c.Func,
			BaseName:  c.BaseName,
			RawType:   c.RawType,
			URL:       g.docURL("type", c.RawType),
		}
		for _, s := range g.structs {
			if s.Interface != def.Name {
				continue
			}
			def.Constructors = append(def.Constructors, s)
		}
		g.interfaces = append(g.interfaces, def)
	}
	return nil
}
