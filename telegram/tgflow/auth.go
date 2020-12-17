package tgflow

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
)

// NewAuth initializes new authentication flow.
func NewAuth(auth UserAuthenticator, opt telegram.SendCodeOptions) Auth {
	return Auth{
		Auth:    auth,
		Options: opt,
	}
}

// Auth simplifies boilerplate for authentication flow.
type Auth struct {
	Auth    UserAuthenticator
	Options telegram.SendCodeOptions
}

// Run starts authentication flow on client.
func (f Auth) Run(ctx context.Context, client AuthFlowClient) error {
	if f.Auth == nil {
		return xerrors.New("no UserAuthenticator provided")
	}
	phone, err := f.Auth.Phone(ctx)
	if err != nil {
		return xerrors.Errorf("failed to get phone: %w", err)
	}
	hash, err := client.AuthSendCode(ctx, phone, f.Options)
	if err != nil {
		return xerrors.Errorf("failed to send code: %w", err)
	}
	code, err := f.Auth.Code(ctx)
	if err != nil {
		return xerrors.Errorf("failed to get code: %w", err)
	}

	signInErr := client.AuthSignIn(ctx, phone, code, hash)
	if errors.Is(signInErr, telegram.ErrPasswordAuthNeeded) {
		password, err := f.Auth.Password(ctx)
		if err != nil {
			return xerrors.Errorf("failed to get password: %w", err)
		}

		if err := client.AuthPassword(ctx, password); err != nil {
			return xerrors.Errorf("failed to sign in with password: %w", err)
		}

		return nil
	}
	if signInErr != nil {
		return xerrors.Errorf("failed to sign in: %w", signInErr)
	}

	return nil
}

// AuthFlowClient abstracts telegram client for Auth.
type AuthFlowClient interface {
	AuthSignIn(ctx context.Context, phone, code, codeHash string) error
	AuthSendCode(ctx context.Context, phone string, options telegram.SendCodeOptions) (codeHash string, err error)
	AuthPassword(ctx context.Context, password string) error
}

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

// ErrPasswordRequestButNotProvided means that password requested by Telegram,
// but not provided by user.
var ErrPasswordRequestButNotProvided = errors.New("password requested but not provided")

type codeOnlyAuth struct {
	phone string
	CodeAuthenticator
}

func (c codeOnlyAuth) Phone(ctx context.Context) (string, error) {
	return c.phone, nil
}

func (c codeOnlyAuth) Password(ctx context.Context) (string, error) {
	return "", ErrPasswordRequestButNotProvided
}

// CodeOnlyAuth creates UserAuthenticator with constant phone and no password.
func CodeOnlyAuth(phone string, code CodeAuthenticator) UserAuthenticator {
	return codeOnlyAuth{
		phone:             phone,
		CodeAuthenticator: code,
	}
}

// TestAuth returns UserAuthenticator that authenticates via testing credentials.
//
// Can be used only with testing server.
func TestAuth(randReader io.Reader, dc int) UserAuthenticator {
	// 99966XYYYY, X = dc_id, Y = random numbers, code = X repeat 5.
	// The n value is from 0000 to 9999.
	n, err := rand.Int(randReader, big.NewInt(1000))
	if err != nil {
		panic(err)
	}
	code := strings.Repeat(strconv.Itoa(dc), 5)
	phone := fmt.Sprintf("99966%d%04d", dc, n)

	return CodeOnlyAuth(phone, CodeAuthenticatorFunc(func(ctx context.Context) (string, error) {
		return code, nil
	}))
}
