package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Support returns support User.
func (m *Manager) Support(ctx context.Context) (User, error) {
	r, err := m.api.HelpGetSupport(ctx)
	if err != nil {
		return User{}, errors.Wrap(err, "get support")
	}

	u, ok := r.User.(*tg.User)
	if !ok {
		return User{}, errors.Errorf("unexpected type %T", u)
	}

	return m.User(u), nil
}
