package telegram

import (
	"context"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// ErrPasswordInvalid means that password provided to AuthPassword is invalid.
//
// Note that telegram does not trim whitespace characters by default, check
// that provided password is expected and clean whitespaces if needed.
// You can use strings.TrimSpace(password) for this.
//
// Deprecated: use auth package.
var ErrPasswordInvalid = auth.ErrPasswordInvalid

// AuthPassword performs login via secure remote password (aka 2FA).
//
// Method can be called after AuthSignIn to provide password if requested.
//
// Deprecated: use auth package.
func (c *Client) AuthPassword(ctx context.Context, password string) (*tg.AuthAuthorization, error) {
	return c.Auth().Password(ctx, password)
}

// SendCodeOptions defines how to send auth code to user.
//
// Deprecated: use auth package.
type SendCodeOptions = auth.SendCodeOptions

// AuthSendCode requests code for provided phone number, returning code hash
// and error if any. Use AuthFlow to reduce boilerplate.
//
// This method should be called first in user authentication flow.
//
// Deprecated: use auth package.
func (c *Client) AuthSendCode(ctx context.Context, phone string, options SendCodeOptions) (*tg.AuthSentCode, error) {
	return c.Auth().SendCode(ctx, phone, options)
}

// ErrPasswordAuthNeeded means that 2FA auth is required.
//
// Call Client.AuthPassword to provide 2FA password.
var ErrPasswordAuthNeeded = auth.ErrPasswordAuthNeeded

// AuthSignIn performs sign in with provided user phone, code and code hash.
//
// If ErrPasswordAuthNeeded is returned, call AuthPassword to provide 2FA
// password.
//
// To obtain codeHash, use AuthSendCode.
func (c *Client) AuthSignIn(ctx context.Context, phone, code, codeHash string) (*tg.AuthAuthorization, error) {
	return c.Auth().SignIn(ctx, phone, code, codeHash)
}

// AuthAcceptTOS accepts version of Terms Of Service.
func (c *Client) AuthAcceptTOS(ctx context.Context, id tg.DataJSON) error {
	return c.Auth().AcceptTOS(ctx, id)
}

// SignUp wraps parameters for AuthSignUp.
type SignUp = auth.SignUp

// AuthSignUp registers a validated phone number in the system.
//
// To obtain codeHash, use AuthSendCode.
// Use AuthFlow helper to handle authentication flow.
func (c *Client) AuthSignUp(ctx context.Context, s SignUp) (*tg.AuthAuthorization, error) {
	return c.Auth().SignUp(ctx, s)
}
