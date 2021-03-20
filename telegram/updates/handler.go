package updates

import (
	"context"

	"github.com/gotd/td/tg"
)

// Handle implements Handler to use Manager as middleware.
func (m *Manager) Handle(ctx context.Context, u *tg.Updates) error {
	return m.applyUpdates(ctx, u)
}

// HandleShort implements Handler to use Manager as middleware.
func (m *Manager) HandleShort(ctx context.Context, u *tg.UpdateShort) error {
	return m.applyUpdates(ctx, &tg.Updates{
		Updates: []tg.UpdateClass{u.Update},
		Date:    u.Date,
	})
}
