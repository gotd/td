package manager

import (
	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mtproto"
	"github.com/nnqq/td/tg"
)

// Handler abstracts updates and session handler.
type Handler interface {
	OnSession(cfg tg.Config, s mtproto.Session) error
	OnMessage(b *bin.Buffer) error
}

// NoopHandler is a noop handler.
type NoopHandler struct{}

// OnSession implements Handler.
func (n NoopHandler) OnSession(cfg tg.Config, s mtproto.Session) error {
	return nil
}

// OnMessage implements Handler
func (n NoopHandler) OnMessage(b *bin.Buffer) error {
	return nil
}
