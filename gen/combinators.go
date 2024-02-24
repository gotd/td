package gen

func generateSliceHelper(s structDef) bool {
	return len(s.Fields) > 0 && !s.Vector && s.Method == ""
}

type simpleField struct {
	Name        string
	Type        string
	Conditional bool
}

func (s simpleField) match(f fieldDef) bool {
	return f.Name == s.Name && f.Type == s.Type && f.Conditional == s.Conditional
}

func collectOnlyFields(fields []fieldDef, matchers ...simpleField) []fieldDef {
	return filterFieldsTo(fields, nil, func(f fieldDef) bool {
		for _, matcher := range matchers {
			if matcher.match(f) {
				return true
			}
		}
		return false
	})
}

func mapCollectableFields(fields []fieldDef) (r []fieldDef) {
	return collectOnlyFields(fields,
		simpleField{Name: "ID", Type: "int"},
		simpleField{Name: "ID", Type: "int64"},
	)
}

func sortableFields(fields []fieldDef) (r []fieldDef) {
	return collectOnlyFields(fields,
		simpleField{Name: "ID", Type: "int"},
		simpleField{Name: "ID", Type: "int64"},
		simpleField{Name: "Date", Type: "int"},
	)
}
