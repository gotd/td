package e2etest

import (
	"strings"

	"github.com/gotd/td/tg"
)

func filterMessage(update *tg.UpdateNewMessage) bool {
	if v, ok := update.Message.(interface{ GetOut() bool }); ok && v.GetOut() {
		return true
	}

	if v, ok := update.Message.(interface{ GetPeerID() tg.PeerClass }); ok && v.GetPeerID() == nil {
		return true
	}
	if _, ok := update.Message.(*tg.MessageService); ok {
		return true
	}
	if v, ok := update.Message.(interface{ GetMessage() string }); ok && strings.HasPrefix(v.GetMessage(), "Login code:") {
		return true
	}

	return false
}
