package tgtest

import (
	"github.com/gotd/td/bin"
)

type Request struct {
	Session Session
	MsgID   int64
	Buf     *bin.Buffer
}

type Handler interface {
	OnMessage(server *Server, req *Request) error
}

// HandlerFunc is functional adapter for Handler.OnMessage method.
type HandlerFunc func(server *Server, req *Request) error

func (h HandlerFunc) OnMessage(server *Server, req *Request) error {
	return h(server, req)
}
