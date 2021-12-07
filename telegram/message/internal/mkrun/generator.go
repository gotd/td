package mkrun

type Generator interface {
	// Name is generator name.
	Name() string
	// Template returns generation template.
	Template() string
	// Data returns associated generation data.
	Data() (interface{}, error)
}
