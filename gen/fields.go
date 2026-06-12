package gen

import (
	"strings"
)

func optionalField(_ structDef, f fieldDef) bool {
	return f.Conditional
}

// parameterField reports whether an unmapped field of the target constructor
// should be exposed as a parameter of the generated As-mapper instead of
// causing the mapping to be skipped.
//
// thumb_size on inputDocumentFileLocation/inputPhotoFileLocation is required by
// the wire format but cannot be derived from the source constructor, so the
// caller selects the thumbnail by passing it in. See issue #376.
func parameterField(_ structDef, f fieldDef) bool {
	return !f.Conditional && f.Type == "string" && f.RawName == "thumb_size"
}

func hasField(fields []fieldDef, name, typ string) bool {
	for _, f := range fields {
		if f.Name == name && f.Type == typ {
			return true
		}
	}

	return false
}

func mappableFields(constructor, to structDef) (constructorMapping, bool) {
	var r []fieldPair
	mapped := map[string]struct{}{}
	for _, a := range constructor.Fields {
		if a.Type == flagsType {
			continue
		}

		for _, b := range to.Fields {
			if b.Type == flagsType {
				continue
			}

			if a.SameType(b) && strings.Contains(b.Name, a.Name) {
				r = append(r, fieldPair{a, b})
				mapped[b.Name] = struct{}{}
			}
		}
	}

	// Fields that can't be derived from the constructor but should still be
	// filled by the caller via generated method parameters.
	var params []fieldDef
	// Return false if we can't fill all fields.
	if len(mapped) != len(to.Fields) {
		for _, field := range to.Fields {
			if _, ok := mapped[field.Name]; ok {
				continue
			}
			switch {
			case optionalField(to, field):
				// Optional in the wire format, left unset.
			case parameterField(to, field):
				params = append(params, field)
			default:
				return constructorMapping{}, false
			}
		}
	}

	mapperName := to.Name
	// Mapping: User => InputUser, so mapperName = "Input"
	// Mapping: Document => InputDocumentFileLocation, so mapperName = "InputDocumentFileLocation"
	if strings.HasSuffix(to.Name, constructor.Name) {
		mapperName = strings.TrimSuffix(to.Name, constructor.Name)
	}

	mapping := constructorMapping{
		Name:        to.Name,
		Constructor: constructor.Name,
		Concrete:    true,
		MapperName:  mapperName,
		Fields:      r,
		Params:      params,
	}

	return mapping, true
}

func intersectFields(a, b []fieldDef) []fieldDef {
	return intersectFieldsBy(a, b, fieldDef.EqualAsField)
}

func intersectFieldsBy(a, b []fieldDef, compare func(a, b fieldDef) bool) []fieldDef {
	// If a is empty, copy all from b to a.
	if a == nil {
		a = make([]fieldDef, len(b))
		copy(a, b)
	} else { // Otherwise intersect.
		a = commonFields(a, b, compare)
	}

	return filterFields(a, func(def fieldDef) bool {
		// Filter bin.Flags fields.
		return def.Type != flagsType
	})
}

func commonFields(a, b []fieldDef, compare func(a, b fieldDef) bool) []fieldDef {
	return filterFields(a, func(def fieldDef) bool {
		for _, x := range b {
			if compare(x, def) {
				return true
			}
		}

		return false
	})
}

func filterFields(a []fieldDef, filter func(def fieldDef) bool) []fieldDef {
	n := 0
	for _, f := range a {
		if filter(f) {
			a[n] = f
			n++
		}
	}
	a = a[:n]

	return a
}

func filterFieldsTo(a, b []fieldDef, filter func(def fieldDef) bool) []fieldDef {
	for _, f := range a {
		if filter(f) {
			b = append(b, f)
		}
	}

	return b
}
