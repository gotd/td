package telegram

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (c *Client) checkAuthResult(a tg.AuthAuthorizationClass) error {
	switch a.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	default:
		return xerrors.Errorf("got unexpected response %T", a)
	}
}
