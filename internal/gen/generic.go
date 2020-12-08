package gen

import (
	"strings"

	"github.com/gotd/tl"
)

func (g *Generator) instantiateVector(className string) (class classBinding, err error) {
	class, ok := g.classes[className]
	if ok {
		return class, nil
	}

	elementName := strings.TrimPrefix(className[:len(className)-1], "Vector<")
	goElementName := g.classes[elementName].Name
	if goElementName == "" {
		goElementName = elementName
	}

	f, err := g.makeField(tl.Parameter{
		Name: "Elems",
		Type: tl.Type{
			Name: "Vector",
			GenericArg: &tl.Type{
				Name: elementName,
			},
		},
	}, nil)
	if err != nil {
		return
	}

	goName := strings.Title(goElementName) + "Vector"
	class = classBinding{
		Name:     goName,
		Func:     f.Type,
		Singular: true,
		Vector:   true,
	}
	g.classes[className] = class

	g.structs = append(g.structs, structDef{
		Name:     goName,
		Receiver: "vec",
		BufArg:   "b",
		Vector:   true,
		RawType:  className,
		Fields:   []fieldDef{f},
		BaseName: goName,
	})

	return class, nil
}
