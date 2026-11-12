package telegram

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// guestLogOutTimeout bounds auth.logOut on each non-primary DC.
const guestLogOutTimeout = 5 * time.Second

// LogOut logs the client out the way Telegram Desktop and TDLib do:
//
//  1. auth.logOut on the home DC. A 401 error counts as success: the
//     authorization is gone either way.
//  2. auth.logOut on every other DC the client imported the authorization
//     to (pools of *_MIGRATE redirects and of FileDC), best effort; then
//     those pools are closed and their sessions forgotten.
//  3. The stored session is wiped, so the next client started with the same
//     storage creates a new auth key and starts unauthorized.
//
// Like TDLib, LogOut is irrevocable: the local steps run even if auth.logOut
// fails, and its error is returned after them. Lost authorization is not
// reported to OnAuthLost while and after logging out, and the client stops
// saving its session until it is authorized again.
//
// The client keeps running, unauthorized, on its current auth key; sign in
// again or stop it. The auth keys are not destroyed on the server
// (destroy_auth_key).
//
// The result carries auth.loggedOut.future_auth_token, if the server sent
// one: pass it to a later sign-in (auth.sendCode logout_tokens) to skip the
// code.
func (c *Client) LogOut(ctx context.Context) (*tg.AuthLoggedOut, error) {
	c.auth.loggedOut.Store(true)

	res, err := c.API().AuthLogOut(ctx)
	if err != nil {
		if rpcErr, ok := tgerr.As(err); ok && rpcErr.Code == 401 {
			// Already logged out or the authorization was revoked.
			res, err = &tg.AuthLoggedOut{}, nil
		} else {
			err = errors.Wrap(err, "auth.logOut")
		}
	}
	c.auth.authorized.Store(false)

	c.logOutGuests(ctx)
	if wErr := c.wipeSession(ctx); wErr != nil && err == nil {
		err = errors.Wrap(wErr, "wipe session")
	}
	return res, err
}

// logOutGuests logs out on every cached non-primary DC pool and closes them.
func (c *Client) logOutGuests(ctx context.Context) {
	home := 0
	if c.session != nil {
		home = c.session.Load().DC
	}
	seen := map[CloseInvoker]struct{}{}
	for _, cache := range []*dcPoolCache{c.subConns, c.filePools} {
		if cache == nil {
			continue
		}
		cache.mux.Lock()
		conns := make(map[int]CloseInvoker, len(cache.conns))
		for dc, conn := range cache.conns {
			conns[dc] = conn
		}
		cache.conns = map[int]CloseInvoker{}
		cache.mux.Unlock()

		for dc, conn := range conns {
			if _, ok := seen[conn]; ok {
				continue
			}
			seen[conn] = struct{}{}
			if dc != home {
				lctx, cancel := context.WithTimeout(ctx, guestLogOutTimeout)
				if _, err := tg.NewClient(conn).AuthLogOut(lctx); err != nil {
					c.log.Debug(ctx, "Guest DC log out failed", log.Int("dc_id", dc), log.Error(err))
				}
				cancel()
			}
			if err := conn.Close(); err != nil {
				c.log.Debug(ctx, "Close guest DC pool", log.Int("dc_id", dc), log.Error(err))
			}
		}
	}

	c.fileInvokersMux.Lock()
	c.fileInvokers = map[int]tg.Invoker{}
	c.fileInvokersMux.Unlock()

	c.sessionsMux.Lock()
	for dc := range c.sessions {
		delete(c.sessions, dc)
	}
	for dc := range c.dcAuthorized {
		delete(c.dcAuthorized, dc)
	}
	for dc := range c.sessionStates {
		delete(c.sessionStates, dc)
	}
	c.sessionsMux.Unlock()
	c.selfID.Store(0)
	c.selfBot.Store(false)
}

// wipeSession clears the stored session: the session loader treats empty
// data as no session.
func (c *Client) wipeSession(ctx context.Context) error {
	l, ok := c.storage.(*session.Loader)
	if !ok || l == nil || l.Storage == nil {
		return nil
	}
	return l.Storage.StoreSession(ctx, nil)
}
