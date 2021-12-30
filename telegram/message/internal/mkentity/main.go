package main

import (
	"flag"
	"reflect"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tdp"
	"github.com/gotd/td/telegram/message/internal/mkrun"
	"github.com/gotd/td/tg"
)

// Field represents type field.
type Field struct {
	// Name is Go name of field.
	Name string
	// Type is Go type of field.
	Type string
}

// Type represents generated type.
type Type struct {
	// Name is Go name of type.
	Name string
	// Fields is slice of type fields.
	Fields []Field
	// SchemaType is related schema type.
	SchemaType tdp.Type
}

var (
	constructors = tg.ClassConstructorsMap()
	create       = tg.TypesConstructorMap()
	templates    = map[string]string{
		"entity":  entityTmpl,
		"styling": stylingTmpl,
	}
)

type generator struct {
	template string
}

func (g *generator) Name() string {
	return "mkentity"
}

func (g *generator) Flags(set *flag.FlagSet) {
	set.StringVar(&g.template, "template", "entity", "template to use")
}

func (g *generator) Template() string {
	return templates[g.template]
}

func (g *generator) Data() (interface{}, error) {
	var types []Type
	for _, typeID := range constructors[tg.MessageEntityClassName] {
		v, ok := create[typeID]().(tdp.Object)
		if !ok {
			return nil, errors.Errorf("bad type %#x", typeID)
		}
		schemaType := v.TypeInfo()
		// Skip messageEntityMentionName because we should use inputMessageEntityMentionName.
		if schemaType.Name == "messageEntityMentionName" {
			continue
		}

		tv := reflect.TypeOf(v).Elem()

		var fields []Field
		for _, field := range schemaType.Fields {
			// These fields set by Formatter callee.
			if field.Name == "Offset" || field.Name == "Length" {
				continue
			}

			rf, ok := tv.FieldByName(field.Name)
			if !ok {
				return nil, errors.Errorf(
					"field of %q type %q not found",
					field.Name, schemaType.Name,
				)
			}
			fields = append(fields, Field{
				Name: field.Name,
				Type: rf.Type.String(),
			})
		}
		types = append(types, Type{
			Name:       tv.Name(),
			Fields:     fields,
			SchemaType: v.TypeInfo(),
		})
	}

	return types, nil
}

func main() {
	mkrun.Main(&generator{})
}
