package tgtest

import (
	"context"

	"github.com/nnqq/td/bin"
)

// Request represents MTProto RPC request structure.
type Request struct {
	// DC ID from server structure.
	// Used to make handler less stateful.
	DC int
	// Session is a user session.
	Session Session
	// MsgID is a message ID of RPC request.
	MsgID int64
	// Buf contains RPC request
	Buf *bin.Buffer
	// RequestCtx is a request context.
	RequestCtx context.Context
}

// Handler is a RPC request handler.
type Handler interface {
	OnMessage(server *Server, req *Request) error
}

var _ Handler = HandlerFunc(nil)

// HandlerFunc is functional adapter for Handler.OnMessage method.
type HandlerFunc func(server *Server, req *Request) error

// OnMessage implements Handler.
func (h HandlerFunc) OnMessage(server *Server, req *Request) error {
	return h(server, req)
}
