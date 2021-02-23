package gen

import (
	"strings"
)

type constructorMapping struct {
	// Name is go name of interface or struct.
	Name string
	// Constructor is go name of mapped constructor.
	// May be empty.
	Constructor string
	// Concrete is flag which is true when Name address a struct, not interface.
	Concrete bool
	// MapperName is name of mapper which created this sub.
	MapperName string
	// Fields is slice of field mappings from this struct to target.
	Fields []fieldPair
}

// interfaceDef represents generic interface, type which has multiple constructors.
type interfaceDef struct {
	// Name of interface.
	Name string
	// RawType is raw type from TL schema.
	RawType string

	// Fields, common for every constructor.
	SharedFields map[string][]fieldDef
	// Sub interfaces of this TL class.
	// Need to create As${sub.MapperName}() ${sub.Name} mappers.
	Mappings []constructorMapping
	// Constructors of interface.
	Constructors []structDef
	Func         string
	Namespace    []string
	BaseName     string
	URL          string
}

func interfaceHasOneSuffix(suffixes ...string) func(s structDef) bool {
	return func(s structDef) bool {
		for _, suffix := range suffixes {
			if strings.HasSuffix(s.Name, suffix) {
				return true
			}
		}
		return false
	}
}

func makeMapping(def *interfaceDef, name string, emptyFilter func(s structDef) bool) {
	var (
		// Fields, common for every non-empty constructor.
		nonEmptyFields []fieldDef
		// Non-empty constructors.
		nonEmptyConstructors []structDef
		// Index of optional empty constructor.
		emptyIdx = -1
	)
	for _, s := range def.Constructors {
		if emptyFilter(s) {
			emptyIdx = len(def.Constructors)
		} else {
			nonEmptyFields = intersectFields(nonEmptyFields, s.Fields)
			nonEmptyConstructors = append(nonEmptyConstructors, s)
		}
	}

	// If have at least one empty constructor.
	hasEmpty := emptyIdx >= 0

	// If all non-empty constructors have common fields.
	nonEmptyHasCommonFields := len(nonEmptyFields) > 0

	if hasEmpty && nonEmptyHasCommonFields && len(nonEmptyConstructors) > 0 {
		def.SharedFields[name] = nonEmptyFields

		// If there are only one non-empty constructor, so we use concrete type.
		concrete := len(nonEmptyConstructors) < 2
		goName := name + strings.TrimSuffix(def.Name, "Class")
		if concrete {
			goName = nonEmptyConstructors[0].Name
		}

		mapping := constructorMapping{
			Name:       goName,
			Concrete:   concrete,
			MapperName: name,
		}
		def.Mappings = append(def.Mappings, mapping)
	}
}

func (g *Generator) collectMappings(def *interfaceDef) {
	for _, s := range g.structs {
		// Filter constructors from same Class and empty constructors.
		if s.Interface == def.Name || len(s.Fields) < 1 {
			continue
		}

		for _, constructor := range def.Constructors {
			// Filter constructors which have not similar name.
			if !strings.HasPrefix(s.Name, "Input") || !strings.Contains(s.Name, constructor.Name) {
				continue
			}

			mapping, ok := mappableFields(constructor, s)
			// Filter constructors which we can't fill completely.
			if !ok {
				continue
			}
			def.Mappings = append(def.Mappings, mapping)
		}
	}

	emptyAnnotations := []struct {
		name   string
		filter func(s structDef) bool
	}{
		{"NotEmpty", interfaceHasOneSuffix("Empty")},
		{"Modified", interfaceHasOneSuffix("NotModified")},
		{"Available", interfaceHasOneSuffix("Unavailable")},
		{"NotForbidden", interfaceHasOneSuffix("Forbidden")},
		{"Full", interfaceHasOneSuffix("Empty", "NotModified", "Forbidden")},
	}
	for _, annotation := range emptyAnnotations {
		// Full annotation is necessary only if there are more than two empty annotations.
		if annotation.name == "Full" && len(def.SharedFields) <= 2 {
			continue
		}
		makeMapping(def, annotation.name, annotation.filter)
	}
}

func (g *Generator) makeInterfaces() {
	// Make interfaces for classes.
	for _, c := range g.classes {
		if c.Singular {
			continue
		}
		def := interfaceDef{
			Name:         c.Name,
			Namespace:    c.Namespace,
			Func:         c.Func,
			BaseName:     c.BaseName,
			RawType:      c.RawType,
			SharedFields: map[string][]fieldDef{},
			URL:          g.docURL("type", c.RawType),
		}

		for _, s := range g.structs {
			if s.Interface != def.Name {
				continue
			}

			def.SharedFields["Common"] = intersectFields(def.SharedFields["Common"], s.Fields)
			def.Constructors = append(def.Constructors, s)
		}
		g.collectMappings(&def)

		g.interfaces = append(g.interfaces, def)
		g.mappings[def.Name] = append(g.mappings[def.Name], def.Mappings...)
	}
}
