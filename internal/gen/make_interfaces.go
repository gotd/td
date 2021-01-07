package gen

// interfaceDef represents generic interface, type which has multiple constructors.
type interfaceDef struct {
	// Name of interface.
	Name string
	// RawType is raw type from TL schema.
	RawType string

	// Constructors of interface.
	Constructors []structDef
	Func         string
	Namespace    []string
	BaseName     string
	URL          string
}

func (g *Generator) makeInterfaces() {
	// Make interfaces for classes.
	for _, c := range g.classes {
		if c.Singular {
			continue
		}
		def := interfaceDef{
			Name:      c.Name,
			Namespace: c.Namespace,
			Func:      c.Func,
			BaseName:  c.BaseName,
			RawType:   c.RawType,
			URL:       g.docURL("type", c.RawType),
		}
		for _, s := range g.structs {
			if s.Interface != def.Name {
				continue
			}
			def.Constructors = append(def.Constructors, s)
		}
		g.interfaces = append(g.interfaces, def)
	}
}
