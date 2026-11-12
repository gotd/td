package mtproto

import "sync"

// SessionState is the server session of a connection slot: session_id and
// the number of content messages sent in it (for seqno). Connections that
// share a SessionState one at a time continue the same server session after
// a reconnect with the same auth key, so the server keeps its queued answers
// and updates and does not send new_session_created, like TDLib and Telegram
// Desktop do. A new session is started when the key changes.
//
// A SessionState must be used by one connection at a time: acquire it with
// TryAcquire before passing it in Options; the connection releases it when
// Run returns.
type SessionState struct {
	mux             sync.Mutex
	keyID           [8]byte
	id              int64
	contentMessages int32
	inUse           bool
}

// NewSessionState returns an empty SessionState.
func NewSessionState() *SessionState {
	return &SessionState{}
}

// TryAcquire marks the state as used by a connection. It reports false if
// another connection uses it.
func (s *SessionState) TryAcquire() bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.inUse {
		return false
	}
	s.inUse = true
	return true
}

// load returns the session of key keyID, if any.
func (s *SessionState) load(keyID [8]byte) (id int64, contentMessages int32, ok bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.id == 0 || s.keyID != keyID {
		return 0, 0, false
	}
	return s.id, s.contentMessages, true
}

// release saves the session of the connection and frees the state. A zero id
// forgets the session.
func (s *SessionState) release(keyID [8]byte, id int64, contentMessages int32) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.keyID = keyID
	s.id = id
	s.contentMessages = contentMessages
	s.inUse = false
}
