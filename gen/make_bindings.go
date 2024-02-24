package gen

import (
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/tl"
)

type typeBinding struct {
	Namespace     []string
	Class         string // user.Auth, class in TL
	Name          string // go type name, like in type Name struct{}
	Interface     string // go type interface, like UserAuthClass
	InterfaceFunc string // go encoding/decoding function postfix, like UserAuth in DecodeUserAuth
	Method        string // go method name if function
}

type classBinding struct {
	// Name as used in go interface type definition,
	// i.e. type Name interface {}.
	Name string
	// Func is used as postfix for Decode and Encode functions.
	Func      string
	Namespace []string
	// Singular is special case for class where single constructor replaces
	// class.
	Singular     bool
	Vector       bool
	Constructors []string

	// BaseName is "Auth" for user.Auth interface.
	BaseName string
	// RawType of class from TL.
	RawType string
}

// namespacedName returns camel-case name with namespace prefix.
func namespacedName(name string, namespace []string) string {
	goName := pascal(name)
	if len(namespace) > 0 {
		var b strings.Builder
		for _, ns := range namespace {
			b.WriteString(pascal(ns))
		}
		b.WriteString(goName)
		goName = b.String()
	}
	return goName
}

// makeBindings fills classes and types fields of Generator.
func (g *Generator) makeBindings() error {
	// 1) Searching for all classes with single constructor.
	// If class has single constructor, it can be reduced to specific type.
	constructors := map[string]int{}
	for _, d := range g.schema.Definitions {
		if d.Category != tl.CategoryType {
			continue
		}
		constructors[d.Definition.Type.String()]++
	}

	// 2) Binding TL types to structures and interfaces.
	for _, sd := range g.schema.Definitions {
		var (
			d        = sd.Definition
			classKey = d.Type.String()
			typeKey  = definitionType(d)
			goName   = namespacedName(d.Name, d.Namespace)
		)

		// Binding bare type.
		tb := typeBinding{
			Namespace: d.Namespace,
			Class:     classKey,
			Name:      goName,
		}

		switch sd.Category {
		case tl.CategoryType:
			constructorsCount, ok := constructors[classKey]
			if constructorsCount == 0 || !ok {
				return errors.Errorf("constructors[%s] not found", classKey)
			}
			if constructorsCount == 1 {
				// Using this constructor instead of generic class for all definitions
				// that depends on that class.
				b := classBinding{
					Namespace: d.Namespace,
					Singular:  true,
					Name:      goName,
					BaseName:  d.Type.Name,
				}
				g.classes[classKey] = b
				g.types[typeKey] = tb
				continue
			}

			if _, ok := g.classes[classKey]; !ok {
				// interfaceDef has multiple constructors and is new.
				className := namespacedName(d.Type.Name, d.Type.Namespace)
				g.classes[classKey] = classBinding{
					Namespace: d.Namespace,
					Singular:  false,
					Func:      className,
					Name:      className + "Class",
					BaseName:  d.Type.Name,
					RawType:   d.Type.String(),
				}
			}

			c := g.classes[classKey]
			c.Constructors = append(c.Constructors, goName)
			g.classes[classKey] = c

			tb.Interface = c.Name
			tb.InterfaceFunc = c.Func
		case tl.CategoryFunction:
			// Just creating new bare type.
			tb.Method = tb.Name
			tb.Name += "Request"
		}
		g.types[typeKey] = tb
	}

	return nil
}
