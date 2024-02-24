package gen

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-faster/errors"

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
	docBase      *url.URL
	doc          *getdoc.Doc
	docLineLimit int

	generateFlags GenerateFlags
}

// NewGenerator initializes and returns new Generator from tl.Schema.
func NewGenerator(s *tl.Schema, genOpt GeneratorOptions) (*Generator, error) {
	genOpt.setDefaults()
	g := &Generator{
		schema:        s,
		classes:       map[string]classBinding{},
		types:         map[string]typeBinding{},
		mappings:      map[string][]constructorMapping{},
		docLineLimit:  genOpt.DocLineLimit,
		generateFlags: genOpt.GenerateFlags,
	}
	if genOpt.DocBaseURL != "" {
		u, err := url.Parse(genOpt.DocBaseURL)
		if err != nil {
			return nil, errors.Wrap(err, "parse docBase")
		}
		g.docBase = u

		if u.Host == "core.telegram.org" {
			// Using embedded documentation.
			// TODO(ernado): Get actual layer
			doc, err := getdoc.Load(getdoc.LayerLatest)
			if err != nil {
				return nil, errors.Wrap(err, "get documentation")
			}
			g.doc = doc
		}
	}
	if err := g.makeBindings(); err != nil {
		return nil, errors.Wrap(err, "make type bindings")
	}
	if err := g.makeStructures(); err != nil {
		return nil, errors.Wrap(err, "generate go structures")
	}
	g.makeInterfaces()
	g.makeErrors()

	return g, nil
}
