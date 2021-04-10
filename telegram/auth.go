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

// checkAuthResult checks that a is *tg.AuthAuthorization and returns user authorization info.
func checkAuthResult(a tg.AuthAuthorizationClass) (*tg.User, error) {
	switch a := a.(type) {
	case *tg.AuthAuthorization:
		switch u := a.User.(type) {
		case *tg.User:
			return u, nil // ok
		case *tg.UserEmpty:
			// Should be unreachable, but just in case
			// map empty user to full user.
			return &tg.User{
				ID:   u.ID,
				Self: u.ID != 0,
			}, nil
		default:
			return nil, xerrors.Errorf("got unexpected user type %T", a)
		}
	case *tg.AuthAuthorizationSignUpRequired:
		return nil, &SignUpRequired{
			TermsOfService: a.TermsOfService,
		}
	default:
		return nil, xerrors.Errorf("got unexpected response %T", a)
	}
}
