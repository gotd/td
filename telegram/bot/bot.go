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
	token auth.BotToken
	rt    *bootstrap.Runtime
}

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
	b := &Bot{token: tok}
	rt, err := bootstrap.New(appID, appHash, bootstrap.Config{
		Client:       opt.Client,
		Dir:          opt.Dir,
		Handler:      opt.Handler,
		Updates:      opt.Updates,
		Peers:        opt.Peers,
		MaxLoginWait: opt.MaxLoginWait,
		IsBot:        true,
		Login:        b.login,
	})
	if err != nil {
		return nil, err
	}
	b.rt = rt
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
	return b.rt.Start(ctx, so...)
}

// login checks the session and signs in with the token if needed.
func (b *Bot) login(ctx context.Context, rt *bootstrap.Runtime) (*tg.User, error) {
	self, err := rt.Peers.Self(ctx)
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

	if err := rt.RetryFlood(ctx, func(ctx context.Context) error {
		_, err := rt.Client.Auth().Bot(ctx, b.token.String())
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "sign in")
	}

	// Fetch self through the peer manager, so that it is cached there.
	self, err = rt.Peers.Self(ctx)
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

// Client returns the underlying client.
func (b *Bot) Client() *telegram.Client { return b.rt.Client }

// API returns the raw Telegram API client.
func (b *Bot) API() *tg.Client { return b.rt.API }

// Peers returns the peer manager.
func (b *Bot) Peers() *peers.Manager { return b.rt.Peers }

// Updates returns the updates manager.
func (b *Bot) Updates() *updates.Manager { return b.rt.Updates }

// Sender returns a message sender.
func (b *Bot) Sender() *message.Sender { return b.rt.Sender }

// ID returns the user ID of the bot, from the token.
func (b *Bot) ID() int64 { return b.token.ID }

// Self returns the user of the bot, nil before Start succeeded.
func (b *Bot) Self() *tg.User { return b.rt.Self() }

// Stop stops the bot without waiting, see telegram.Client.Stop. It is safe to
// call from handlers.
func (b *Bot) Stop() { b.rt.Stop() }

// Shutdown stops the bot and waits until the client and the updates manager
// are stopped or ctx is done. It returns the error that stopped the bot before
// Shutdown was called, nil if Shutdown stopped it or if Start failed, or the
// wrapped ctx.Err().
//
// Never call it from handlers, use Stop there.
func (b *Bot) Shutdown(ctx context.Context) error { return b.rt.Shutdown(ctx) }

// Close is Shutdown without a deadline. Never call it from handlers, use Stop
// there.
func (b *Bot) Close() error { return b.rt.Close() }

// Done returns a channel that is closed when the bot stops.
func (b *Bot) Done() <-chan struct{} { return b.rt.Done() }

// Wait blocks until the bot stops or ctx is done. It returns the error that
// stopped the bot (e.g. *telegram.AuthLostError, or an error of the updates
// manager), nil if it was stopped by Stop or Close, or the wrapped ctx.Err().
func (b *Bot) Wait(ctx context.Context) error { return b.rt.Wait(ctx) }
