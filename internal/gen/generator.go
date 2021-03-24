package gen

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/getdoc"
	"github.com/gotd/tl"
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
	// errorChecks definitions.
	errorChecks []errCheckDef

	// constructor mappings.
	mappings map[string][]constructorMapping

	// registry of type ids.
	registry []bindingDef

	// docBase is base url for documentation.
	docBase *url.URL
	doc     *getdoc.Doc
	docLineLimit int
}

// NewGenerator initializes and returns new Generator from tl.Schema.
//
// The docBase value is base url for documentation, like:
// 	* https://core.telegram.org/
// If blank string provided, no documentation links are generated.
func NewGenerator(s *tl.Schema, docBase string) (*Generator, error) {
	g := &Generator{
		schema:   s,
		classes:  map[string]classBinding{},
		types:    map[string]typeBinding{},
		mappings: map[string][]constructorMapping{},
		docLineLimit: 87,
	}
	if docBase != "" {
		u, err := url.Parse(docBase)
		if err != nil {
			return nil, xerrors.Errorf("failed to parse docBase: %w", err)
		}
		g.docBase = u

		if u.Host == "core.telegram.org" {
			// Using embedded documentation.
			// TODO(ernado): Get actual layer
			doc, err := getdoc.Load(getdoc.LayerLatest)
			if err != nil {
				return nil, xerrors.Errorf("failed to get documentation: %w", err)
			}
			g.doc = doc
		}
	}
	if err := g.makeBindings(); err != nil {
		return nil, xerrors.Errorf("failed to make type bindings: %w", err)
	}
	if err := g.makeStructures(); err != nil {
		return nil, xerrors.Errorf("failed to generate go structures: %w", err)
	}
	g.makeInterfaces()
	g.makeErrors()

	return g, nil
}
