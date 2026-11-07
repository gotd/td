// Package takeout provides a wrapper for Telegram takeout sessions.
//
// Takeout sessions allow exporting user data from Telegram. See
// https://core.telegram.org/api/takeout for more information.
package takeout

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// Client wraps tg.Invoker to use a takeout session.
type Client struct {
	id  int64
	raw tg.Invoker
}

// NewClient creates Client for existing takeout session with given ID, e.g.
// to resume a takeout session saved before restart.
//
// Telegram keeps the session until it is finished, so long-running exports
// can persist Client.ID and resume with it instead of starting a new session
// (which may require confirmation by the user again).
func NewClient(id int64, invoker tg.Invoker) *Client {
	return &Client{id: id, raw: invoker}
}

// Finish finishes takeout session.
func (c *Client) Finish(ctx context.Context, success bool) error {
	req := &tg.AccountFinishTakeoutSessionRequest{}
	req.SetSuccess(success)
	if _, err := tg.NewClient(c.raw).AccountFinishTakeoutSession(ctx, req); err != nil {
		return errors.Wrap(err, "finish takeout session")
	}
	return nil
}

// Invoke implements tg.Invoker.
func (c *Client) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	obj, ok := input.(bin.Object)
	if !ok {
		return errors.Errorf("input %T does not implement bin.Object", input)
	}
	return c.raw.Invoke(ctx, &tg.InvokeWithTakeoutRequest{
		TakeoutID: c.id,
		Query:     obj,
	}, output)
}

// ID returns takeout session ID.
func (c *Client) ID() int64 {
	return c.id
}

// Raw returns underlying invoker.
func (c *Client) Raw() tg.Invoker {
	return c.raw
}

// Config configures takeout session.
type Config struct {
	// Contacts enables exporting contacts.
	Contacts bool
	// MessageUsers enables exporting messages from private chats.
	MessageUsers bool
	// MessageChats enables exporting messages from basic groups.
	MessageChats bool
	// MessageMegagroups enables exporting messages from supergroups.
	MessageMegagroups bool
	// MessageChannels enables exporting messages from channels.
	MessageChannels bool
	// Files enables exporting files.
	Files bool
	// FileMaxSize sets maximum file size to export.
	// Only used if Files is true.
	FileMaxSize int64

	// OnInitDelay is called when init fails with TAKEOUT_INIT_DELAY_X: the
	// user must confirm the export on another device, or wait d (X seconds)
	// for the session to become available (optional).
	//
	// Return nil to keep polling init every InitPollInterval (or d, if it is
	// shorter) until it succeeds, or an error to abort. If not set, the error
	// is returned.
	OnInitDelay func(ctx context.Context, d time.Duration) error
	// InitPollInterval is the interval of init attempts while waiting for
	// the confirmation. Defaults to 30 seconds.
	InitPollInterval time.Duration
	// Clock is the time source. Defaults to clock.System.
	Clock clock.Clock
}

// defaultInitPollInterval is the default Config.InitPollInterval.
const defaultInitPollInterval = 30 * time.Second

// AsInitDelay returns delay of TAKEOUT_INIT_DELAY_X error.
func AsInitDelay(err error) (time.Duration, bool) {
	rpcErr, ok := tgerr.AsType(err, "TAKEOUT_INIT_DELAY")
	if !ok {
		return 0, false
	}
	return time.Duration(rpcErr.Argument) * time.Second, true
}

// Init initializes new takeout session.
//
// See Config.OnInitDelay for TAKEOUT_INIT_DELAY handling.
func Init(ctx context.Context, invoker tg.Invoker, cfg Config) (*Client, error) {
	raw := tg.NewClient(invoker)
	clk := cfg.Clock
	if clk == nil {
		clk = clock.System
	}
	poll := cfg.InitPollInterval
	if poll <= 0 {
		poll = defaultInitPollInterval
	}

	for {
		takeout, err := raw.AccountInitTakeoutSession(ctx, cfg.request())
		if err == nil {
			return NewClient(takeout.ID, invoker), nil
		}
		delay, ok := AsInitDelay(err)
		if !ok || cfg.OnInitDelay == nil {
			return nil, errors.Wrap(err, "init takeout session")
		}
		if cbErr := cfg.OnInitDelay(ctx, delay); cbErr != nil {
			return nil, errors.Wrap(cbErr, "init delay")
		}

		wait := poll
		if delay > 0 && delay < wait {
			wait = delay
		}
		timer := clk.Timer(wait)
		select {
		case <-ctx.Done():
			clock.StopTimer(timer)
			return nil, errors.Wrap(ctx.Err(), "wait for takeout confirmation")
		case <-timer.C():
		}
	}
}

// finishTimeout bounds finishing session in Run.
const finishTimeout = 10 * time.Second

func (c Config) request() *tg.AccountInitTakeoutSessionRequest {
	r := &tg.AccountInitTakeoutSessionRequest{}
	r.SetContacts(c.Contacts)
	r.SetMessageUsers(c.MessageUsers)
	r.SetMessageChats(c.MessageChats)
	r.SetMessageMegagroups(c.MessageMegagroups)
	r.SetMessageChannels(c.MessageChannels)
	r.SetFiles(c.Files)
	if c.FileMaxSize > 0 {
		r.SetFileMaxSize(c.FileMaxSize)
	}
	return r
}

// Run initializes a takeout session and calls the provided function.
//
// The session is automatically finished when the function returns. If the
// function returns nil, the session is marked as successful. Otherwise, the
// session is marked as failed.
//
// The session is finished even if ctx is canceled (with a short timeout).
func Run(ctx context.Context, invoker tg.Invoker, cfg Config, f func(ctx context.Context, client *Client) error) error {
	client, err := Init(ctx, invoker, cfg)
	if err != nil {
		return err
	}

	fnErr := f(ctx, client)

	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), finishTimeout)
	defer cancel()
	if err := client.Finish(finishCtx, fnErr == nil); err != nil {
		if fnErr != nil {
			return errors.Wrap(fnErr, "function failed")
		}
		return err
	}

	return fnErr
}
