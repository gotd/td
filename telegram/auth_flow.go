package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/tg"
)

// NewAuth initializes new authentication flow.
func NewAuth(auth UserAuthenticator, opt SendCodeOptions) AuthFlow {
	return AuthFlow{
		Auth:    auth,
		Options: opt,
	}
}

// AuthFlow simplifies boilerplate for authentication flow.
type AuthFlow struct {
	Auth    UserAuthenticator
	Options SendCodeOptions
}

// Run starts authentication flow on client.
func (f AuthFlow) Run(ctx context.Context, client AuthFlowClient) error {
	if f.Auth == nil {
		return xerrors.New("no UserAuthenticator provided")
	}
	phone, err := f.Auth.Phone(ctx)
	if err != nil {
		return xerrors.Errorf("get phone: %w", err)
	}
	hash, err := client.AuthSendCode(ctx, phone, f.Options)
	if err != nil {
		return xerrors.Errorf("send code: %w", err)
	}
	code, err := f.Auth.Code(ctx)
	if err != nil {
		return xerrors.Errorf("get code: %w", err)
	}

	signInErr := client.AuthSignIn(ctx, phone, code, hash)

	if errors.Is(signInErr, ErrPasswordAuthNeeded) {
		password, err := f.Auth.Password(ctx)
		if err != nil {
			return xerrors.Errorf("get password: %w", err)
		}
		if err := client.AuthPassword(ctx, password); err != nil {
			return xerrors.Errorf("sign in with password: %w", err)
		}

		return nil
	}

	var signUpRequired *SignUpRequired
	if errors.As(signInErr, &signUpRequired) {
		if err := f.Auth.AcceptTermsOfService(ctx, signUpRequired.TermsOfService); err != nil {
			return xerrors.Errorf("confirm TOS: %w", err)
		}
		info, err := f.Auth.SignUp(ctx)
		if err != nil {
			return xerrors.Errorf("sign up info not provided: %w", err)
		}
		if err := client.AuthSignUp(ctx, SignUp{
			PhoneNumber:   phone,
			PhoneCodeHash: hash,
			FirstName:     info.FirstName,
			LastName:      info.LastName,
		}); err != nil {
			return xerrors.Errorf("sign up: %w", err)
		}

		return nil
	}

	if signInErr != nil {
		return xerrors.Errorf("sign in: %w", signInErr)
	}

	return nil
}

// AuthFlowClient abstracts telegram client for AuthFlow.
type AuthFlowClient interface {
	AuthSignIn(ctx context.Context, phone, code, codeHash string) error
	AuthSendCode(ctx context.Context, phone string, options SendCodeOptions) (codeHash string, err error)
	AuthPassword(ctx context.Context, password string) error
	AuthSignUp(ctx context.Context, s SignUp) error
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

// UserInfo represents user info required for sign up.
type UserInfo struct {
	FirstName string
	LastName  string
}

// UserAuthenticator asks user for phone, password and received authentication code.
type UserAuthenticator interface {
	Phone(ctx context.Context) (string, error)
	Password(ctx context.Context) (string, error)
	AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error
	SignUp(ctx context.Context) (UserInfo, error)
	CodeAuthenticator
}

type noSignUp struct{}

func (c noSignUp) SignUp(ctx context.Context) (UserInfo, error) {
	return UserInfo{}, xerrors.New("not implemented")
}

func (c noSignUp) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &SignUpRequired{TermsOfService: tos}
}

type constantAuth struct {
	phone, password string
	CodeAuthenticator
	noSignUp
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

type envAuth struct {
	prefix string
	CodeAuthenticator
	noSignUp
}

func (e envAuth) lookup(k string) (string, error) {
	env := e.prefix + k
	v, ok := os.LookupEnv(env)
	if !ok {
		return "", xerrors.Errorf("environment variable %q not set", env)
	}
	return v, nil
}

func (e envAuth) Phone(ctx context.Context) (string, error) {
	return e.lookup("PHONE")
}

func (e envAuth) Password(ctx context.Context) (string, error) {
	p, err := e.lookup("PASSWORD")
	if err != nil {
		return "", ErrPasswordNotProvided
	}
	return p, nil
}

// EnvAuth creates UserAuthenticator which gets phone and password from environment variables.
func EnvAuth(prefix string, code CodeAuthenticator) UserAuthenticator {
	return envAuth{
		prefix:            prefix,
		CodeAuthenticator: code,
		noSignUp:          noSignUp{},
	}
}

// ErrPasswordNotProvided means that password requested by Telegram,
// but not provided by user.
var ErrPasswordNotProvided = errors.New("password requested but not provided")

type codeOnlyAuth struct {
	phone string
	CodeAuthenticator
	noSignUp
}

func (c codeOnlyAuth) Phone(ctx context.Context) (string, error) {
	return c.phone, nil
}

func (c codeOnlyAuth) Password(ctx context.Context) (string, error) {
	return "", ErrPasswordNotProvided
}

// CodeOnlyAuth creates UserAuthenticator with constant phone and no password.
func CodeOnlyAuth(phone string, code CodeAuthenticator) UserAuthenticator {
	return codeOnlyAuth{
		phone:             phone,
		CodeAuthenticator: code,
	}
}

type testAuth struct {
	code  string
	phone string
}

func (t testAuth) Phone(ctx context.Context) (string, error)    { return t.phone, nil }
func (t testAuth) Password(ctx context.Context) (string, error) { return "", ErrPasswordNotProvided }
func (t testAuth) Code(ctx context.Context) (string, error)     { return t.code, nil }

func (t testAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (t testAuth) SignUp(ctx context.Context) (UserInfo, error) {
	return UserInfo{
		FirstName: "Test",
		LastName:  "User",
	}, nil
}

// TestAuth returns UserAuthenticator that authenticates via testing credentials.
//
// Can be used only with testing server. Will perform sign up if test user is
// not registered.
func TestAuth(randReader io.Reader, dc int) UserAuthenticator {
	// 99966XYYYY, X = dc_id, Y = random numbers, code = X repeat 5.
	// The n value is from 0000 to 9999.
	n, err := crypto.RandInt64n(randReader, 1000)
	if err != nil {
		panic(err)
	}
	code := strings.Repeat(strconv.Itoa(dc), 5)
	phone := fmt.Sprintf("99966%d%04d", dc, n)

	return testAuth{
		code:  code,
		phone: phone,
	}
}
