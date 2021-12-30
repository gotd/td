package mkrun

import "flag"

// Generator represents generator script.
type Generator interface {
	// Name is generator name.
	Name() string
	// Flags sets generator flags.
	Flags(set *flag.FlagSet)
	// Template returns generation template.
	Template() string
	// Data returns associated generation data.
	Data() (interface{}, error)
}
