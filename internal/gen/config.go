package gen

type generateOptions struct {
	docBaseURL       string
	generateClient   bool
	generateRegistry bool
	generateServer   bool
	docLineLimit     int
}

func (s *generateOptions) setDefaults() {
	// Zero value docBaseURL handled by NewGenerator.
	// It's okay to use zero value generateClient.
	// It's okay to use zero value generateRegistry.
	// It's okay to use zero value generateServer.
	if s.docLineLimit == 0 {
		s.docLineLimit = 87
	}
}

// Option that configures generation.
type Option func(o *generateOptions)

// WithClient enables client generation.
func WithClient() Option {
	return func(o *generateOptions) {
		o.generateClient = true
	}
}

// WithRegistry enables type ID registry generation.
func WithRegistry() Option {
	return func(o *generateOptions) {
		o.generateRegistry = true
	}
}

// WithServer enables experimental server generation.
func WithServer() Option {
	return func(o *generateOptions) {
		o.generateServer = true
	}
}

// WithDocLineLimit sets GoDoc comment line length limit.
func WithDocLineLimit(limit int) Option {
	return func(o *generateOptions) {
		o.docLineLimit = limit
	}
}

// WithDocumentation will embed documentation references to generated code.
//
// If base is https://core.telegram.org, documentation content will be also
// embedded.
func WithDocumentation(base string) Option {
	return func(o *generateOptions) {
		o.docBaseURL = base
	}
}
