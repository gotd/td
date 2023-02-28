// Package qrlogin provides QR login flow implementation.
//
// See https://core.telegram.org/api/qr-login.
package qrlogin

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/tg"
)

// QR implements Telegram QR login flow.
type QR struct {
	api     *tg.Client
	appID   int
	appHash string
	migrate func(ctx context.Context, dcID int) error
	clock   clock.Clock
}

// NewQR creates new QR
func NewQR(api *tg.Client, appID int, appHash string, opts Options) QR {
	opts.setDefaults()
	return QR{
		api:     api,
		appID:   appID,
		appHash: appHash,
		clock:   opts.Clock,
		migrate: opts.Migrate,
	}
}

// Export exports new login token.
//
// See https://core.telegram.org/api/qr-login#exporting-a-login-token.
func (q QR) Export(ctx context.Context, exceptIDs ...int64) (Token, error) {
	result, err := q.api.AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
		APIID:     q.appID,
		APIHash:   q.appHash,
		ExceptIDs: exceptIDs,
	})
	if err != nil {
		return Token{}, errors.Wrap(err, "export")
	}

	t, ok := result.(*tg.AuthLoginToken)
	if !ok {
		return Token{}, errors.Errorf("unexpected type %T", result)
	}
	return NewToken(t.Token, t.Expires), nil
}

// Accept accepts given token.
//
// See https://core.telegram.org/api/qr-login#accepting-a-login-token.
func (q QR) Accept(ctx context.Context, t Token) (*tg.Authorization, error) {
	return AcceptQR(ctx, q.api, t)
}

// Import imports accepted token.
//
// See https://core.telegram.org/api/qr-login#confirming-importing-the-login-token.
func (q QR) Import(ctx context.Context) (*tg.AuthAuthorization, error) {
	result, err := q.api.AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
		APIID:   q.appID,
		APIHash: q.appHash,
	})
	if err != nil {
		return nil, errors.Wrap(err, "import")
	}

	switch t := result.(type) {
	case *tg.AuthLoginTokenMigrateTo:
		if q.migrate == nil {
			return nil, &MigrationNeededError{
				MigrateTo: t,
			}
		}
		if err := q.migrate(ctx, t.DCID); err != nil {
			return nil, errors.Wrap(err, "migrate")
		}

		res, err := q.api.AuthImportLoginToken(ctx, t.Token)
		if err != nil {
			return nil, errors.Wrap(err, "import")
		}

		success, ok := res.(*tg.AuthLoginTokenSuccess)
		if !ok {
			return nil, errors.Errorf("unexpected type %T", res)
		}

		auth, ok := success.Authorization.(*tg.AuthAuthorization)
		if !ok {
			return nil, errors.Errorf("unexpected type %T", success.Authorization)
		}
		return auth, nil
	case *tg.AuthLoginTokenSuccess:
		auth, ok := t.Authorization.(*tg.AuthAuthorization)
		if !ok {
			return nil, errors.Errorf("unexpected type %T", t.Authorization)
		}
		return auth, nil
	default:
		return nil, errors.Errorf("unexpected type %T", result)
	}
}

// LoggedIn is signal channel to notify about tg.UpdateLoginToken.
type LoggedIn <-chan struct{}

// OnLoginToken sets handler for given dispatcher and returns signal channel.
func OnLoginToken(d interface {
	OnLoginToken(tg.LoginTokenHandler)
},
) LoggedIn {
	loggedIn := make(chan struct{})
	d.OnLoginToken(func(ctx context.Context, e tg.Entities, update *tg.UpdateLoginToken) error {
		select {
		case loggedIn <- struct{}{}:
			return nil
		default:
		}
		return nil
	})
	return loggedIn
}

// Auth generates new QR login token, shows it and awaits acceptation.
//
// NB: Show callback may be called more than once if QR expires.
func (q QR) Auth(
	ctx context.Context,
	loggedIn LoggedIn,
	show func(ctx context.Context, token Token) error,
	exceptIDs ...int64,
) (*tg.AuthAuthorization, error) {
	until := func(token Token) time.Duration {
		return token.Expires().Sub(q.clock.Now()).Truncate(time.Second)
	}

	token, err := q.Export(ctx, exceptIDs...)
	if err != nil {
		return nil, err
	}
	timer := q.clock.Timer(until(token))
	defer clock.StopTimer(timer)

	for {
		if err := show(ctx, token); err != nil {
			return nil, errors.Wrap(err, "show")
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C():
			t, err := q.Export(ctx, exceptIDs...)
			if err != nil {
				return nil, err
			}
			token = t
			timer.Reset(until(token))

			continue
		case <-loggedIn:
		}

		return q.Import(ctx)
	}
}
