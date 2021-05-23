package telegram

import (
	"github.com/gotd/td/telegram/auth"
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

// Auth returns auth client.
func (c *Client) Auth() *auth.Client {
	return auth.NewClient(
		c.tg, c.rand, c.appID, c.appHash,
	)
}
