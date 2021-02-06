package telegram

import (
	"context"

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
func (c *Client) checkAuthResult(a tg.AuthAuthorizationClass) error {
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

// transfer exports current authorization to another DC.
// See https://core.telegram.org/api/datacenter#authorization-transfer.
func (c *Client) transfer(ctx context.Context, to *tg.Client, dc int) (tg.AuthAuthorizationClass, error) {
	auth, err := c.tg.AuthExportAuthorization(ctx, dc)
	if err != nil {
		return nil, xerrors.Errorf("export to %d: %w", dc, err)
	}

	r, err := to.AuthImportAuthorization(ctx, &tg.AuthImportAuthorizationRequest{
		ID:    auth.ID,
		Bytes: auth.Bytes,
	})
	if err != nil {
		return nil, xerrors.Errorf("import from %d: %w", dc, err)
	}

	return r, nil
}
