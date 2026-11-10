package telegram

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tgerr"
)

// ConfigError is returned by Client.Run when the server rejects the
// connection setup (initConnection) with an error that reconnecting cannot
// fix, e.g. API_ID_INVALID, API_ID_PUBLISHED_FLOOD or
// CONNECTION_API_ID_INVALID: the client is misconfigured.
//
// Use errors.As to match it.
type ConfigError struct {
	// Err is the RPC error returned to initConnection.
	Err *tgerr.Error
}

// Error implements error.
func (e *ConfigError) Error() string {
	return "connection setup rejected: " + e.Err.Error()
}

// Unwrap returns the RPC error.
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// asConfigError returns ConfigError if err is a permanent error of
// initConnection.
//
// These are 400 and 406 errors: the same request with the same parameters
// fails again. Official clients send initConnection as a prefix of the
// first real request, so such an error fails that request instead of
// reconnecting.
//
// CONNECTION_NOT_INITED and CONNECTION_LAYER_INVALID are excluded: TDLib and
// Telegram Desktop handle them by sending initConnection again. Errors of
// IsAuthLost (e.g. AUTH_KEY_DUPLICATED) keep their own handling.
func asConfigError(err error) (*ConfigError, bool) {
	var initErr *manager.InitError
	if !errors.As(err, &initErr) {
		return nil, false
	}
	rpcErr, ok := tgerr.As(initErr.Err)
	if !ok {
		return nil, false
	}
	if rpcErr.Code != 400 && rpcErr.Code != 406 {
		return nil, false
	}
	if rpcErr.IsOneOf("CONNECTION_NOT_INITED", "CONNECTION_LAYER_INVALID") || IsAuthLost(rpcErr) {
		return nil, false
	}
	return &ConfigError{Err: rpcErr}, true
}
