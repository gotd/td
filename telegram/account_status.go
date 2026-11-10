package telegram

import (
	"context"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// AuthLostKind is why authorization is lost, see AuthLostKindOf.
type AuthLostKind int

const (
	// AuthLostUnknown is an error that is not one of IsAuthLost errors.
	AuthLostUnknown AuthLostKind = iota
	// AuthLostBanned means that the account is banned by Telegram
	// (USER_DEACTIVATED_BAN).
	AuthLostBanned
	// AuthLostDeleted means that the account is deleted (USER_DEACTIVATED).
	AuthLostDeleted
	// AuthLostRevoked means that the session was terminated, e.g. from
	// another device (SESSION_REVOKED), or the server asked to drop the key
	// (AUTH_KEY_DROP_* service notification).
	AuthLostRevoked
	// AuthLostUnregistered means that the server doesn't know the auth key
	// as an authorized one (AUTH_KEY_UNREGISTERED). It is what a terminated
	// session usually returns, but also what the keys of a deleted or banned
	// account return once they are dropped.
	AuthLostUnregistered
	// AuthLostExpired means that the session expired (SESSION_EXPIRED).
	AuthLostExpired
	// AuthLostKeyInvalid means that the auth key is invalid or used by
	// another connection (AUTH_KEY_INVALID, AUTH_KEY_DUPLICATED).
	AuthLostKeyInvalid
)

// String implements fmt.Stringer. Values are used in logs and metrics.
func (k AuthLostKind) String() string {
	switch k {
	case AuthLostBanned:
		return "banned"
	case AuthLostDeleted:
		return "deleted"
	case AuthLostRevoked:
		return "revoked"
	case AuthLostUnregistered:
		return "unregistered"
	case AuthLostExpired:
		return "expired"
	case AuthLostKeyInvalid:
		return "key_invalid"
	default:
		return "unknown"
	}
}

// AuthLostKindOf classifies an error of IsAuthLost.
func AuthLostKindOf(err error) AuthLostKind {
	rpcErr, ok := tgerr.As(err)
	if !ok {
		return AuthLostUnknown
	}
	switch rpcErr.Type {
	case "USER_DEACTIVATED_BAN":
		return AuthLostBanned
	case "USER_DEACTIVATED":
		return AuthLostDeleted
	case "SESSION_REVOKED":
		return AuthLostRevoked
	case "AUTH_KEY_UNREGISTERED":
		return AuthLostUnregistered
	case "SESSION_EXPIRED":
		return AuthLostExpired
	case "AUTH_KEY_INVALID", "AUTH_KEY_DUPLICATED":
		return AuthLostKeyInvalid
	}
	if strings.HasPrefix(rpcErr.Type, authKeyDropPrefix) {
		return AuthLostRevoked
	}
	return AuthLostUnknown
}

// IsAccountBanned reports whether err means that the account is banned by
// Telegram: USER_DEACTIVATED_BAN for a signed in account, PHONE_NUMBER_BANNED
// when signing in.
func IsAccountBanned(err error) bool {
	return tgerr.Is(err, "USER_DEACTIVATED_BAN", tg.ErrPhoneNumberBanned)
}

// Kind returns why authorization is lost.
func (e *AuthLostError) Kind() AuthLostKind {
	return AuthLostKindOf(e.Err)
}

// authKeyDropPrefix is the type prefix of service notifications asking to
// drop the auth key, e.g. after a key leak. Telegram Desktop logs out on
// them, mobile apps offer to.
const authKeyDropPrefix = "AUTH_KEY_DROP_"

// AuthKeyDropError returns the error that is reported via OnAuthLost when
// the server sends an AUTH_KEY_DROP_* service notification: a *tgerr.Error
// with code 401, the notification type as Type and its text as Message.
func AuthKeyDropError(u *tg.UpdateServiceNotification) error {
	return &tgerr.Error{
		Code:    401,
		Type:    u.Type,
		Message: u.Message,
	}
}

// handleServiceNotification logs service notifications and reports
// AUTH_KEY_DROP_* ones as lost authorization.
func (c *Client) handleServiceNotification(u *tg.UpdateServiceNotification) {
	ctx := context.Background()
	if strings.HasPrefix(u.Type, authKeyDropPrefix) {
		c.log.Warn(ctx, "Server asked to drop the auth key",
			log.String("type", u.Type),
			log.String("message", u.Message),
		)
		c.authLostDetected(AuthKeyDropError(u))
		return
	}
	if u.Popup {
		// Popups are how Telegram tells about restrictions, e.g. bans for
		// spam: worth seeing in logs.
		c.log.Info(ctx, "Service notification",
			log.String("type", u.Type),
			log.String("message", u.Message),
		)
		return
	}
	c.log.Debug(ctx, "Service notification", log.String("type", u.Type))
}

// FreezeState is the freeze state of the account, from app config.
//
// A frozen account is signed in and can read, but most methods fail with
// FROZEN_METHOD_INVALID until Until or a successful appeal.
//
// See https://core.telegram.org/api/config#freeze-since-date.
type FreezeState struct {
	// Since is when the account was frozen, zero if it is not.
	Since time.Time
	// Until is when the account will be deleted unless the appeal succeeds.
	Until time.Time
	// AppealURL is a t.me link to the bot to appeal with.
	AppealURL string
}

// Frozen reports whether the account is frozen.
func (s FreezeState) Frozen() bool {
	return !s.Since.IsZero()
}

// FreezeStateFrom returns the freeze state from help.getAppConfig result.
func FreezeStateFrom(cfg tg.JSONValueClass) FreezeState {
	var s FreezeState
	obj, ok := cfg.(*tg.JSONObject)
	if !ok {
		return s
	}
	unix := func(v tg.JSONValueClass) time.Time {
		n, ok := v.(*tg.JSONNumber)
		if !ok || n.Value <= 0 {
			return time.Time{}
		}
		return time.Unix(int64(n.Value), 0)
	}
	for _, kv := range obj.Value {
		switch kv.Key {
		case "freeze_since_date":
			s.Since = unix(kv.Value)
		case "freeze_until_date":
			s.Until = unix(kv.Value)
		case "freeze_appeal_url":
			if str, ok := kv.Value.(*tg.JSONString); ok {
				s.AppealURL = str.Value
			}
		}
	}
	return s
}

// FreezeState fetches app config and returns the freeze state of the account.
func (c *Client) FreezeState(ctx context.Context) (FreezeState, error) {
	cfg, err := c.API().HelpGetAppConfig(ctx, 0)
	if err != nil {
		return FreezeState{}, errors.Wrap(err, "get app config")
	}
	app, ok := cfg.(*tg.HelpAppConfig)
	if !ok {
		// helpAppConfigNotModified is not returned for hash 0.
		return FreezeState{}, errors.Errorf("unexpected app config %T", cfg)
	}
	return FreezeStateFrom(app.Config), nil
}

// IsFrozen reports whether err means that the request was refused because
// the account is frozen: FROZEN_METHOD_INVALID (code 420, not a flood
// wait) or FROZEN_PARTICIPANT_MISSING.
func IsFrozen(err error) bool {
	return tgerr.Is(err, tg.ErrFrozenMethodInvalid, tg.ErrFrozenParticipantMissing)
}

// freezeCheckInterval limits app config fetches on frozen errors, as in
// Telegram Desktop.
const freezeCheckInterval = time.Hour

// observeRestriction logs and counts errors that tell about restrictions
// of the account, and checks the freeze state on frozen errors.
func (c *Client) observeRestriction(input bin.Encoder, err error) {
	rpcErr, ok := tgerr.As(err)
	if !ok {
		return
	}
	frozen := IsFrozen(err)
	if !frozen && !rpcErr.IsType(tg.ErrPhoneNumberBanned) {
		return
	}
	ctx := context.Background()
	c.meters().restricted.Add(ctx, 1, metric.WithAttributes(
		attribute.String("reason", rpcErr.Type),
	))
	method := methodName(input)
	if !frozen {
		c.log.Error(ctx, "Phone number is banned by Telegram. "+
			"If this is a mistake, log in from an official mobile app and follow "+
			"its instructions to recover the account",
			log.String("method", method),
		)
		return
	}
	c.log.Debug(ctx, "Request refused: account is frozen",
		log.String("reason", rpcErr.Type),
		log.String("method", method),
	)

	now := c.clock.Now()
	c.freezeMux.Lock()
	if !c.freezeChecked.IsZero() && now.Sub(c.freezeChecked) < freezeCheckInterval {
		c.freezeMux.Unlock()
		return
	}
	c.freezeChecked = now
	c.freezeMux.Unlock()

	runCtx := c.ctx
	if runCtx == nil || runCtx.Err() != nil {
		return
	}
	// Not in the failing request's goroutine: app config is another request.
	go c.checkFreeze(runCtx, method)
}

// checkFreeze fetches the freeze state, logs it and calls OnFrozen.
func (c *Client) checkFreeze(ctx context.Context, method string) {
	state, err := c.FreezeState(ctx)
	if err != nil {
		if ctx.Err() == nil {
			c.log.Warn(ctx, "Failed to check freeze state", log.Error(err))
		}
		return
	}
	if !state.Frozen() {
		c.log.Warn(ctx, "Frozen error, but app config has no freeze state",
			log.String("method", method),
		)
		return
	}
	c.log.Warn(ctx, "Account is frozen",
		log.Time("since", state.Since),
		log.Time("until", state.Until),
		log.String("appeal_url", state.AppealURL),
		log.String("method", method),
	)
	if c.onFrozen != nil {
		c.onFrozen(ctx, state)
	}
}
