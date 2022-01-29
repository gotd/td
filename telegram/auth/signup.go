package auth

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// SignUpRequired means that log in failed because corresponding account
// does not exist, so sign up is required.
type SignUpRequired struct {
	TermsOfService tg.HelpTermsOfService
}

// Is returns true if err is SignUpRequired.
func (s *SignUpRequired) Is(err error) bool {
	_, ok := err.(*SignUpRequired)
	return ok
}

func (s *SignUpRequired) Error() string {
	return "account with provided number does not exist (sign up required)"
}

// checkResult checks that `a` is *tg.AuthAuthorization and returns authorization result or error.
func checkResult(a tg.AuthAuthorizationClass) (*tg.AuthAuthorization, error) {
	switch a := a.(type) {
	case *tg.AuthAuthorization:
		return a, nil // ok
	case *tg.AuthAuthorizationSignUpRequired:
		return nil, &SignUpRequired{
			TermsOfService: a.TermsOfService,
		}
	default:
		return nil, errors.Errorf("got unexpected response %T", a)
	}
}
