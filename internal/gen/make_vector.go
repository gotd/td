package gen

import (
	"strings"

	"github.com/gotd/tl"
)

func (g *Generator) makeVector(className string) (class classBinding, err error) {
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
		return class, err
	}
	f.Comment = []string{"Elements of " + className}

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
		RawType:  className,
		Vector:   true,
		Fields:   []fieldDef{f},
		BaseName: goName,
		Comment:  goName + " is a box for " + className,
	})

	return class, nil
}
