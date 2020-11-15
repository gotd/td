package parser

// Category of Definition.
type Category byte

const (
	CategoryType Category = iota
	CategoryFunction
)

func (c Category) String() string {
	switch c {
	case CategoryFunction:
		return "function"
	default:
		return "type"
	}
}

func (c *Category) UnmarshalText(text []byte) error {
	switch string(text) {
	case "function":
		*c = CategoryFunction
	default:
		*c = CategoryType
	}
	return nil
}

func (c Category) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}
