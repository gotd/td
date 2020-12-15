package telegram

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// checkAuthResult checks that a is *tg.AuthAuthorization.
func (c *Client) checkAuthResult(a tg.AuthAuthorizationClass) error {
	switch a.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	default:
		return xerrors.Errorf("got unexpected response %T", a)
	}
}
