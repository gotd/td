package telegram

import (
	"context"
	"strings"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// AuthLostError is returned by Client.Run when the authorization of a client
// is lost (e.g. the session was terminated from another device) and
// Options.OnAuthLost returned an error.
//
// Use errors.As to match it.
type AuthLostError struct {
	// Err is the RPC error that revealed the lost authorization, e.g.
	// AUTH_KEY_UNREGISTERED or SESSION_REVOKED.
	Err error
	// Reason is the error returned by Options.OnAuthLost.
	Reason error
}

// Error implements error.
func (e *AuthLostError) Error() string {
	return "authorization lost: " + e.Err.Error()
}

// Unwrap returns both underlying errors.
func (e *AuthLostError) Unwrap() []error {
	if e.Reason == nil || e.Reason == e.Err {
		return []error{e.Err}
	}
	return []error{e.Err, e.Reason}
}

// IsAuthLost reports whether err means that the authorization of the session
// is lost: the auth key is unregistered, the session is revoked or expired,
// the user is deactivated, or the key is used by another connection
// (AUTH_KEY_DUPLICATED).
//
// SESSION_PASSWORD_NEEDED and other "not authorized yet" errors are not
// included.
//
// Errors made from AUTH_KEY_DROP_* service notifications (see
// AuthKeyDropError) are included.
func IsAuthLost(err error) bool {
	if tgerr.Is(err,
		"AUTH_KEY_UNREGISTERED",
		"AUTH_KEY_INVALID",
		"SESSION_REVOKED",
		"SESSION_EXPIRED",
		"USER_DEACTIVATED",
		"USER_DEACTIVATED_BAN",
		"AUTH_KEY_DUPLICATED",
	) {
		return true
	}
	rpcErr, ok := tgerr.As(err)
	return ok && strings.HasPrefix(rpcErr.Type, authKeyDropPrefix)
}

// authState tracks whether the client is authorized in current Run.
type authState struct {
	authorized atomic.Bool
	reported   atomic.Bool
	// stopped is set when OnAuthLost stopped Run.
	stopped atomic.Bool
	// done is closed when the last OnAuthLost call returns, nil if there
	// was none. Calls are chained, so it is closed after all of them.
	done atomic.Pointer[chan struct{}]
}

// waitAuthLost waits for the running OnAuthLost call, if any.
func (c *Client) waitAuthLost(ctx context.Context) error {
	done := c.auth.done.Load()
	if done == nil {
		return nil
	}
	select {
	case <-*done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// queryWrapper is a request wrapping another one, e.g. invokeWithLayer or
// invokeWithoutUpdates.
type queryWrapper interface {
	GetQuery() bin.Object
}

// unwrapQuery returns the innermost request of invokeWith* and
// initConnection wrappers.
func unwrapQuery(input bin.Encoder) bin.Encoder {
	for i := 0; i < 8; i++ {
		w, ok := input.(queryWrapper)
		if !ok {
			return input
		}
		q := w.GetQuery()
		if q == nil {
			return input
		}
		input = q
	}
	return input
}

// authEffect is how a successful request changes authorization.
type authEffect int

const (
	// authNone means that the request tells nothing about authorization.
	authNone authEffect = iota
	// authAuthorized means that the client is authorized.
	authAuthorized
	// authLoggedOut means that the client logged out.
	authLoggedOut
)

// noAuthRequests are requests outside of auth.*, help.* and langpack.*
// that succeed before sign in is complete (2FA, login email setup).
var noAuthRequests = map[string]struct{}{ // nolint:gochecknoglobals
	"account.getPassword":         {},
	"account.deleteAccount":       {},
	"account.sendVerifyEmailCode": {},
	"account.verifyEmail":         {},
}

// authEffectOf reports how the successful result of input changes
// authorization.
func authEffectOf(input bin.Encoder, output bin.Decoder) authEffect {
	input = unwrapQuery(input)
	switch input.(type) {
	case *tg.AuthLogOutRequest:
		return authLoggedOut
	case *tg.AuthSignInRequest,
		*tg.AuthSignUpRequest,
		*tg.AuthCheckPasswordRequest,
		*tg.AuthImportBotAuthorizationRequest,
		*tg.AuthImportAuthorizationRequest,
		*tg.AuthImportWebTokenAuthorizationRequest,
		*tg.AuthRecoverPasswordRequest,
		*tg.AuthImportLoginTokenRequest,
		*tg.AuthExportLoginTokenRequest:
		if isAuthorization(output) {
			return authAuthorized
		}
		return authNone
	}
	named, ok := input.(interface{ TypeName() string })
	if !ok {
		return authNone
	}
	name := named.TypeName()
	if _, ok := noAuthRequests[name]; ok {
		return authNone
	}
	for _, prefix := range []string{"auth.", "help.", "langpack."} {
		if strings.HasPrefix(name, prefix) {
			return authNone
		}
	}
	// Any other successful request requires authorization.
	return authAuthorized
}

// isAuthorization reports whether output is a complete authorization, not
// auth.authorizationSignUpRequired or a login token to wait for.
func isAuthorization(output bin.Decoder) bool {
	switch v := output.(type) {
	case *tg.AuthAuthorizationBox:
		_, ok := v.Authorization.(*tg.AuthAuthorization)
		return ok
	case *tg.AuthAuthorization:
		return true
	case *tg.AuthLoginTokenBox:
		success, ok := v.LoginToken.(*tg.AuthLoginTokenSuccess)
		if !ok {
			return false
		}
		_, ok = success.Authorization.(*tg.AuthAuthorization)
		return ok
	case *tg.AuthLoginTokenSuccess:
		_, ok := v.Authorization.(*tg.AuthAuthorization)
		return ok
	default:
		return false
	}
}

// observeAuth tracks authorization by request results and reports lost
// authorization of the primary session.
func (c *Client) observeAuth(input bin.Encoder, output bin.Decoder, err error) {
	if err == nil {
		switch authEffectOf(input, output) {
		case authAuthorized:
			c.markAuthorized()
		case authLoggedOut:
			c.auth.authorized.Store(false)
		}
		return
	}
	c.observeRestriction(input, err)
	if !IsAuthLost(err) {
		return
	}
	c.authLostDetected(err)
}

// authLostDetected reports lost authorization revealed by err, if the client
// is authorized.
func (c *Client) authLostDetected(err error) {
	if !c.auth.authorized.Load() {
		return
	}
	if c.ctx != nil && c.ctx.Err() != nil {
		// Run is stopping: it no longer waits for reports.
		return
	}
	if !c.auth.reported.CompareAndSwap(false, true) {
		return
	}
	c.auth.authorized.Store(false)

	// Reported once per authorized -> lost transition, with or without
	// OnAuthLost. The reason is one of IsAuthLost error types.
	reason := "unknown"
	if rpcErr, ok := tgerr.As(err); ok {
		reason = rpcErr.Type
	}
	kind := AuthLostKindOf(err)
	bg := context.Background()
	c.meters().authLost.Add(bg, 1, metric.WithAttributes(
		attribute.String("reason", reason),
		attribute.String("kind", kind.String()),
	))
	fields := []log.Attr{
		log.String("reason", reason),
		log.String("kind", kind.String()),
		log.Bool("callback", c.onAuthLost != nil),
	}
	switch kind {
	case AuthLostBanned:
		// Same advice as TDLib gives.
		c.log.Error(bg, "Authorization lost: the account is banned by Telegram. "+
			"If this is a mistake, log in from an official mobile app and follow "+
			"its instructions to recover the account", fields...)
	case AuthLostDeleted:
		c.log.Error(bg, "Authorization lost: the account is deleted", fields...)
	default:
		c.log.Warn(bg, "Authorization lost", fields...)
	}
	if c.onAuthLost == nil {
		return
	}

	// Call OnAuthLost on its own goroutine, so that the failing request
	// returns at once and the callback can make requests. Run and the
	// connect-time Self wait for it via done.
	//
	// A report may start while the previous one still runs, if the callback
	// signed in again and authorization was lost again: reports are chained,
	// so they run in order and done of the last one covers all.
	done := make(chan struct{})
	prev := c.auth.done.Swap(&done)
	go func() {
		defer close(done)
		if prev != nil {
			<-*prev
		}
		c.reportAuthLost(err)
	}()
}

// reportAuthLost calls OnAuthLost and stops Run if it returns an error.
func (c *Client) reportAuthLost(err error) {
	// The failing request's context may be about to expire: the callback
	// gets the context of Run, so that it can sign in again.
	ctx := c.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	reason := c.onAuthLost(ctx, err)
	if reason == nil {
		// Allow reporting again after next authorization.
		c.auth.reported.Store(false)
		return
	}
	c.auth.stopped.Store(true)
	select {
	case c.authLost <- &AuthLostError{Err: err, Reason: reason}:
	default:
	}
}

func (c *Client) markAuthorized() {
	if c.auth.authorized.Load() {
		return
	}
	c.auth.authorized.Store(true)
	c.auth.reported.Store(false)
}
