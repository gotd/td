package telegram

import (
	"context"
	"sync/atomic"

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
func IsAuthLost(err error) bool {
	return tgerr.Is(err,
		"AUTH_KEY_UNREGISTERED",
		"AUTH_KEY_INVALID",
		"SESSION_REVOKED",
		"SESSION_EXPIRED",
		"USER_DEACTIVATED",
		"USER_DEACTIVATED_BAN",
		"AUTH_KEY_DUPLICATED",
	)
}

// authState tracks whether the client is authorized in current Run.
type authState struct {
	authorized atomic.Bool
	reported   atomic.Bool
}

// observeAuth tracks authorization by request results and reports lost
// authorization of the primary session.
func (c *Client) observeAuth(ctx context.Context, input bin.Encoder, err error) {
	if err == nil {
		switch input.(type) {
		case *tg.AuthSignInRequest,
			*tg.AuthSignUpRequest,
			*tg.AuthCheckPasswordRequest,
			*tg.AuthImportBotAuthorizationRequest,
			*tg.AuthImportLoginTokenRequest,
			*tg.AuthImportAuthorizationRequest,
			*tg.AuthImportWebTokenAuthorizationRequest,
			*tg.AuthRecoverPasswordRequest:
			c.markAuthorized()
		case *tg.AuthLogOutRequest:
			c.auth.authorized.Store(false)
		}
		return
	}
	if c.onAuthLost == nil || !c.auth.authorized.Load() || !IsAuthLost(err) {
		return
	}
	if !c.auth.reported.CompareAndSwap(false, true) {
		return
	}
	c.auth.authorized.Store(false)

	reason := c.onAuthLost(ctx, err)
	if reason == nil {
		// Allow reporting again after next authorization.
		c.auth.reported.Store(false)
		return
	}
	select {
	case c.authLost <- &AuthLostError{Err: err, Reason: reason}:
	default:
	}
}

func (c *Client) markAuthorized() {
	c.auth.authorized.Store(true)
	c.auth.reported.Store(false)
}
