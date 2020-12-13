// Package tmap provides type mapping facility that maps type id to type name.
package tmap

// Map is type mapping.
type Map struct {
	types map[uint32]string
}

func (m *Map) add(mapping map[uint32]string) {
	for k, v := range mapping {
		m.types[k] = v
	}
}

// Get returns type string or blank.
func (m *Map) Get(id uint32) string {
	if m == nil || len(m.types) == 0 {
		return ""
	}
	return m.types[id]
}

// New creates new Map from mappings.
func New(mappings ...map[uint32]string) *Map {
	m := &Map{
		types: map[uint32]string{},
	}
	for _, mapping := range mappings {
		m.add(mapping)
	}
	return m
}
