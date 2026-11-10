// Package qrlogin provides QR login flow implementation.
//
// See https://core.telegram.org/api/qr-login.
package qrlogin

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// QR implements Telegram QR login flow.
type QR struct {
	api      *tg.Client
	appID    int
	appHash  string
	migrate  func(ctx context.Context, dcID int) error
	clock    clock.Clock
	password func(ctx context.Context) (string, error)
}

// NewQR creates new QR
func NewQR(api *tg.Client, appID int, appHash string, opts Options) QR {
	opts.setDefaults()
	return QR{
		api:      api,
		appID:    appID,
		appHash:  appHash,
		clock:    opts.Clock,
		migrate:  opts.Migrate,
		password: opts.Password,
	}
}

// ExportResult is a result of exporting login token.
//
// Exactly one field is set.
type ExportResult struct {
	// Token is set for auth.loginToken: show it and wait for acceptance.
	Token Token
	// Authorization is set for auth.loginTokenSuccess: token was already
	// accepted and login is complete.
	Authorization *tg.AuthAuthorization
	// MigrateTo is set for auth.loginTokenMigrateTo: login should be
	// completed via auth.importLoginToken on given DC.
	MigrateTo *tg.AuthLoginTokenMigrateTo
}

// ExportResult exports new login token, returning authorization if token
// was already accepted.
//
// See https://core.telegram.org/api/qr-login#exporting-a-login-token.
func (q QR) ExportResult(ctx context.Context, exceptIDs ...int64) (ExportResult, error) {
	result, err := q.export(ctx, exceptIDs...)
	if err != nil {
		return ExportResult{}, err
	}

	switch t := result.(type) {
	case *tg.AuthLoginToken:
		return ExportResult{Token: NewToken(t.Token, t.Expires)}, nil
	case *tg.AuthLoginTokenSuccess:
		a, ok := t.Authorization.(*tg.AuthAuthorization)
		if !ok {
			return ExportResult{}, errors.Errorf("unexpected type %T", t.Authorization)
		}
		return ExportResult{Authorization: a}, nil
	case *tg.AuthLoginTokenMigrateTo:
		return ExportResult{MigrateTo: t}, nil
	default:
		return ExportResult{}, errors.Errorf("unexpected type %T", result)
	}
}

// Export exports new login token.
//
// If token was already accepted (auth.loginTokenSuccess), Export returns an
// empty token and the authorization is discarded; use ExportResult to get it.
// If migration is needed, Export returns *MigrationNeededError.
//
// See https://core.telegram.org/api/qr-login#exporting-a-login-token.
func (q QR) Export(ctx context.Context, exceptIDs ...int64) (Token, error) {
	result, err := q.export(ctx, exceptIDs...)
	if err != nil {
		return Token{}, err
	}

	switch t := result.(type) {
	case *tg.AuthLoginToken:
		return NewToken(t.Token, t.Expires), nil
	case *tg.AuthLoginTokenSuccess:
		// Token was already accepted, authentication successful.
		return Token{}, nil
	case *tg.AuthLoginTokenMigrateTo:
		return Token{}, &MigrationNeededError{
			MigrateTo: t,
		}
	default:
		return Token{}, errors.Errorf("unexpected type %T", result)
	}
}

func (q QR) export(ctx context.Context, exceptIDs ...int64) (tg.AuthLoginTokenClass, error) {
	result, err := q.api.AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
		APIID:     q.appID,
		APIHash:   q.appHash,
		ExceptIDs: exceptIDs,
	})
	if err != nil {
		return nil, errors.Wrap(err, "export")
	}
	return result, nil
}

// Accept accepts given token.
//
// See https://core.telegram.org/api/qr-login#accepting-a-login-token.
func (q QR) Accept(ctx context.Context, t Token) (*tg.Authorization, error) {
	return AcceptQR(ctx, q.api, t)
}

// Import imports accepted token.
//
// If the account has 2FA password and Options.Password is set, the password
// is checked, otherwise SESSION_PASSWORD_NEEDED error is returned.
//
// See https://core.telegram.org/api/qr-login#confirming-importing-the-login-token.
func (q QR) Import(ctx context.Context) (*tg.AuthAuthorization, error) {
	a, err := q.importToken(ctx)
	if q.needPassword(err) {
		return q.checkPassword(ctx)
	}
	return a, err
}

// needPassword reports whether err requests 2FA password and it can be
// provided.
func (q QR) needPassword(err error) bool {
	return err != nil && q.password != nil && tgerr.Is(err, "SESSION_PASSWORD_NEEDED")
}

func (q QR) checkPassword(ctx context.Context) (*tg.AuthAuthorization, error) {
	password, err := q.password(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get password")
	}
	a, err := auth.NewClient(q.api, crypto.DefaultRand(), q.appID, q.appHash).Password(ctx, password)
	if err != nil {
		return nil, errors.Wrap(err, "check password")
	}
	return a, nil
}

func (q QR) importToken(ctx context.Context) (*tg.AuthAuthorization, error) {
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

	r, err := q.ExportResult(ctx, exceptIDs...)
	if err != nil {
		if q.needPassword(err) {
			return q.checkPassword(ctx)
		}
		return nil, err
	}
	switch {
	case r.Authorization != nil:
		// Token was already accepted (e.g. before restart on the same
		// session): the update has already happened, do not wait for it.
		return r.Authorization, nil
	case r.MigrateTo != nil:
		return nil, &MigrationNeededError{MigrateTo: r.MigrateTo}
	}
	token := r.Token

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
			r, err := q.ExportResult(ctx, exceptIDs...)
			if err != nil {
				// Token may be accepted right before expiration (or the
				// update is missed), so 2FA can be requested here too.
				if q.needPassword(err) {
					return q.checkPassword(ctx)
				}
				return nil, err
			}
			switch {
			case r.Authorization != nil:
				// QR was scanned and accepted, no need to export again.
				return r.Authorization, nil
			case r.MigrateTo != nil:
				return nil, &MigrationNeededError{MigrateTo: r.MigrateTo}
			}

			token = r.Token
			timer.Reset(until(token))
			continue
		case <-loggedIn:
		}

		return q.Import(ctx)
	}
}
