package messages

import (
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
)

// nextMessage allocates a new message ID and PTS. Caller must hold the mutex.
func (s *Service) nextMessage() (msgID, pts int) {
	s.lastMsgID++
	s.lastPts++
	return s.lastMsgID, s.lastPts
}

// store appends msg to the (self, peer) dialog. Caller must hold the mutex.
func (s *Service) store(self, peer int64, msg *tg.Message) {
	k := dialogKey{self: self, peer: peer}
	s.history[k] = append(s.history[k], msg)
}

func (s *Service) sendMessage(server *tgtest.Server, req *tgtest.Request, r *tg.MessagesSendMessageRequest) error {
	self := s.selfUser(req.Session)
	peer, err := s.resolvePeerUser(self, r.Peer)
	if err != nil {
		return err
	}

	date := int(s.clock.Now().Unix())

	s.mux.Lock()
	msgID, pts := s.nextMessage()
	// Sender's copy of the message: outgoing, addressed to the peer.
	out := &tg.Message{
		ID:      msgID,
		Out:     true,
		FromID:  &tg.PeerUser{UserID: self.ID},
		PeerID:  &tg.PeerUser{UserID: peer.ID},
		Date:    date,
		Message: r.Message,
	}
	out.SetFlags()
	s.store(self.ID, peer.ID, out)

	users := []tg.UserClass{self}
	var (
		incoming    *tg.Message
		incomingPts int
		recipients  []tgtest.Session
	)
	if peer.ID != self.ID {
		users = append(users, peer)
		// Recipient's copy: incoming, the dialog peer is the sender.
		var inMsgID int
		inMsgID, incomingPts = s.nextMessage()
		incoming = &tg.Message{
			ID:      inMsgID,
			FromID:  &tg.PeerUser{UserID: self.ID},
			PeerID:  &tg.PeerUser{UserID: self.ID},
			Date:    date,
			Message: r.Message,
		}
		incoming.SetFlags()
		s.store(peer.ID, self.ID, incoming)

		for _, session := range s.sessions[peer.ID] {
			recipients = append(recipients, session)
		}
	}
	s.mux.Unlock()

	// Best-effort delivery to connected recipients: a disconnected recipient
	// must not fail the sender's request.
	for _, session := range recipients {
		_ = server.SendUpdates(req.RequestCtx, session, &tg.UpdateNewMessage{
			Message:  incoming,
			Pts:      incomingPts,
			PtsCount: 1,
		})
	}

	return server.SendResult(req, &tg.Updates{
		Updates: []tg.UpdateClass{
			&tg.UpdateMessageID{ID: msgID, RandomID: r.RandomID},
			&tg.UpdateNewMessage{Message: out, Pts: pts, PtsCount: 1},
		},
		Users: users,
		Date:  date,
	})
}
