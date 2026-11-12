package manager

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

// Handler abstracts updates and session handler.
type Handler interface {
	OnSession(cfg tg.Config, s mtproto.Session) error
	OnMessage(b *bin.Buffer) error
}

// NewSessionHandler is implemented by a Handler that wants to know when the
// server created a new session (new_session_created).
type NewSessionHandler interface {
	OnNewSession(firstMsgID int64) error
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
