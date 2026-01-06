// Package takeout provides a wrapper for Telegram takeout sessions.
//
// Takeout sessions allow exporting user data from Telegram. See
// https://core.telegram.org/api/takeout for more information.
package takeout

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// Client wraps tg.Invoker to use a takeout session.
type Client struct {
	id  int64
	raw tg.Invoker
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
}

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
func Run(ctx context.Context, invoker tg.Invoker, cfg Config, f func(ctx context.Context, client *Client) error) error {
	raw := tg.NewClient(invoker)

	takeout, err := raw.AccountInitTakeoutSession(ctx, cfg.request())
	if err != nil {
		return errors.Wrap(err, "init takeout session")
	}

	client := &Client{
		id:  takeout.ID,
		raw: invoker,
	}

	fnErr := f(ctx, client)

	finishReq := &tg.AccountFinishTakeoutSessionRequest{}
	finishReq.SetSuccess(fnErr == nil)

	if _, err := raw.AccountFinishTakeoutSession(ctx, finishReq); err != nil {
		if fnErr != nil {
			return errors.Wrap(fnErr, "function failed")
		}
		return errors.Wrap(err, "finish takeout session")
	}

	return fnErr
}
