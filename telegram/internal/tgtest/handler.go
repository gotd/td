package tgtest

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

type Handler interface {
	OnNewClient(k crypto.AuthKey) error
	OnMessage(k crypto.AuthKey, msgID int64, in *bin.Buffer) error
}
