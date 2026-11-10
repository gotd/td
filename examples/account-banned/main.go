// Binary account-banned shows how to detect a banned, deleted or frozen
// account and tell it from a terminated session or a network outage.
//
// Polling getMe is unreliable: a ban often drops the auth key, so the
// connection is reset and getMe lands in the reconnect window. Instead, the
// client watches results of all requests:
//   - OnAuthLost is called once when an authorized session gets a 401 error
//     (USER_DEACTIVATED_BAN, SESSION_REVOKED, ...) or an AUTH_KEY_DROP_*
//     service notification. AuthLostKindOf tells why.
//   - OnFrozen is called when a request fails with FROZEN_METHOD_INVALID and
//     app config confirms that the account is frozen. A frozen account stays
//     signed in and can read, but most methods fail.
//
// Run with the session of a signed in user (see telegram.ClientFromEnvironment).
package main

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
)

func main() {
	examples.Run(func(ctx context.Context, lg *zap.Logger) error {
		opts := telegram.Options{
			Logger: logzap.New(lg),
			// Called once per lost authorization, on its own goroutine.
			OnAuthLost: func(ctx context.Context, err error) error {
				kind := telegram.AuthLostKindOf(err)
				lg.Warn("Authorization lost", zap.Stringer("kind", kind), zap.Error(err))
				switch kind {
				case telegram.AuthLostBanned, telegram.AuthLostDeleted:
					// Mark the account dead in your storage: signing in
					// again won't help.
				default:
					// Session terminated: the account may sign in again.
				}
				// Stop Run with *telegram.AuthLostError.
				return err
			},
			OnFrozen: func(ctx context.Context, state telegram.FreezeState) {
				lg.Warn("Account is frozen",
					zap.Time("since", state.Since),
					zap.Time("until", state.Until),
					zap.String("appeal_url", state.AppealURL),
				)
			},
		}

		client, err := telegram.ClientFromEnvironment(opts)
		if err != nil {
			return errors.Wrap(err, "create client")
		}
		err = client.Run(ctx, func(ctx context.Context) error {
			// Check freeze state at start: OnFrozen is called only after
			// a request fails.
			state, err := client.FreezeState(ctx)
			if err != nil {
				return errors.Wrap(err, "freeze state")
			}
			lg.Info("Started", zap.Bool("frozen", state.Frozen()))

			<-ctx.Done()
			return ctx.Err()
		})

		var lost *telegram.AuthLostError
		if errors.As(err, &lost) {
			lg.Error("Stopped",
				zap.Stringer("kind", lost.Kind()),
				zap.Bool("banned", telegram.IsAccountBanned(lost.Err)),
			)
			return nil
		}
		return err
	})
}
