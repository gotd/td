package gen

// interfaceDef represents generic interface, type which has multiple constructors.
type interfaceDef struct {
	// Name of interface.
	Name string
	// RawType is raw type from TL schema.
	RawType string

	// Fields, common for every constructor.
	SharedFields []fieldDef
	// Constructors of interface.
	Constructors []structDef
	Func         string
	Namespace    []string
	BaseName     string
	URL          string
}

func commonFields(a, b []fieldDef) []fieldDef {
	keep := func(f fieldDef, b []fieldDef) bool {
		for _, x := range b {
			if x.EqualAsField(f) {
				return true
			}
		}

		return false
	}

	n := 0
	for _, x := range a {
		if keep(x, b) {
			a[n] = x
			n++
		}
	}
	a = a[:n]

	return a
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

			if def.SharedFields == nil {
				def.SharedFields = make([]fieldDef, len(s.Fields))
				copy(def.SharedFields, s.Fields)
			} else {
				def.SharedFields = commonFields(def.SharedFields, s.Fields)
			}

			def.Constructors = append(def.Constructors, s)
		}

		n := 0
		for _, f := range def.SharedFields {
			if f.Type != flagsType {
				def.SharedFields[n] = f
				n++
			}
		}
		def.SharedFields = def.SharedFields[:n]

		g.interfaces = append(g.interfaces, def)
	}
}
