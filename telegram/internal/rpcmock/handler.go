package rpcmock

import "github.com/gotd/td/bin"

// Handler is a RPC call handler.
type Handler interface {
	Handle(body bin.Encoder) (bin.Encoder, error)
}

// HandlerFunc is a function adapter for Handler.
type HandlerFunc func(body bin.Encoder) (bin.Encoder, error)

// Handle implements Handler.
func (h HandlerFunc) Handle(body bin.Encoder) (bin.Encoder, error) {
	return h(body)
}
