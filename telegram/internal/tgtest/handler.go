package tgtest

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

type Handler interface {
	OnNewClient(k crypto.AuthKeyWithID) error
	OnMessage(s Session, msgID int64, in *bin.Buffer) error
}
