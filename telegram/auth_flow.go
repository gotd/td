package telegram

import (
	"io"

	"github.com/gotd/td/telegram/auth"
)

// NewAuth initializes new authentication flow.
//
// Deprecated: use auth.NewFlow.
func NewAuth(u UserAuthenticator, opt SendCodeOptions) AuthFlow {
	return AuthFlow{
		Auth:    u,
		Options: opt,
	}
}

// AuthFlow simplifies boilerplate for authentication flow.
//
// Deprecated: use auth.Flow.
type AuthFlow = auth.Flow

// AuthFlowClient abstracts telegram client for AuthFlow.
//
// Deprecated: use auth.FlowClient.
type AuthFlowClient = auth.FlowClient

// CodeAuthenticator asks user for received authentication code.
//
// Deprecated: use auth.CodeAuthenticator.
type CodeAuthenticator = auth.CodeAuthenticator

// CodeAuthenticatorFunc is functional wrapper for CodeAuthenticator.
type CodeAuthenticatorFunc = auth.CodeAuthenticatorFunc

// UserInfo represents user info required for sign up.
//
// Deprecated: use auth package.
type UserInfo = auth.UserInfo

// UserAuthenticator asks user for phone, password and received authentication code.
//
// Deprecated: use auth package.
type UserAuthenticator = auth.UserAuthenticator

// ConstantAuth creates UserAuthenticator with constant phone and password.
//
// Deprecated: use auth package.
func ConstantAuth(phone, password string, code CodeAuthenticator) UserAuthenticator {
	return auth.Constant(phone, password, code)
}

// EnvAuth creates UserAuthenticator which gets phone and password from environment variables.
//
// Deprecated: use auth package.
func EnvAuth(prefix string, code CodeAuthenticator) UserAuthenticator {
	return auth.Env(prefix, code)
}

// ErrPasswordNotProvided means that password requested by Telegram,
// but not provided by user.
//
// Deprecated: use auth package.
var ErrPasswordNotProvided = auth.ErrPasswordNotProvided

// CodeOnlyAuth creates UserAuthenticator with constant phone and no password.
//
// Deprecated: use auth package.
func CodeOnlyAuth(phone string, code CodeAuthenticator) UserAuthenticator {
	return auth.CodeOnly(phone, code)
}

// TestAuth returns UserAuthenticator that authenticates via testing credentials.
//
// Can be used only with testing server. Will perform sign up if test user is
// not registered.
//
// Deprecated: use auth package.
func TestAuth(randReader io.Reader, dc int) UserAuthenticator {
	return auth.Test(randReader, dc)
}
