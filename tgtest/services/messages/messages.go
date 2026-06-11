// Package messages contains a personal-messages service implementation for the
// tgtest server.
//
// It implements a minimal subset of the messages.* API needed to exercise
// 1:1 (user-to-user) messaging end-to-end against the in-process tgtest server:
// sending messages (messages.sendMessage), reading them back
// (messages.getHistory) and delivering them to connected recipients as
// tg.UpdateNewMessage updates.
//
// Only user peers (tg.InputPeerSelf and tg.InputPeerUser) are supported; chats
// and channels are intentionally out of scope.
package messages

import (
	"sync"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/services"
)

// Service is a Telegram personal-messages service.
//
// It is safe for concurrent use.
type Service struct {
	clock clock.Clock

	mux sync.Mutex
	// users is the registry of known users, by ID.
	users map[int64]*tg.User
	// self maps an auth key ID to the ID of the user it represents.
	self map[[8]byte]int64
	// sessions maps a user ID to the set of sessions (by session ID) it is
	// connected with, used to deliver updates.
	sessions map[int64]map[int64]tgtest.Session
	// history maps a (selfUserID, peerUserID) pair to the dialog messages,
	// oldest first.
	history map[dialogKey][]*tg.Message

	// resolveSelf assigns an identity to a freshly seen session.
	resolveSelf func(tgtest.Session) *tg.User
	// nextUserID is used by the default self resolver to allocate IDs.
	nextUserID int64
	// lastMsgID is the last allocated message ID.
	lastMsgID int
	// lastPts is the last allocated PTS value.
	lastPts int
}

// dialogKey identifies a dialog from the point of view of self.
type dialogKey struct {
	self int64
	peer int64
}

// Option configures Service.
type Option func(*Service)

// WithClock sets clock to use for message dates.
func WithClock(c clock.Clock) Option {
	return func(s *Service) {
		s.clock = c
	}
}

// WithSelfResolver sets a function that assigns an identity (self user) to a
// session the first time it is seen.
//
// By default each distinct auth key is assigned a fresh synthetic user with a
// sequential ID.
func WithSelfResolver(f func(tgtest.Session) *tg.User) Option {
	return func(s *Service) {
		s.resolveSelf = f
	}
}

// NewService creates new messages Service.
func NewService(opts ...Option) *Service {
	s := &Service{
		clock:      clock.System,
		users:      map[int64]*tg.User{},
		self:       map[[8]byte]int64{},
		sessions:   map[int64]map[int64]tgtest.Session{},
		history:    map[dialogKey][]*tg.Message{},
		nextUserID: 1,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// AddUser registers user in the service registry and returns it.
//
// Registered users are resolved by ID when an InputPeerUser refers to them and
// are included in the Users field of responses and updates.
func (s *Service) AddUser(u *tg.User) *tg.User {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.users[u.ID] = u
	if u.ID >= s.nextUserID {
		s.nextUserID = u.ID + 1
	}
	return u
}

// Register registers service handlers in the dispatcher.
func (s *Service) Register(d *tgtest.Dispatcher) {
	d.HandleFunc(tg.MessagesSendMessageRequestTypeID, s.handle)
	d.HandleFunc(tg.MessagesGetHistoryRequestTypeID, s.handle)
}

func (s *Service) handle(server *tgtest.Server, req *tgtest.Request) error {
	id, err := req.Buf.PeekID()
	if err != nil {
		return err
	}

	switch id {
	case tg.MessagesSendMessageRequestTypeID:
		r := &tg.MessagesSendMessageRequest{}
		if err := r.Decode(req.Buf); err != nil {
			return err
		}
		return s.sendMessage(server, req, r)
	case tg.MessagesGetHistoryRequestTypeID:
		r := &tg.MessagesGetHistoryRequest{}
		if err := r.Decode(req.Buf); err != nil {
			return err
		}
		return s.getHistory(server, req, r)
	default:
		return services.ErrMethodNotImplemented
	}
}
