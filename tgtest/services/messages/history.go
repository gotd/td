package messages

import (
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
)

// defaultHistoryLimit is used when a getHistory request does not set a limit.
const defaultHistoryLimit = 100

func (s *Service) getHistory(server *tgtest.Server, req *tgtest.Request, r *tg.MessagesGetHistoryRequest) error {
	self := s.selfUser(req.Session)
	peer, err := s.resolvePeerUser(self, r.Peer)
	if err != nil {
		return err
	}

	limit := r.Limit
	if limit <= 0 {
		limit = defaultHistoryLimit
	}

	s.mux.Lock()
	stored := s.history[dialogKey{self: self.ID, peer: peer.ID}]
	// Telegram returns messages newest first.
	messages := make([]tg.MessageClass, 0, len(stored))
	for i := len(stored) - 1; i >= 0 && len(messages) < limit; i-- {
		messages = append(messages, stored[i])
	}
	users := []tg.UserClass{self}
	if peer.ID != self.ID {
		users = append(users, peer)
	}
	s.mux.Unlock()

	return server.SendResult(req, &tg.MessagesMessages{
		Messages: messages,
		Users:    users,
	})
}
