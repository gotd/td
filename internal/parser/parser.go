package parser

import (
	"bufio"
	"io"
	"strings"

	"golang.org/x/xerrors"
)

// SchemaDefinition is annotated Definition with Category.
type SchemaDefinition struct {
	Annotations []Annotation `json:"annotations,omitempty"` // annotations (comments)
	Definition  Definition   `json:"definition"`            // definition
	Category    Category     `json:"category"`              // category of definition (function or type)
}

type Class struct {
	Name        string
	Description string
}

// Schema represents single TL file with information about definitions and
// so called "Classes" aka non-bare types with one or multiple constructors.
type Schema struct {
	Definitions []SchemaDefinition `json:"definitions"`
	Classes     []Class            `json:"classes,omitempty"`
}

const (
	vectorDefinition       = "vector {t:Type} # [ t ] = Vector t;"
	vectorDefinitionWithID = "vector#1cb5c415 {t:Type} # [ t ] = Vector t;"
)

func Parse(reader io.Reader) (*Schema, error) {
	var (
		def      SchemaDefinition
		line     int
		category Category

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
			category = CategoryFunction
			continue
		case tokTypes:
			category = CategoryType
			continue
		case vectorDefinition, vectorDefinitionWithID:
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
					if a.Name == "class" {
						class.Name = a.Value
					}
					if a.Name == "description" {
						class.Description = a.Value
					}
				}
				if class.Name != "" {
					schema.Classes = append(schema.Classes, class)
				}
				// Reset annotations so we don't include them to next type.
				ann = ann[:0]
			}

			def.Annotations = append(def.Annotations, ann...)
			continue
		}
		if strings.HasPrefix(s, "//") {
			continue
		}
		if err := def.Definition.Parse(s); err != nil {
			return nil, xerrors.Errorf("failed to parse line %d: definition: %w", line, err)
		}

		schema.Definitions = append(schema.Definitions, def)
		def = SchemaDefinition{
			Category: category,
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, xerrors.Errorf("failed to scan: %w", err)
	}

	// Remaining type.
	if def.Definition.ID != 0 {
		schema.Definitions = append(schema.Definitions, def)
	}

	return schema, nil
}
