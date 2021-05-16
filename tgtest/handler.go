package tgtest

import (
	"context"

	"github.com/gotd/td/bin"
)

type Request struct {
	DC         int
	Session    Session
	MsgID      int64
	Buf        *bin.Buffer
	RequestCtx context.Context
}

type Handler interface {
	OnMessage(server *Server, req *Request) error
}

// HandlerFunc is functional adapter for Handler.OnMessage method.
type HandlerFunc func(server *Server, req *Request) error

func (h HandlerFunc) OnMessage(server *Server, req *Request) error {
	return h(server, req)
}
