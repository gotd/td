package telegram

import "context"

// CodeAuthenticator asks user for received authentication code.
type CodeAuthenticator interface {
	Code(ctx context.Context) (string, error)
}

// CodeAuthenticatorFunc is functional wrapper for CodeAuthenticator.
type CodeAuthenticatorFunc func(ctx context.Context) (string, error)

// Code implements CodeAuthenticator interface.
func (c CodeAuthenticatorFunc) Code(ctx context.Context) (string, error) {
	return c(ctx)
}

// UserAuthenticator asks user for phone, password and received authentication code.
type UserAuthenticator interface {
	Phone(ctx context.Context) (string, error)
	Password(ctx context.Context) (string, error)
	CodeAuthenticator
}

type constantAuth struct {
	phone, password string
	CodeAuthenticator
}

func (c constantAuth) Phone(ctx context.Context) (string, error) {
	return c.phone, nil
}

func (c constantAuth) Password(ctx context.Context) (string, error) {
	return c.password, nil
}

// ConstantAuth creates UserAuthenticator with constant phone and password.
func ConstantAuth(phone, password string, code CodeAuthenticator) UserAuthenticator {
	return constantAuth{
		phone:             phone,
		password:          password,
		CodeAuthenticator: code,
	}
}
