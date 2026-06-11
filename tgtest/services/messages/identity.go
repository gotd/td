package messages

import (
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgtest"
)

// selfUser returns the user the given session represents, binding the session
// to an identity the first time it is seen.
//
// It also records the session for update delivery. Caller must not hold the
// mutex.
func (s *Service) selfUser(session tgtest.Session) *tg.User {
	s.mux.Lock()
	defer s.mux.Unlock()

	id, ok := s.self[session.AuthKey.ID]
	if !ok {
		var u *tg.User
		if s.resolveSelf != nil {
			u = s.resolveSelf(session)
		}
		if u == nil {
			u = &tg.User{ID: s.nextUserID, Self: true}
		}
		if _, ok := s.users[u.ID]; !ok {
			s.users[u.ID] = u
		}
		if u.ID >= s.nextUserID {
			s.nextUserID = u.ID + 1
		}
		id = u.ID
		s.self[session.AuthKey.ID] = id
	}

	s.bindSession(id, session)
	return s.users[id]
}

// bindSession records session for the given user. Caller must hold the mutex.
func (s *Service) bindSession(userID int64, session tgtest.Session) {
	sessions, ok := s.sessions[userID]
	if !ok {
		sessions = map[int64]tgtest.Session{}
		s.sessions[userID] = sessions
	}
	sessions[session.ID] = session
}

// user returns the registered user by ID, creating a stub if unknown. Caller
// must hold the mutex.
func (s *Service) user(id int64) *tg.User {
	if u, ok := s.users[id]; ok {
		return u
	}
	u := &tg.User{ID: id}
	s.users[id] = u
	return u
}

// resolvePeerUser resolves an input peer to a user, from the point of view of
// self. Only user peers are supported.
func (s *Service) resolvePeerUser(self *tg.User, peer tg.InputPeerClass) (*tg.User, error) {
	switch p := peer.(type) {
	case *tg.InputPeerSelf:
		return self, nil
	case *tg.InputPeerUser:
		s.mux.Lock()
		defer s.mux.Unlock()
		return s.user(p.UserID), nil
	default:
		return nil, tgerr.New(400, "PEER_ID_INVALID")
	}
}
