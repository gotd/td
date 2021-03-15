package telegram

import (
	"golang.org/x/xerrors"

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

// checkAuthResult checks that a is *tg.AuthAuthorization.
func checkAuthResult(a tg.AuthAuthorizationClass) error {
	switch v := a.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	case *tg.AuthAuthorizationSignUpRequired:
		return &SignUpRequired{
			TermsOfService: v.TermsOfService,
		}
	default:
		return xerrors.Errorf("got unexpected response %T", a)
	}
}
