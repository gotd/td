package tdp

// Field of TL type, non-recursive.
type Field struct {
	Name       string
	SchemaName string
	Null       bool
}

// Type info for TL type, non-recursive.
type Type struct {
	// Name in TL schema.
	Name string
	// ID is type id.
	ID uint32
	// Fields of type.
	Fields []Field
	// Null denotes whether value is null.
	Null bool
}

// Object of TL schema that can return type info.
type Object interface {
	TypeInfo() Type
}
