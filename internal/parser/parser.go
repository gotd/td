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

type SchemaType struct {
	Annotations []Annotation
	Definition  Definition
	Kind        Kind
}

type Class struct {
	Name        string
	Description string
}

type Schema struct {
	Types   []SchemaType
	Classes []Class
}

type section byte

const (
	sectionTypes section = iota
	sectionFunctions
)

func Parse(reader io.Reader) (*Schema, error) {
	var (
		typ  SchemaType // current type
		line int        // current line
		sec  section    // current section

		schema  = &Schema{}
		scanner = bufio.NewScanner(reader)
	)
	for scanner.Scan() {
		line++
		s := strings.TrimSpace(scanner.Text())
		s = strings.ReplaceAll(s, "///", "//") // normalize comments
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
			if strings.HasPrefix(s, "//@class") {
				var class Class
				for _, a := range ann {
					if a.Key == "class" {
						class.Name = a.Value
					}
					if a.Key == "description" {
						class.Description = a.Value
					}
				}
				if class.Name != "" {
					schema.Classes = append(schema.Classes, class)
				}
				// Reset annotations so we don't include them to next type.
				ann = ann[:0]
			}

			typ.Annotations = append(typ.Annotations, ann...)
			continue
		}
		if strings.HasPrefix(s, "//") {
			continue
		}

		def, err := parseDefinition(s)
		if err != nil {
			return nil, xerrors.Errorf("failed to parse line %d: definition: %w", line, err)
		}

		typ.Definition = def
		schema.Types = append(schema.Types, typ)
		typ = SchemaType{
			Kind: map[section]Kind{
				sectionTypes:     KindType,
				sectionFunctions: KindFunction,
			}[sec],
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, xerrors.Errorf("failed to scan: %w", err)
	}

	// Remaining type.
	if typ.Definition.ID != 0 {
		schema.Types = append(schema.Types, typ)
	}

	return schema, nil
}
