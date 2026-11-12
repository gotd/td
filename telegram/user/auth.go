package user

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
)

// Authenticator signs in an unauthorized session: Phone, QR or AuthFunc.
//
// It is called at most once per Start, plus once more after each FLOOD_WAIT
// (bounded by Options.MaxLoginWait); other errors fail Start.
type Authenticator interface {
	signIn(ctx context.Context, c signInContext) error
	// retryFlood reports whether a FLOOD_WAIT of the last signIn may be
	// waited out and signIn called again.
	retryFlood() bool
}

// signInContext is what an Authenticator gets.
type signInContext struct {
	client  *telegram.Client
	appID   int
	appHash string
	tokens  *loginTokens
}

// Phone signs in with a phone code (and the 2FA password if the account has
// one), see auth.NewFlow.
//
// A FLOOD_WAIT is waited out (up to Options.MaxLoginWait) only before a code
// was requested from the authenticator: after that, signing in again would
// send a new code, so the error is returned to the caller.
func Phone(a auth.UserAuthenticator, opt auth.SendCodeOptions) Authenticator {
	return &phoneAuth{auth: a, opt: opt}
}

type phoneAuth struct {
	auth auth.UserAuthenticator
	opt  auth.SendCodeOptions
	// codeAsked is set once the authenticator was asked for a code, i.e.
	// auth.sendCode succeeded.
	codeAsked atomic.Bool
}

func (p *phoneAuth) signIn(ctx context.Context, c signInContext) error {
	p.codeAsked.Store(false)
	return auth.NewFlow(codeTracker{UserAuthenticator: p.auth, asked: &p.codeAsked}, p.opt).Run(ctx, c.client.Auth())
}

func (p *phoneAuth) retryFlood() bool { return !p.codeAsked.Load() }

// codeTracker records that the code was requested.
type codeTracker struct {
	auth.UserAuthenticator
	asked *atomic.Bool
}

func (t codeTracker) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	t.asked.Store(true)
	return t.UserAuthenticator.Code(ctx, sentCode)
}

// QROptions are options of QR.
type QROptions struct {
	// Show displays the login token, e.g. as a QR code of token.URL() to scan
	// with Telegram on another device (Settings > Devices > Link Desktop
	// Device). It is called again with a new token when the shown one
	// expires. Required.
	Show func(ctx context.Context, token qrlogin.Token) error
	// Password returns the 2FA password of the account, if it has one
	// (optional; without it an account with 2FA fails with
	// SESSION_PASSWORD_NEEDED).
	Password func(ctx context.Context) (string, error)
}

// QR signs in by a login token accepted on another device, see
// qrlogin.QR.Auth. Tokens accepted by an account of another DC are imported
// there.
func QR(opt QROptions) Authenticator {
	return qrAuth{opt: opt}
}

type qrAuth struct {
	opt QROptions
}

// retryFlood: waiting out a FLOOD_WAIT of exporting a login token is safe.
func (q qrAuth) retryFlood() bool { return true }

func (q qrAuth) signIn(ctx context.Context, c signInContext) error {
	if q.opt.Show == nil {
		return errors.New("QROptions.Show is nil")
	}
	qr := qrlogin.NewQR(c.client.API(), c.appID, c.appHash, qrlogin.Options{
		Migrate:  c.client.MigrateTo,
		Password: q.opt.Password,
	})
	loggedIn := qrlogin.OnLoginToken(c.tokens)
	defer c.tokens.reset()
	_, err := qr.Auth(ctx, loggedIn, q.opt.Show)
	return err
}

// AuthFunc is an Authenticator calling f, e.g. for a custom sign in with
// client.Auth().
func AuthFunc(f func(ctx context.Context, client *telegram.Client) error) Authenticator {
	return funcAuth(f)
}

type funcAuth func(ctx context.Context, client *telegram.Client) error

// retryFlood: f may have sent a login code, so it is never called again.
func (f funcAuth) retryFlood() bool { return false }

func (f funcAuth) signIn(ctx context.Context, c signInContext) error {
	return f(ctx, c.client)
}

// loginTokens passes updateLoginToken from the client to a QR sign in.
type loginTokens struct {
	mux sync.Mutex
	h   tg.LoginTokenHandler
}

// OnLoginToken implements the dispatcher interface of qrlogin.OnLoginToken.
func (t *loginTokens) OnLoginToken(h tg.LoginTokenHandler) {
	t.mux.Lock()
	t.h = h
	t.mux.Unlock()
}

func (t *loginTokens) reset() { t.OnLoginToken(nil) }

// onUpdates calls the handler for every updateLoginToken in u.
func (t *loginTokens) onUpdates(ctx context.Context, u tg.UpdatesClass) {
	t.mux.Lock()
	h := t.h
	t.mux.Unlock()
	if h == nil {
		return
	}
	var list []tg.UpdateClass
	switch u := u.(type) {
	case *tg.UpdateShort:
		list = []tg.UpdateClass{u.Update}
	case *tg.Updates:
		list = u.Updates
	case *tg.UpdatesCombined:
		list = u.Updates
	}
	for _, upd := range list {
		if lt, ok := upd.(*tg.UpdateLoginToken); ok {
			_ = h(ctx, tg.Entities{}, lt)
		}
	}
}
