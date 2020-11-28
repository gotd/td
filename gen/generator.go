package gen

import (
	"fmt"
	"strings"

	"github.com/ernado/tl"
	"golang.org/x/xerrors"
)

func definitionType(d tl.Definition) string {
	if len(d.Namespace) == 0 {
		return d.Name
	}
	return fmt.Sprintf("%s.%s", strings.Join(d.Namespace, "."), d.Name)
}

// Generator generates go types from tl.Schema.
type Generator struct {
	schema *tl.Schema

	// classes type bindings, key is TL type.
	classes map[string]classBinding
	// types bindings, key is TL type.
	types map[string]typeBinding

	// structs definitions.
	structs []structDef
	// interfaces definitions.
	interfaces []interfaceDef
}

// NewGenerator initializes and returns new Generator from tl.Schema.
func NewGenerator(s *tl.Schema) (*Generator, error) {
	g := &Generator{
		schema:  s,
		classes: map[string]classBinding{},
		types:   map[string]typeBinding{},
	}
	if err := g.makeBindings(); err != nil {
		return nil, xerrors.Errorf("failed to make type bindings: %w", err)
	}
	if err := g.makeStructures(); err != nil {
		return nil, xerrors.Errorf("failed to generate go structures: %w", err)
	}
	if err := g.makeInterfaces(); err != nil {
		return nil, xerrors.Errorf("failed go generate go interfaces: %w", err)
	}
	return g, nil
}
