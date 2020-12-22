package tgtest

import (
	"github.com/gotd/td/bin"
)

type Handler interface {
	OnMessage(s Session, msgID int64, in *bin.Buffer) error
}

// HandlerFunc is functional adapter for Handler.OnMessage method.
type HandlerFunc func(s Session, msgID int64, in *bin.Buffer) error

func (h HandlerFunc) OnMessage(s Session, msgID int64, in *bin.Buffer) error {
	return h(s, msgID, in)
}
