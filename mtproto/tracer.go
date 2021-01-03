package mtproto

import "github.com/gotd/td/bin"

//go:generate go run tracer_generator.go
//procm:use=tracer
type tracer struct {
	// Message is called on every incoming message if set.
	Message func(b *bin.Buffer)
}
