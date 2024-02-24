package gen

import "flag"

// GenerateFlags is flags for optional generation.
type GenerateFlags struct {
	// Client enables client generation.
	Client bool
	// Registry enables type ID registry generation.
	Registry bool
	// Server enables experimental server generation.
	Server bool
	// Handlers enables update handler generation.
	Handlers bool
	// UpdatesClassifier enables updates classifier generation.
	UpdatesClassifier bool
	// GetSet enables getters and setters generation.
	GetSet bool
	// Mapping enables mapping helpers generation.
	Mapping bool
	// Slices enables slice helpers generation.
	Slices bool
	// TDLibJSON enables TDLib API JSON encoders and decoders generation.
	TDLibJSON bool
}

// RegisterFlags registers GenerateFlags fields in given flag set.
func (s *GenerateFlags) RegisterFlags(set *flag.FlagSet) {
	set.BoolVar(&s.Client, "client", true, "Enables client generation")
	set.BoolVar(&s.Registry, "registry", true, "Enables type ID registry generation")
	set.BoolVar(&s.Server, "server", false, "Enables experimental server generation")
	set.BoolVar(&s.Handlers, "handlers", false, "Enables update handler generation")
	set.BoolVar(&s.UpdatesClassifier, "updates-classifier", true, "Enables updates classifier generation")
	set.BoolVar(&s.GetSet, "getset", true, "Enables getters and setters generation")
	set.BoolVar(&s.Mapping, "mapping", false, "Enables mapping helpers generation")
	set.BoolVar(&s.Slices, "slices", false, "Enables slice helpers generation")
	set.BoolVar(&s.TDLibJSON, "tdlib-json", false, "Enables TDLib JSON encoding generation")
}

// GeneratorOptions is a Generator options structure.
type GeneratorOptions struct {
	// DocBaseURL is a documentation base URL.
	// If DocBaseURL is set, Generator will embed documentation references to generated code.
	//
	// If base is https://core.telegram.org, documentation content will be also
	// embedded.
	DocBaseURL string
	// DocLineLimit sets GoDoc comment line length limit.
	DocLineLimit int
	GenerateFlags
}

// RegisterFlags registers GeneratorOptions fields in given flag set.
func (s *GeneratorOptions) RegisterFlags(set *flag.FlagSet) {
	set.StringVar(&s.DocBaseURL, "doc", "", "Base documentation url")
	set.IntVar(&s.DocLineLimit, "line-limit", 0, "GoDoc comment line length limit")
	s.GenerateFlags.RegisterFlags(set)
}

func (s *GeneratorOptions) setDefaults() {
	// Zero value DocBaseURL handled by NewGenerator.
	// It's okay to use zero value GenerateClient.
	// It's okay to use zero value GenerateRegistry.
	// It's okay to use zero value GenerateServer.
	// It's okay to use zero value GenerateHelpers.
	// It's okay to use zero value GenerateSlices.
	if s.DocLineLimit == 0 {
		s.DocLineLimit = 87
	}
}
