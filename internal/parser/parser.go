package parser

import (
	"bufio"
	"io"
	"strings"

	"golang.org/x/xerrors"
)

type Kind byte

const (
	KindFunction Kind = iota
	KindType
)

type Type struct {
	Annotations []Annotation
	Definition  Definition
	Kind        Kind
}

type Schema struct {
	Types map[string]Type
}

type section byte

const (
	sectionDefinitions section = iota
	sectionFunctions
	sectionTypes
)

func Parse(reader io.Reader) (*Schema, error) {
	scanner := bufio.NewScanner(reader)
	var (
		typ  Type
		line int
		sec  section
	)
	schema := &Schema{
		Types: map[string]Type{},
	}
	for scanner.Scan() {
		line++
		s := strings.TrimSpace(scanner.Text())
		switch s {
		case "":
			continue
		case tokFunctions:
			sec = sectionFunctions
			continue
		case tokTypes:
			sec = sectionTypes
			continue
		case "vector {t:Type} # [ t ] = Vector t;":
			// Special case for vector
			continue
		}
		if strings.HasPrefix(s, "//@") {
			ann, err := parseAnnotation(s)
			if err != nil {
				return nil, xerrors.Errorf("failed to parse line %d: %w", line, err)
			}
			typ.Annotations = append(typ.Annotations, ann...)
			continue
		}
		if strings.HasPrefix(s, "//") {
			continue
		}

		// New type definition.
		if typ.Definition.ID != 0 {
			schema.Types[typ.Definition.Name] = typ
			typ = Type{
				Kind: map[section]Kind{
					sectionTypes:     KindType,
					sectionFunctions: KindFunction,
				}[sec],
			}
		}

		def, err := parseDefinition(s)
		if err != nil {
			return nil, xerrors.Errorf("failed to parse line %d: definition: %w", line, err)
		}

		typ.Definition = def
	}

	// Last type.
	if typ.Definition.ID != 0 {
		schema.Types[typ.Definition.Name] = typ
	}

	return schema, nil
}
