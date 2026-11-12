// Package bootstrap is the machinery shared by telegram/bot and telegram/user:
// a telegram.Client with an updates manager and a peer manager, their wiring,
// startup and lifecycle.
package bootstrap

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
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// File names in Config.Dir.
const (
	SessionFile = "session.json"
	UpdatesFile = "updates.json"
)

const (
	defaultMaxLoginWait = 5 * time.Minute
	// defaultChannelDifferenceConcurrency bounds getChannelDifference calls
	// of an account in many active channels.
	defaultChannelDifferenceConcurrency = 10
)

// Config of a Runtime.
type Config struct {
	Client  telegram.Options
	Dir     string
	Handler telegram.UpdateHandler
	Updates updates.Config
	Peers   peers.Options
	// MaxLoginWait is the maximum total FLOOD_WAIT waited out by
	// RetryFlood. Zero means 5 minutes, negative means do not wait.
	MaxLoginWait time.Duration
	// IsBot is passed to the updates manager.
	IsBot bool
	// OnUpdates is called with every update received by the client before
	// the updates manager gets it (optional), e.g. for updateLoginToken
	// during sign in.
	OnUpdates func(ctx context.Context, u tg.UpdatesClass)
	// Login checks the session and signs in if needed, returning the
	// authorized user.
	Login func(ctx context.Context, r *Runtime) (*tg.User, error)
}

// Runtime is a telegram.Client with an updates manager and a peer manager.
type Runtime struct {
	Client  *telegram.Client
	API     *tg.Client
	Peers   *peers.Manager
	Updates *updates.Manager
	Sender  *message.Sender
	Clock   clock.Clock

	maxLoginWait time.Duration
	isBot        bool
	login        func(ctx context.Context, r *Runtime) (*tg.User, error)

	state atomic.Int32 // lifecycle
	self  atomic.Pointer[tg.User]

	done        chan struct{}
	doneOnce    sync.Once
	err         error // written before done is closed
	startFailed bool  // written before done is closed
}

// lifecycle is the state of a Runtime.
type lifecycle int32

const (
	stateIdle lifecycle = iota
	stateStarted
	stateClosed // stopped before Start
)

// New creates a Runtime without connecting.
func New(appID int, appHash string, cfg Config) (*Runtime, error) {
	if cfg.Client.UpdateHandler != nil {
		return nil, errors.New("Options.Client.UpdateHandler must be nil, use Options.Handler")
	}
	if cfg.Login == nil {
		return nil, errors.New("no login")
	}

	updCfg := cfg.Updates
	if cfg.Dir != "" {
		if err := os.MkdirAll(cfg.Dir, 0o700); err != nil {
			return nil, errors.Wrap(err, "create dir")
		}
		if cfg.Client.SessionStorage == nil {
			cfg.Client.SessionStorage = &session.FileStorage{Path: filepath.Join(cfg.Dir, SessionFile)}
		}
		if updCfg.Storage == nil || updCfg.AccessHasher == nil {
			fs := updates.NewFileStorage(filepath.Join(cfg.Dir, UpdatesFile))
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
		updCfg.Logger = cfg.Client.Logger
	}
	if cfg.Peers.Cache == nil {
		// Cache entities (PEER-1), so seen peers are resolved without RPCs.
		cfg.Peers.Cache = &peers.InmemoryCache{}
	}
	if cfg.Peers.Logger == nil {
		cfg.Peers.Logger = cfg.Client.Logger
	}
	maxLoginWait := cfg.MaxLoginWait
	if maxLoginWait == 0 {
		maxLoginWait = defaultMaxLoginWait
	}
	clk := cfg.Client.Clock
	if clk == nil {
		clk = clock.System
	}

	r := &Runtime{
		Clock:        clk,
		maxLoginWait: maxLoginWait,
		isBot:        cfg.IsBot,
		login:        cfg.Login,
		done:         make(chan struct{}),
	}
	// The client needs the updates manager as its handler, and the managers
	// need the API of the client.
	onUpdates := cfg.OnUpdates
	cfg.Client.UpdateHandler = telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		if onUpdates != nil {
			onUpdates(ctx, u)
		}
		return r.Updates.Handle(ctx, u)
	})
	r.Client = telegram.NewClient(appID, appHash, cfg.Client)
	r.API = r.Client.API()
	r.Peers = cfg.Peers.Build(r.API)
	r.Sender = message.NewSender(r.API)

	handler := cfg.Handler
	if handler == nil {
		handler = telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })
	}
	updCfg.Handler = r.Peers.UpdateHook(handler)
	r.Updates = updates.New(updCfg)
	return r, nil
}

// Start connects, signs in with Config.Login and starts the updates manager.
func (r *Runtime) Start(ctx context.Context, so ...telegram.StartOption) error {
	if !r.state.CompareAndSwap(int32(stateIdle), int32(stateStarted)) {
		if lifecycle(r.state.Load()) == stateClosed {
			return telegram.ErrClientClosed
		}
		return telegram.ErrAlreadyStarted
	}
	if err := r.start(ctx, so...); err != nil {
		r.finish(err, true)
		return err
	}
	return nil
}

func (r *Runtime) start(ctx context.Context, so ...telegram.StartOption) error {
	if err := r.Client.Start(ctx, so...); err != nil {
		return errors.Wrap(err, "start client")
	}
	fail := func(err error) error {
		_ = r.Client.Close()
		return err
	}

	self, err := r.login(ctx, r)
	if err != nil {
		return fail(err)
	}
	r.self.Store(self)

	// The context of the updates manager keeps the values of ctx, but its
	// lifetime is the one of the client.
	updCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	ready := make(chan struct{})
	updDone := make(chan error, 1)
	go func() {
		updDone <- r.Updates.Run(updCtx, r.API, self.ID, updates.AuthOptions{
			IsBot:   r.isBot,
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
	case <-r.Client.Done():
		cancel()
		<-updDone
		if err := r.Client.Wait(context.Background()); err != nil {
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
		case <-r.Client.Done():
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
			r.Client.Stop()
			<-r.Client.Done()
		}
		if cerr := r.Client.Wait(context.Background()); cerr != nil {
			// The client error caused the stop.
			err = cerr
		}
		r.finish(err, false)
	}()
	return nil
}

// RetryFlood calls f, waiting out FLOOD_WAIT errors of it up to the total
// MaxLoginWait. Other errors are returned at once.
func (r *Runtime) RetryFlood(ctx context.Context, f func(ctx context.Context) error) error {
	var waited time.Duration
	for {
		err := f(ctx)
		if err == nil {
			return nil
		}
		d, ok := tgerr.AsFloodWait(err)
		if !ok || r.maxLoginWait < 0 || waited+d > r.maxLoginWait {
			return err
		}
		waited += d
		t := r.Clock.Timer(d)
		select {
		case <-t.C():
		case <-ctx.Done():
			t.Stop()
			return errors.Wrap(ctx.Err(), "wait to sign in")
		}
	}
}

func (r *Runtime) finish(err error, startFailed bool) {
	r.doneOnce.Do(func() {
		r.err = err
		r.startFailed = startFailed
		close(r.done)
	})
}

// Self returns the authorized user, nil before Start succeeded.
func (r *Runtime) Self() *tg.User { return r.self.Load() }

// Stop stops without waiting, see telegram.Client.Stop.
func (r *Runtime) Stop() {
	if r.state.CompareAndSwap(int32(stateIdle), int32(stateClosed)) {
		r.finish(nil, false)
	}
	r.Client.Stop()
}

// Shutdown stops and waits until the client and the updates manager are
// stopped or ctx is done. It returns the error that stopped the runtime
// before Shutdown was called, nil if Shutdown stopped it or if Start failed,
// or the wrapped ctx.Err().
func (r *Runtime) Shutdown(ctx context.Context) error {
	r.Stop()
	select {
	case <-r.done:
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "shutdown")
	}
	if r.startFailed {
		return nil
	}
	return r.err
}

// Close is Shutdown without a deadline.
func (r *Runtime) Close() error { return r.Shutdown(context.Background()) }

// Done returns a channel that is closed when the runtime stops.
func (r *Runtime) Done() <-chan struct{} { return r.done }

// Wait blocks until the runtime stops or ctx is done. It returns the error
// that stopped it, nil if it was stopped by Stop or Close, or the wrapped
// ctx.Err().
func (r *Runtime) Wait(ctx context.Context) error {
	select {
	case <-r.done:
		return r.err
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "wait")
	}
}
