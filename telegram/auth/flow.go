package auth

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/tg"
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

func (f Flow) handleSignUp(ctx context.Context, client FlowClient, phone, hash string, s *SignUpRequired) error {
	if err := f.Auth.AcceptTermsOfService(ctx, s.TermsOfService); err != nil {
		return errors.Wrap(err, "confirm TOS")
	}
	info, err := f.Auth.SignUp(ctx)
	if err != nil {
		return errors.Wrap(err, "sign up info not provided")
	}
	if _, err := client.SignUp(ctx, SignUp{
		PhoneNumber:   phone,
		PhoneCodeHash: hash,
		FirstName:     info.FirstName,
		LastName:      info.LastName,
	}); err != nil {
		return errors.Wrap(err, "sign up")
	}
	return nil
}

// Run starts authentication flow on client.
func (f Flow) Run(ctx context.Context, client FlowClient) error {
	if f.Auth == nil {
		return errors.New("no UserAuthenticator provided")
	}

	phone, err := f.Auth.Phone(ctx)
	if err != nil {
		return errors.Wrap(err, "get phone")
	}

	sentCode, err := client.SendCode(ctx, phone, f.Options)
	if err != nil {
		return errors.Wrap(err, "send code")
	}
	switch s := sentCode.(type) {
	case *tg.AuthSentCode:
		hash := s.PhoneCodeHash
		code, err := f.Auth.Code(ctx, s)
		if err != nil {
			return errors.Wrap(err, "get code")
		}

		_, signInErr := client.SignIn(ctx, phone, code, hash)
		if errors.Is(signInErr, ErrPasswordAuthNeeded) {
			password, err := f.Auth.Password(ctx)
			if err != nil {
				return errors.Wrap(err, "get password")
			}
			if _, err := client.Password(ctx, password); err != nil {
				return errors.Wrap(err, "sign in with password")
			}
			return nil
		}
		var signUpRequired *SignUpRequired
		if errors.As(signInErr, &signUpRequired) {
			return f.handleSignUp(ctx, client, phone, hash, signUpRequired)
		}

		if signInErr != nil {
			return errors.Wrap(signInErr, "sign in")
		}

		return nil
	case *tg.AuthSentCodeSuccess:
		switch a := s.Authorization.(type) {
		case *tg.AuthAuthorization:
			// Looks that we are already authorized.
			return nil
		case *tg.AuthAuthorizationSignUpRequired:
			if err := f.handleSignUp(ctx, client, phone, "", &SignUpRequired{
				TermsOfService: a.TermsOfService,
			}); err != nil {
				// TODO: not sure that blank hash will work here
				return errors.Wrap(err, "sign up after auth sent code success")
			}
			return nil
		default:
			return errors.Errorf("unexpected authorization type: %T", a)
		}
	default:
		return errors.Errorf("unexpected sent code type: %T", sentCode)
	}
}

// FlowClient abstracts telegram client for Flow.
type FlowClient interface {
	SignIn(ctx context.Context, phone, code, codeHash string) (*tg.AuthAuthorization, error)
	SendCode(ctx context.Context, phone string, options SendCodeOptions) (tg.AuthSentCodeClass, error)
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
	return UserInfo{}, errors.New("not implemented")
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
		return "", errors.Errorf("environment variable %q not set", env)
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

	length := 5
	if sentCode != nil {
		typ, ok := sentCode.Type.(notFlashing)
		if !ok {
			return "", errors.Errorf("unexpected type: %T", sentCode.Type)
		}
		length = typ.GetLength()
	}

	return strings.Repeat(strconv.Itoa(t.dc), length), nil
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
