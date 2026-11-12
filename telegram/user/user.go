// Package user is a batteries-included bootstrap for user accounts (userbots,
// archivers, services) on top of telegram.Client.
//
// One call connects, signs in if needed, and wires what a long-running user
// client needs and is easy to get wrong by hand:
//   - the session in Options.Dir, so that the account keeps its authorization
//     across restarts: an authorized session is never signed in again;
//   - the session is checked: a bot session, or another account than
//     Options.UserID, fails with ErrSessionMismatch;
//   - an unauthorized session is signed in with Options.Auth (phone code or
//     QR), or Start fails with ErrNotAuthorized; FLOOD_WAIT on sign in is
//     waited out, up to Options.MaxLoginWait;
//   - the updates manager (telegram/updates) with its state and the access
//     hashes of channels and users in Options.Dir, so that updates sent while
//     the client was stopped are delivered after the restart, gaps are
//     recovered, and known channels need no extra getDifference;
//   - the peer manager (telegram/peers) with an entity cache, fed with the
//     entities of every update, including the ones of getDifference.
//
// Usage:
//
//	d := tg.NewUpdateDispatcher()
//	u, err := user.New(appID, appHash, user.Options{
//		Dir:     "state",
//		Handler: d,
//		Auth:    user.Phone(auth.Constant(phone, password, codeAuth), auth.SendCodeOptions{}),
//	})
//	if err != nil {
//		return err
//	}
//	d.OnNewMessage(func(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage) error {
//		// ...
//		return nil
//	})
//	if err := u.Start(ctx); err != nil {
//		return err
//	}
//	defer func() { _ = u.Close() }()
//	return u.Wait(ctx)
package user

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/internal/bootstrap"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

// File names in Options.Dir.
const (
	// SessionFile is the name of the session file in Options.Dir.
	SessionFile = bootstrap.SessionFile
	// UpdatesFile is the name of the update state file in Options.Dir.
	UpdatesFile = bootstrap.UpdatesFile
)

var (
	// ErrSessionMismatch is returned by Start if the stored or signed in
	// account is a bot, or another user than Options.UserID. Use another
	// Options.Dir, or remove the session to sign in again.
	ErrSessionMismatch = errors.New("session belongs to another account")
	// ErrNotAuthorized is returned by Start if the session is not authorized
	// and Options.Auth is nil. Nothing is sent to sign in.
	ErrNotAuthorized = errors.New("session is not authorized and no authenticator is set")
)

// Options of a User.
type Options struct {
	// Client is the options of the telegram.Client.
	//
	// Client.UpdateHandler must be nil: the updates manager is the update
	// handler of the client, use Handler instead. If Client.SessionStorage
	// is nil and Dir is set, the session is kept in Dir.
	Client telegram.Options

	// Dir is the directory for the session (SessionFile) and the update
	// state with access hashes (UpdatesFile). It is created if needed. Empty
	// means memory only: every start needs a sign in, and updates sent while
	// stopped are not delivered.
	//
	// Use one Dir per account, and never share it by running processes.
	Dir string

	// Handler is called with the updates after gap recovery, with the
	// entities of every update applied to the peer manager first (optional).
	//
	// Handlers run inside the client: they may use the User, but must not
	// call Close or Shutdown (use Stop).
	Handler telegram.UpdateHandler

	// Updates is the base configuration of the updates manager (optional).
	// Handler is set from Handler above. Storage and AccessHasher default to
	// an updates.FileStorage in Dir (memory without Dir); the file storage
	// also keeps user access hashes. MaxChannelDifferenceConcurrency
	// defaults to 10. BeforeRunPass (the zero value) is replaced by
	// BeforeRunBuffer, so updates received during sign in are not lost.
	Updates updates.Config

	// Peers is the options of the peer manager (optional). The entity cache
	// is in memory by default.
	Peers peers.Options

	// Auth signs in an unauthorized session (optional): Phone or QR. If nil,
	// Start fails with ErrNotAuthorized instead of signing in, e.g. for a
	// service that must only run with an existing session.
	Auth Authenticator

	// UserID is the expected account (optional). If set, a session of
	// another user fails with ErrSessionMismatch.
	UserID int64

	// MaxLoginWait is the maximum total FLOOD_WAIT to wait out on sign in
	// before Start fails. Defaults to 5 minutes; negative means do not wait.
	MaxLoginWait time.Duration
}

// User is a started user account: a telegram.Client with an updates manager
// and a peer manager.
type User struct {
	opt     Options
	appID   int
	appHash string
	rt      *bootstrap.Runtime
	qr      loginTokens
}

// New creates a User without connecting, so that handlers can be set up with
// the User before Start.
func New(appID int, appHash string, opt Options) (*User, error) {
	u := &User{opt: opt, appID: appID, appHash: appHash}
	rt, err := bootstrap.New(appID, appHash, bootstrap.Config{
		Client:       opt.Client,
		Dir:          opt.Dir,
		Handler:      opt.Handler,
		Updates:      opt.Updates,
		Peers:        opt.Peers,
		MaxLoginWait: opt.MaxLoginWait,
		OnUpdates:    u.qr.onUpdates,
		Login:        u.login,
	})
	if err != nil {
		return nil, err
	}
	u.rt = rt
	return u, nil
}

// Start creates a User with New and starts it with User.Start.
func Start(ctx context.Context, appID int, appHash string, opt Options, so ...telegram.StartOption) (*User, error) {
	u, err := New(appID, appHash, opt)
	if err != nil {
		return nil, err
	}
	if err := u.Start(ctx, so...); err != nil {
		return nil, err
	}
	return u, nil
}

// Start connects, signs in if needed and starts the updates manager. It
// returns once the account is signed in and the update state is loaded, so
// missed updates are being fetched; they are passed to Options.Handler in the
// background.
//
// The ctx bounds the startup only (including an interactive sign in), see
// telegram.Client.Start, which gets the StartOptions. An authorized session is
// never signed in again. If it is a bot or another user than Options.UserID,
// Start fails with ErrSessionMismatch; if it is not authorized and
// Options.Auth is nil, with ErrNotAuthorized. On error the client is fully
// stopped.
//
// A User can be started once.
func (u *User) Start(ctx context.Context, so ...telegram.StartOption) error {
	return u.rt.Start(ctx, so...)
}

// login checks the session and signs in with Options.Auth if needed.
func (u *User) login(ctx context.Context, rt *bootstrap.Runtime) (*tg.User, error) {
	self, err := rt.Peers.Self(ctx)
	switch {
	case err == nil:
		if err := u.check(self.Raw()); err != nil {
			return nil, err
		}
		return self.Raw(), nil
	case !auth.IsUnauthorized(err):
		return nil, errors.Wrap(err, "check authorization")
	}
	if u.opt.Auth == nil {
		return nil, ErrNotAuthorized
	}

	if err := rt.RetryFlood(ctx, func(ctx context.Context) error {
		return u.opt.Auth.signIn(ctx, signInContext{
			client: rt.Client, appID: u.appID, appHash: u.appHash, tokens: &u.qr,
		})
	}); err != nil {
		return nil, errors.Wrap(err, "sign in")
	}

	// Fetch self through the peer manager, so that it is cached there.
	self, err = rt.Peers.Self(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get self")
	}
	if err := u.check(self.Raw()); err != nil {
		return nil, err
	}
	return self.Raw(), nil
}

// check verifies that the authorized account is the expected user.
func (u *User) check(self *tg.User) error {
	if self.Bot {
		return errors.Wrapf(ErrSessionMismatch, "session is authorized as bot %d", self.ID)
	}
	if u.opt.UserID != 0 && self.ID != u.opt.UserID {
		return errors.Wrapf(ErrSessionMismatch, "session is authorized as %d, expected %d", self.ID, u.opt.UserID)
	}
	return nil
}

// Client returns the underlying client.
func (u *User) Client() *telegram.Client { return u.rt.Client }

// API returns the raw Telegram API client.
func (u *User) API() *tg.Client { return u.rt.API }

// Peers returns the peer manager.
func (u *User) Peers() *peers.Manager { return u.rt.Peers }

// Updates returns the updates manager.
func (u *User) Updates() *updates.Manager { return u.rt.Updates }

// Sender returns a message sender.
func (u *User) Sender() *message.Sender { return u.rt.Sender }

// Self returns the authorized user, nil before Start succeeded.
func (u *User) Self() *tg.User { return u.rt.Self() }

// ID returns the ID of the authorized user, zero before Start succeeded.
func (u *User) ID() int64 {
	if self := u.rt.Self(); self != nil {
		return self.ID
	}
	return 0
}

// Stop stops the client without waiting, see telegram.Client.Stop. It is safe
// to call from handlers.
func (u *User) Stop() { u.rt.Stop() }

// Shutdown stops the client and waits until it and the updates manager are
// stopped or ctx is done. It returns the error that stopped the client before
// Shutdown was called, nil if Shutdown stopped it or if Start failed, or the
// wrapped ctx.Err().
//
// Never call it from handlers, use Stop there.
func (u *User) Shutdown(ctx context.Context) error { return u.rt.Shutdown(ctx) }

// Close is Shutdown without a deadline. Never call it from handlers, use Stop
// there.
func (u *User) Close() error { return u.rt.Close() }

// Done returns a channel that is closed when the client stops.
func (u *User) Done() <-chan struct{} { return u.rt.Done() }

// Wait blocks until the client stops or ctx is done. It returns the error that
// stopped it (e.g. *telegram.AuthLostError, or an error of the updates
// manager), nil if it was stopped by Stop or Close, or the wrapped ctx.Err().
func (u *User) Wait(ctx context.Context) error { return u.rt.Wait(ctx) }
