// Package bot is a batteries-included bootstrap for Telegram bots on top of
// telegram.Client.
//
// One call connects, signs in with the bot token, and wires what every bot
// needs and is easy to get wrong by hand:
//   - the session in Options.Dir, so that the bot keeps its authorization
//     across restarts;
//   - the token is checked against the session: a session of another account
//     fails with ErrSessionMismatch instead of silently running as it;
//   - FLOOD_WAIT on sign in is waited out, up to Options.MaxLoginWait;
//   - the updates manager (telegram/updates) with its state in Options.Dir,
//     so that updates sent while the bot was stopped are delivered after the
//     restart, and gaps are recovered;
//   - the peer manager (telegram/peers), fed with the entities of every
//     update, including the ones of getDifference.
//
// Usage:
//
//	d := tg.NewUpdateDispatcher()
//	b, err := bot.New(appID, appHash, token, bot.Options{Dir: "state", Handler: d})
//	if err != nil {
//		return err
//	}
//	d.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
//		m, ok := u.Message.(*tg.Message)
//		if !ok || m.Out {
//			return nil
//		}
//		_, err := b.Sender().Reply(e, u).Text(ctx, m.Message)
//		return err
//	})
//	if err := b.Start(ctx); err != nil {
//		return err
//	}
//	defer func() { _ = b.Close() }()
//	return b.Wait(ctx)
package bot

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// File names in Options.Dir.
const (
	// SessionFile is the name of the session file in Options.Dir.
	SessionFile = "session.json"
	// UpdatesFile is the name of the update state file in Options.Dir.
	UpdatesFile = "updates.json"
)

const (
	defaultMaxLoginWait = 5 * time.Minute
	// defaultChannelDifferenceConcurrency bounds getChannelDifference calls of
	// a bot in many active groups.
	defaultChannelDifferenceConcurrency = 10
)

var (
	// ErrSessionMismatch is returned by Start if the stored session is
	// authorized as another account than the bot of the token: another
	// bot, or a user. Use another Options.Dir, or remove the session to sign
	// in again.
	ErrSessionMismatch = errors.New("session belongs to another account")
	// ErrNotBot is returned by Start if the token signed in an account that
	// is not a bot.
	ErrNotBot = errors.New("signed in account is not a bot")
)

// Options of a Bot.
type Options struct {
	// Client is the options of the telegram.Client.
	//
	// Client.UpdateHandler must be nil: the updates manager is the update
	// handler of the client, use Handler instead. If Client.SessionStorage
	// is nil and Dir is set, the session is kept in Dir.
	Client telegram.Options

	// Dir is the directory for the session (SessionFile) and the update
	// state (UpdatesFile). It is created if needed. Empty means memory
	// only: the bot signs in again on every start and gets no updates sent
	// while it was stopped.
	//
	// Use one Dir per bot, and never share it by running processes.
	Dir string

	// Handler is called with the updates after gap recovery, with the
	// entities of every update applied to the peer manager first (optional).
	//
	// Handlers run inside the client: they may use the Bot, but must not
	// call Close or Shutdown (use Stop).
	Handler telegram.UpdateHandler

	// Updates is the base configuration of the updates manager (optional).
	// Handler is set from Handler above. Storage and AccessHasher default to
	// an updates.FileStorage in Dir (memory without Dir).
	// MaxChannelDifferenceConcurrency defaults to 10. BeforeRunPass (the
	// zero value) is replaced by BeforeRunBuffer, so updates received during
	// sign in are not lost.
	Updates updates.Config

	// Peers is the options of the peer manager (optional). The storage and
	// the entity cache are in memory by default; unlike peers.Options, the
	// cache is on, so that peers seen in updates are resolved without RPCs.
	Peers peers.Options

	// MaxLoginWait is the maximum total FLOOD_WAIT to wait out on sign in
	// before Start fails. Defaults to 5 minutes; negative means do not wait.
	MaxLoginWait time.Duration
}

// Bot is a started Telegram bot: a telegram.Client with an updates manager
// and a peer manager.
type Bot struct {
	token   auth.BotToken
	client  *telegram.Client
	api     *tg.Client
	peers   *peers.Manager
	updates *updates.Manager
	sender  *message.Sender
	clock   clock.Clock

	maxLoginWait time.Duration

	state atomic.Int32 // botState
	self  atomic.Pointer[tg.User]

	done        chan struct{}
	doneOnce    sync.Once
	err         error // written before done is closed
	startFailed bool  // written before done is closed
}

// botState is the lifecycle state of a Bot.
type botState int32

const (
	stateIdle botState = iota
	stateStarted
	stateClosed // stopped before Start
)

// New creates a Bot from the bot token without connecting, so that handlers
// can be set up with the Bot before Start.
//
// It returns an error wrapping auth.ErrInvalidBotToken for a malformed token,
// without the token in the message.
func New(appID int, appHash, token string, opt Options) (*Bot, error) {
	tok, err := auth.ParseBotToken(token)
	if err != nil {
		return nil, err
	}
	if opt.Client.UpdateHandler != nil {
		return nil, errors.New("Options.Client.UpdateHandler must be nil, use Options.Handler")
	}

	updCfg := opt.Updates
	if opt.Dir != "" {
		if err := os.MkdirAll(opt.Dir, 0o700); err != nil {
			return nil, errors.Wrap(err, "create dir")
		}
		if opt.Client.SessionStorage == nil {
			opt.Client.SessionStorage = &session.FileStorage{Path: filepath.Join(opt.Dir, SessionFile)}
		}
		if updCfg.Storage == nil || updCfg.AccessHasher == nil {
			fs := updates.NewFileStorage(filepath.Join(opt.Dir, UpdatesFile))
			if updCfg.Storage == nil {
				updCfg.Storage = fs
			}
			if updCfg.AccessHasher == nil {
				updCfg.AccessHasher = fs
			}
		}
	}
	if updCfg.MaxChannelDifferenceConcurrency == 0 {
		updCfg.MaxChannelDifferenceConcurrency = defaultChannelDifferenceConcurrency
	}
	if updCfg.BeforeRun == updates.BeforeRunPass {
		updCfg.BeforeRun = updates.BeforeRunBuffer
	}
	if updCfg.Logger == nil {
		updCfg.Logger = opt.Client.Logger
	}
	if opt.Peers.Cache == nil {
		// Unlike peers.Options, cache entities by default (PEER-1), so seen
		// peers are resolved without RPCs.
		opt.Peers.Cache = &peers.InmemoryCache{}
	}
	if opt.Peers.Logger == nil {
		opt.Peers.Logger = opt.Client.Logger
	}
	maxLoginWait := opt.MaxLoginWait
	if maxLoginWait == 0 {
		maxLoginWait = defaultMaxLoginWait
	}
	clk := opt.Client.Clock
	if clk == nil {
		clk = clock.System
	}

	b := &Bot{
		token:        tok,
		clock:        clk,
		maxLoginWait: maxLoginWait,
		done:         make(chan struct{}),
	}
	// The client needs the updates manager as its handler, and the managers
	// need the API of the client.
	opt.Client.UpdateHandler = telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		return b.updates.Handle(ctx, u)
	})
	b.client = telegram.NewClient(appID, appHash, opt.Client)
	b.api = b.client.API()
	b.peers = opt.Peers.Build(b.api)
	b.sender = message.NewSender(b.api)

	handler := opt.Handler
	if handler == nil {
		handler = telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })
	}
	updCfg.Handler = b.peers.UpdateHook(handler)
	b.updates = updates.New(updCfg)
	return b, nil
}

// Start creates a Bot with New and starts it with Bot.Start.
func Start(
	ctx context.Context, appID int, appHash, token string, opt Options, so ...telegram.StartOption,
) (*Bot, error) {
	b, err := New(appID, appHash, token, opt)
	if err != nil {
		return nil, err
	}
	if err := b.Start(ctx, so...); err != nil {
		return nil, err
	}
	return b, nil
}

// Start connects, signs in and starts the updates manager. It returns once the
// bot is signed in and the update state is loaded, so missed updates are
// being fetched; they are passed to Options.Handler in the background.
//
// The ctx bounds the startup only, see telegram.Client.Start, which gets the
// StartOptions. If the stored session is authorized as another account, Start
// fails with ErrSessionMismatch. An invalid or revoked token fails with the
// RPC error, e.g. ACCESS_TOKEN_INVALID (see tgerr.Is). On error the bot is
// fully stopped.
//
// A Bot can be started once.
func (b *Bot) Start(ctx context.Context, so ...telegram.StartOption) error {
	if !b.state.CompareAndSwap(int32(stateIdle), int32(stateStarted)) {
		if botState(b.state.Load()) == stateClosed {
			return telegram.ErrClientClosed
		}
		return telegram.ErrAlreadyStarted
	}
	if err := b.start(ctx, so...); err != nil {
		b.finishStart(err, true)
		return err
	}
	return nil
}

func (b *Bot) start(ctx context.Context, so ...telegram.StartOption) error {
	if err := b.client.Start(ctx, so...); err != nil {
		return errors.Wrap(err, "start client")
	}
	fail := func(err error) error {
		_ = b.client.Close()
		return err
	}

	self, err := b.login(ctx)
	if err != nil {
		return fail(err)
	}
	b.self.Store(self)

	// The context of the updates manager keeps the values of ctx, but its
	// lifetime is the one of the client.
	updCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	ready := make(chan struct{})
	updDone := make(chan error, 1)
	go func() {
		updDone <- b.updates.Run(updCtx, b.api, self.ID, updates.AuthOptions{
			IsBot:   true,
			OnStart: func(context.Context) { close(ready) },
		})
	}()
	select {
	case <-ready:
	case err := <-updDone:
		cancel()
		if err == nil {
			err = errors.New("stopped")
		}
		return fail(errors.Wrap(err, "updates"))
	case <-b.client.Done():
		cancel()
		<-updDone
		if err := b.client.Wait(context.Background()); err != nil {
			return errors.Wrap(err, "client stopped")
		}
		return errors.New("client stopped")
	case <-ctx.Done():
		cancel()
		<-updDone
		return fail(errors.Wrap(ctx.Err(), "start updates"))
	}

	go func() {
		var err error
		select {
		case <-b.client.Done():
			cancel()
			<-updDone
		case updErr := <-updDone:
			// The updates manager stopped on its own, e.g. on a storage
			// error with Updates.StopOnApplyError: stop the client too.
			cancel()
			if updErr == nil {
				updErr = errors.New("stopped")
			}
			err = errors.Wrap(updErr, "updates")
			b.client.Stop()
			<-b.client.Done()
		}
		if cerr := b.client.Wait(context.Background()); cerr != nil {
			// The client error caused the stop.
			err = cerr
		}
		b.finishStart(err, false)
	}()
	return nil
}

// login checks the session and signs in with the token if needed.
func (b *Bot) login(ctx context.Context) (*tg.User, error) {
	self, err := b.peers.Self(ctx)
	switch {
	case err == nil:
		u := self.Raw()
		if !u.Bot || u.ID != b.token.ID {
			return nil, errors.Wrapf(ErrSessionMismatch,
				"session is authorized as %d (bot: %t), token is of bot %d", u.ID, u.Bot, b.token.ID)
		}
		return u, nil
	case !auth.IsUnauthorized(err):
		return nil, errors.Wrap(err, "check authorization")
	}

	var waited time.Duration
	for {
		_, err := b.client.Auth().Bot(ctx, b.token.String())
		if err == nil {
			break
		}
		d, ok := tgerr.AsFloodWait(err)
		if !ok || b.maxLoginWait < 0 || waited+d > b.maxLoginWait {
			return nil, errors.Wrap(err, "sign in")
		}
		waited += d
		t := b.clock.Timer(d)
		select {
		case <-t.C():
		case <-ctx.Done():
			t.Stop()
			return nil, errors.Wrap(ctx.Err(), "wait to sign in")
		}
	}

	// Fetch self through the peer manager, so that it is cached there.
	self, err = b.peers.Self(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get self")
	}
	u := self.Raw()
	if !u.Bot {
		return nil, ErrNotBot
	}
	if u.ID != b.token.ID {
		return nil, errors.Wrapf(ErrSessionMismatch, "signed in as %d, token is of bot %d", u.ID, b.token.ID)
	}
	return u, nil
}

func (b *Bot) finishStart(err error, startFailed bool) {
	b.doneOnce.Do(func() {
		b.err = err
		b.startFailed = startFailed
		close(b.done)
	})
}

// Client returns the underlying client.
func (b *Bot) Client() *telegram.Client { return b.client }

// API returns the raw Telegram API client.
func (b *Bot) API() *tg.Client { return b.api }

// Peers returns the peer manager.
func (b *Bot) Peers() *peers.Manager { return b.peers }

// Updates returns the updates manager.
func (b *Bot) Updates() *updates.Manager { return b.updates }

// Sender returns a message sender.
func (b *Bot) Sender() *message.Sender { return b.sender }

// ID returns the user ID of the bot, from the token.
func (b *Bot) ID() int64 { return b.token.ID }

// Self returns the user of the bot, nil before Start succeeded.
func (b *Bot) Self() *tg.User { return b.self.Load() }

// Stop stops the bot without waiting, see telegram.Client.Stop. It is safe to
// call from handlers.
func (b *Bot) Stop() {
	if b.state.CompareAndSwap(int32(stateIdle), int32(stateClosed)) {
		b.finishStart(nil, false)
	}
	b.client.Stop()
}

// Shutdown stops the bot and waits until the client and the updates manager
// are stopped or ctx is done. It returns the error that stopped the bot before
// Shutdown was called, nil if Shutdown stopped it or if Start failed, or the
// wrapped ctx.Err().
//
// Never call it from handlers, use Stop there.
func (b *Bot) Shutdown(ctx context.Context) error {
	b.Stop()
	select {
	case <-b.done:
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "shutdown")
	}
	if b.startFailed {
		return nil
	}
	return b.err
}

// Close is Shutdown without a deadline. Never call it from handlers, use Stop
// there.
func (b *Bot) Close() error { return b.Shutdown(context.Background()) }

// Done returns a channel that is closed when the bot stops.
func (b *Bot) Done() <-chan struct{} { return b.done }

// Wait blocks until the bot stops or ctx is done. It returns the error that
// stopped the bot (e.g. *telegram.AuthLostError, or an error of the updates
// manager), nil if it was stopped by Stop or Close, or the wrapped ctx.Err().
func (b *Bot) Wait(ctx context.Context) error {
	select {
	case <-b.done:
		return b.err
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "wait")
	}
}
