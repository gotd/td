package e2etest

import "github.com/gotd/td/tg"

func filterMessage(update *tg.UpdateNewMessage) bool {
	if v, ok := update.Message.(interface{ GetOut() bool }); ok && v.GetOut() {
		return true
	}

	if v, ok := update.Message.(interface{ GetPeerID() tg.PeerClass }); ok && v.GetPeerID() == nil {
		return true
	}

	return false
}
