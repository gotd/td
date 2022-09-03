package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

func (m *Manager) editAbout(ctx context.Context, p tg.InputPeerClass, about string) error {
	if _, err := m.api.MessagesEditChatAbout(ctx, &tg.MessagesEditChatAboutRequest{
		Peer:  p,
		About: about,
	}); err != nil {
		if _, ok := p.(*tg.InputPeerChat); ok {
			return errors.Wrap(err, "edit chat about")
		}
		return errors.Wrap(err, "edit channel about")
	}
	return nil
}

func (m *Manager) editReactions(ctx context.Context, p tg.InputPeerClass, reactions tg.ChatReactionsClass) error {
	if _, err := m.api.MessagesSetChatAvailableReactions(ctx, &tg.MessagesSetChatAvailableReactionsRequest{
		Peer:               p,
		AvailableReactions: reactions,
	}); err != nil {
		return errors.Wrap(err, "set reactions")
	}
	return nil
}
