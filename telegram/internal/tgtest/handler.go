package tgtest

import (
	"github.com/gotd/td/bin"
)

type Handler interface {
	OnNewClient(s Session) error
	OnMessage(s Session, msgID int64, in *bin.Buffer) error
}
