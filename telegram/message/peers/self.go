package peers

import (
	"context"

	"github.com/gotd/td/tg"
)

// Self returns current User.
func (m *Manager) Self(ctx context.Context) (User, error) {
	return m.GetUser(ctx, &tg.InputUserSelf{})
}

func (m *Manager) selfIsBot() bool  {
	u, ok := m.me.Load()
	return ok && u.Bot
}
