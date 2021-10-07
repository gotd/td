package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/tg"
)

// NewFlow initializes new authentication flow.
func NewFlow(auth UserAuthenticator, opt SendCodeOptions) Flow {
	return Flow{
		Auth:    auth,
		Options: opt,
	}
}

// Flow simplifies boilerplate for authentication flow.
type Flow struct {
	Auth    UserAuthenticator
	Options SendCodeOptions
}

// Run starts authentication flow on client.
func (f Flow) Run(ctx context.Context, client FlowClient) error {
	if f.Auth == nil {
		return xerrors.New("no UserAuthenticator provided")
	}

	phone, err := f.Auth.Phone(ctx)
	if err != nil {
		return xerrors.Errorf("get phone: %w", err)
	}

	sentCode, err := client.SendCode(ctx, phone, f.Options)
	if err != nil {
		return xerrors.Errorf("send code: %w", err)
	}
	hash := sentCode.PhoneCodeHash

	code, err := f.Auth.Code(ctx, sentCode)
	if err != nil {
		return xerrors.Errorf("get code: %w", err)
	}

	_, signInErr := client.SignIn(ctx, phone, code, hash)

	if xerrors.Is(signInErr, ErrPasswordAuthNeeded) {
		password, err := f.Auth.Password(ctx)
		if err != nil {
			return xerrors.Errorf("get password: %w", err)
		}
		if _, err := client.Password(ctx, password); err != nil {
			return xerrors.Errorf("sign in with password: %w", err)
		}
		return nil
	}

	var signUpRequired *SignUpRequired
	if xerrors.As(signInErr, &signUpRequired) {
		if err := f.Auth.AcceptTermsOfService(ctx, signUpRequired.TermsOfService); err != nil {
			return xerrors.Errorf("confirm TOS: %w", err)
		}
		info, err := f.Auth.SignUp(ctx)
		if err != nil {
			return xerrors.Errorf("sign up info not provided: %w", err)
		}
		if _, err := client.SignUp(ctx, SignUp{
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

// FlowClient abstracts telegram client for Flow.
type FlowClient interface {
	SignIn(ctx context.Context, phone, code, codeHash string) (*tg.AuthAuthorization, error)
	SendCode(ctx context.Context, phone string, options SendCodeOptions) (*tg.AuthSentCode, error)
	Password(ctx context.Context, password string) (*tg.AuthAuthorization, error)
	SignUp(ctx context.Context, s SignUp) (*tg.AuthAuthorization, error)
}

// CodeAuthenticator asks user for received authentication code.
type CodeAuthenticator interface {
	Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error)
}

// CodeAuthenticatorFunc is functional wrapper for CodeAuthenticator.
type CodeAuthenticatorFunc func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error)

// Code implements CodeAuthenticator interface.
func (c CodeAuthenticatorFunc) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	return c(ctx, sentCode)
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

// Constant creates UserAuthenticator with constant phone and password.
func Constant(phone, password string, code CodeAuthenticator) UserAuthenticator {
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

// Env creates UserAuthenticator which gets phone and password from environment variables.
func Env(prefix string, code CodeAuthenticator) UserAuthenticator {
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

// CodeOnly creates UserAuthenticator with constant phone and no password.
func CodeOnly(phone string, code CodeAuthenticator) UserAuthenticator {
	return codeOnlyAuth{
		phone:             phone,
		CodeAuthenticator: code,
	}
}

type testAuth struct {
	dc    int
	phone string
}

func (t testAuth) Phone(ctx context.Context) (string, error)    { return t.phone, nil }
func (t testAuth) Password(ctx context.Context) (string, error) { return "", ErrPasswordNotProvided }
func (t testAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	type notFlashing interface {
		GetLength() int
	}

	typ, ok := sentCode.Type.(notFlashing)
	if !ok {
		return "", xerrors.Errorf("unexpected type: %T", sentCode.Type)
	}

	return strings.Repeat(strconv.Itoa(t.dc), typ.GetLength()), nil
}

func (t testAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (t testAuth) SignUp(ctx context.Context) (UserInfo, error) {
	return UserInfo{
		FirstName: "Test",
		LastName:  "User",
	}, nil
}

// Test returns UserAuthenticator that authenticates via testing credentials.
//
// Can be used only with testing server. Will perform sign up if test user is
// not registered.
func Test(randReader io.Reader, dc int) UserAuthenticator {
	// 99966XYYYY, X = dc_id, Y = random numbers, code = X repeat 6.
	// The n value is from 0000 to 9999.
	n, err := crypto.RandInt64n(randReader, 1000)
	if err != nil {
		panic(err)
	}
	phone := fmt.Sprintf("99966%d%04d", dc, n)

	return TestUser(phone, dc)
}

// TestUser returns UserAuthenticator that authenticates via testing credentials.
// Uses given phone to sign in/sign up.
//
// Can be used only with testing server. Will perform sign up if test user is
// not registered.
func TestUser(phone string, dc int) UserAuthenticator {
	return testAuth{
		dc:    dc,
		phone: phone,
	}
}
