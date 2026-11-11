package telegram

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/go-faster/errors"
)

// lifecycleMode is how a Client is driven.
type lifecycleMode int32

const (
	// modeIdle means that neither Run nor Start was called.
	modeIdle lifecycleMode = iota
	// modeRun means that the client is driven by Run.
	modeRun
	// modeStart means that the client is driven by Start and Close.
	modeStart
	// modeClosed means that Close was called before Start.
	modeClosed
)

var (
	// ErrAlreadyStarted is returned by Start and Run if the client was
	// already started by Start or Run: a Client is single-use, and Start and
	// Run are mutually exclusive. It is also returned by Close and Shutdown
	// of a client driven by Run.
	ErrAlreadyStarted = errors.New("client already started")
	// ErrClientClosed is returned by Start and Run after Close, Shutdown or
	// Stop of a client that was never started, and by Start canceled by them.
	ErrClientClosed = errors.New("client already closed")

	errDrivenByRun   = errors.Wrap(ErrAlreadyStarted, "client is driven by Run")
	errDrivenByStart = errors.Wrap(ErrAlreadyStarted,
		"client is driven by Start, use Close and Wait instead of Run")
)

// lifecycle is the state of Start, Close, Done and Wait.
type lifecycle struct {
	mode atomic.Int32 // lifecycleMode

	mux sync.Mutex
	// done is closed after the client stops, lazily created by doneChan.
	done chan struct{}
	// cancel stops the Run started by Start, nil until it is created.
	cancel context.CancelFunc
	// closing is set by Stop, so that Start cancels Run at once if Stop
	// was called before cancel was set.
	closing bool
	// err is the error that stopped the client, written before done is
	// closed.
	err error
	// startFailed is set when Start returned an error, written before done
	// is closed.
	startFailed bool
}

// doneChan returns the done channel, creating it if needed.
func (l *lifecycle) doneChan() chan struct{} {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.done == nil {
		l.done = make(chan struct{})
	}
	return l.done
}

// startOptions are options of Start.
type startOptions struct {
	lifetime context.Context
}

// StartOption configures Client.Start.
type StartOption interface {
	applyStart(o *startOptions)
}

type startOptionFunc func(o *startOptions)

func (f startOptionFunc) applyStart(o *startOptions) { f(o) }

// WithLifetime sets the parent context of the client started by Start: when
// ctx is done, the client stops as if Stop was called, and Wait and Close
// return nil.
//
// By default the lifetime is context.WithoutCancel of the ctx passed to Start,
// so values like loggers and trace spans are kept, but the cancellation and
// deadline of the startup ctx do not stop the started client.
func WithLifetime(ctx context.Context) StartOption {
	return startOptionFunc(func(o *startOptions) {
		if ctx != nil {
			o.lifetime = ctx
		}
	})
}

// Start connects the client in the background and blocks until it is ready
// for requests, which is the moment Run would call its callback, or until it
// fails.
//
// The ctx bounds the startup only, like in net.Dialer.DialContext: once Start
// returned, the client runs until Stop or Close, a fatal error (e.g.
// *AuthLostError or a permanent connection error) or the end of the context
// passed via WithLifetime. Note that the client reconnects indefinitely on
// transient errors, so use ctx with a deadline to bound the startup.
//
// Start returns the wrapped ctx.Err() if ctx is done before the client is
// ready, or the error that stopped the client before it got ready, e.g.
// *ConfigError. On error the client is fully stopped: no goroutines are left
// behind, and Close returns nil at once.
//
// The client is valid for requests, e.g. via API, between a successful Start
// and Close, as inside the Run callback. Use Done or Wait to learn that it
// stopped, and Close to release resources:
//
//	if err := client.Start(ctx); err != nil {
//		return err
//	}
//	defer func() { _ = client.Close() }()
//
// Start and Run are mutually exclusive, and a Client can be started once:
// create a new Client to start again. Start supersedes the
// github.com/gotd/contrib/bg package.
func (c *Client) Start(ctx context.Context, opts ...StartOption) error {
	if !c.life.mode.CompareAndSwap(int32(modeIdle), int32(modeStart)) {
		switch lifecycleMode(c.life.mode.Load()) {
		case modeRun:
			return errDrivenByRun
		case modeClosed:
			return ErrClientClosed
		default:
			return ErrAlreadyStarted
		}
	}
	if err := ctx.Err(); err != nil {
		err = errors.Wrap(err, "start")
		c.finishStart(err, true)
		return err
	}

	o := startOptions{lifetime: context.WithoutCancel(ctx)}
	for _, opt := range opts {
		opt.applyStart(&o)
	}
	runCtx, cancel := context.WithCancel(o.lifetime)
	c.life.mux.Lock()
	c.life.cancel = cancel
	if c.life.closing {
		// Close was called concurrently.
		cancel()
	}
	c.life.mux.Unlock()

	ready := make(chan struct{})
	runDone := make(chan error, 1)
	go func() {
		defer cancel()
		err := c.run(runCtx, func(ctx context.Context) error {
			// Run calls f only once the primary connection is ready.
			close(ready)
			// Keep Run alive until Stop, the end of the lifetime or a
			// fatal error.
			<-ctx.Done()
			return nil
		})
		if lifeErr := o.lifetime.Err(); lifeErr != nil && errors.Is(err, lifeErr) {
			// The end of the lifetime is a clean stop, e.g. on deadline,
			// unless the authorization was lost.
			var lost *AuthLostError
			if !errors.As(err, &lost) {
				err = nil
			}
		}
		runDone <- err
	}()

	select {
	case <-ready:
		go func() {
			c.finishStart(<-runDone, false)
		}()
		return nil
	case err := <-runDone:
		select {
		case <-ready:
			// Got ready, then stopped: report the stop via Done and Wait.
			c.finishStart(err, false)
			return nil
		default:
		}
		if err == nil {
			// Stopped by Close or by the end of the lifetime before the
			// client got ready.
			c.life.mux.Lock()
			closing := c.life.closing
			c.life.mux.Unlock()
			switch lifeErr := o.lifetime.Err(); {
			case closing:
				err = errors.Wrap(ErrClientClosed, "start")
			case lifeErr != nil:
				err = errors.Wrap(lifeErr, "start")
			default:
				err = errors.New("start: client stopped")
			}
		}
		c.finishStart(err, true)
		return err
	case <-ctx.Done():
		select {
		case <-ready:
			// Readiness wins, see waitReady.
			go func() {
				c.finishStart(<-runDone, false)
			}()
			return nil
		default:
		}
		cancel()
		<-runDone
		err := errors.Wrap(ctx.Err(), "start")
		c.finishStart(err, true)
		return err
	}
}

// finishStart records the error that stopped the client and closes done.
func (c *Client) finishStart(err error, startFailed bool) {
	done := c.life.doneChan()
	c.life.mux.Lock()
	c.life.err = err
	c.life.startFailed = startFailed
	c.life.mux.Unlock()
	close(done)
}

// Stop stops the client started by Start without waiting for it: use Done,
// Wait or Close to wait. Unlike Close and Shutdown, Stop is safe to call from
// update handlers, OnAuthLost and other callbacks running inside the client.
//
// Stop is idempotent and safe to call concurrently, also before Start (then
// the client can not be started anymore). Stop does nothing for a client
// driven by Run: cancel the Run context or return from its callback instead.
func (c *Client) Stop() {
	c.life.mode.CompareAndSwap(int32(modeIdle), int32(modeClosed))
	switch lifecycleMode(c.life.mode.Load()) {
	case modeRun:
		return
	case modeClosed:
		done := c.life.doneChan()
		c.life.mux.Lock()
		select {
		case <-done:
		default:
			close(done)
		}
		c.life.mux.Unlock()
		return
	}

	c.life.mux.Lock()
	c.life.closing = true
	cancel := c.life.cancel
	c.life.mux.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Shutdown stops the client started by Start, like Stop, and waits until all
// of its background work (connections, pools, update handlers and OnAuthLost
// callbacks) is finished or ctx is done, like http.Server.Shutdown.
//
// It returns the error that stopped the client before Shutdown was called,
// e.g. *AuthLostError (the same as Wait), or nil if Shutdown stopped it or if
// Start failed, as Start already returned the error. If ctx is done first, it
// returns the wrapped ctx.Err(), and the client keeps stopping in background.
//
// Do not call Shutdown or Close from update handlers, OnAuthLost or other
// callbacks running inside the client: the client waits for them to return,
// so the call blocks until ctx is done. Use Stop there.
//
// Shutdown returns ErrAlreadyStarted for a client driven by Run.
func (c *Client) Shutdown(ctx context.Context) error {
	if lifecycleMode(c.life.mode.Load()) == modeRun {
		return errDrivenByRun
	}
	c.Stop()
	select {
	case <-c.life.doneChan():
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "shutdown")
	}

	c.life.mux.Lock()
	defer c.life.mux.Unlock()
	if c.life.startFailed {
		return nil
	}
	return c.life.err
}

// Close is Shutdown without a deadline: it stops the client started by Start
// and waits until all of its background work is finished.
//
// It returns the error that stopped the client before Close was called, e.g.
// *AuthLostError (the same as Wait), or nil if Close stopped it or if Start
// failed.
//
// Close is idempotent and safe to call concurrently, also before Start (then
// the client can not be started anymore). Never call Close from update
// handlers, OnAuthLost or other callbacks running inside the client, it would
// deadlock: use Stop there. Close returns ErrAlreadyStarted for a client
// driven by Run.
func (c *Client) Close() error {
	return c.Shutdown(context.Background())
}

// Done returns a channel that is closed when the client stops for any reason:
// Stop or Close, a fatal error, the end of the lifetime context, a failed Start
// or the return of Run. Use Wait to get the reason.
func (c *Client) Done() <-chan struct{} {
	return c.life.doneChan()
}

// Wait blocks until the client stops or ctx is done.
//
// It returns the error that stopped the client, e.g. *AuthLostError, a
// permanent connection error, the error returned by a failed Start or by
// Run, nil if it was stopped by Stop or Close or by the end of the lifetime
// context, or the wrapped ctx.Err() if ctx is done first.
//
// Unlike Close, Wait does not stop the client, so it is the blocking "run
// until disconnected" call of the Start style:
//
//	return client.Wait(ctx)
func (c *Client) Wait(ctx context.Context) error {
	select {
	case <-c.life.doneChan():
		c.life.mux.Lock()
		defer c.life.mux.Unlock()
		return c.life.err
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "wait")
	}
}

// Connect creates a new client with NewClient and starts it with Start.
//
// The ctx bounds the startup only, see Start. Call Close on the returned
// client to stop it. On error the client is fully stopped and nil is
// returned.
func Connect(ctx context.Context, appID int, appHash string, opt Options, so ...StartOption) (*Client, error) {
	c := NewClient(appID, appHash, opt)
	if err := c.Start(ctx, so...); err != nil {
		return nil, err
	}
	return c, nil
}
