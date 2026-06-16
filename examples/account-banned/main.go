// Binary account-banned shows how to reliably detect a banned/deactivated
// account and distinguish it from a dead proxy or network outage.
//
// Why this is not trivial: USER_DEACTIVATED (and friends) is a 401 error. On a
// long-lived client a ban usually tears the connection down at the transport
// level, so a side getMe lands in the reconnect window and returns
// "connection dead" — indistinguishable from a dead proxy. Polling getMe is
// therefore unreliable.
//
// Instead, catch the 401 at its source. On every (re)connect the client calls
// Self(); the OnSelfError hook receives that error. Branch on it:
//   - 401 (auth.IsUnauthorized) -> account is dead, return the error to make
//     reconnection permanent and stop the client.
//   - anything else -> network/proxy, return nil to keep retrying.
package main

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// isAccountDead reports whether err means the account itself is gone and no
// amount of reconnecting will help. All of these are 401 errors, but we also
// match by type for clearer intent and logging.
func isAccountDead(err error) bool {
	return auth.IsUnauthorized(err) || tgerr.Is(err,
		"USER_DEACTIVATED",     // account deleted by the user.
		"USER_DEACTIVATED_BAN", // account banned by Telegram.
		"AUTH_KEY_UNREGISTERED",
		"SESSION_REVOKED",
		"SESSION_EXPIRED",
	)
}

func main() {
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// banned is set the moment we observe a dead-account error, from any
		// source (OnSelfError or an update handler). Use atomic.Bool because the
		// callbacks run on different goroutines than your main logic.
		var banned atomic.Bool

		dispatcher := tg.NewUpdateDispatcher()
		opts := telegram.Options{
			Logger:        logzap.New(log),
			UpdateHandler: dispatcher,

			// OnSelfError is the reliable signal. It is invoked with the error
			// from the client's own Self() call on each (re)connect.
			OnSelfError: func(ctx context.Context, err error) error {
				if isAccountDead(err) {
					log.Warn("Account is dead, stopping reconnection", zap.Error(err))
					banned.Store(true)
					// Returning a non-nil error makes the error permanent
					// (see isPermanentError): reconnection stops and Run returns.
					return err
				}
				// Network / proxy issue: keep reconnecting silently.
				log.Debug("Self failed, will retry", zap.Error(err))
				return nil
			},
		}

		return telegram.BotFromEnvironment(ctx, opts, func(ctx context.Context, client *telegram.Client) error {
			// You can also catch the ban directly where it first arrives — e.g.
			// in your reactions/message handler — and flip the same flag.
			dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
				_, err := client.API().UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
				if isAccountDead(err) {
					log.Warn("Account is dead (seen in handler)", zap.Error(err))
					banned.Store(true)
					return err
				}
				return err
			})

			// In your own code, prefer the banned flag over probing getMe:
			//
			//	if banned.Load() {
			//		// mark the account dead in your storage, drop the worker, etc.
			//	}
			return nil
		}, telegram.RunUntilCanceled)
	})
}
