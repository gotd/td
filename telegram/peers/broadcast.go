package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Broadcast is a broadcast Channel.
type Broadcast struct {
	Channel
}

// Signatures whether signatures are enabled (channels).
func (b Broadcast) Signatures() bool {
	return b.raw.GetSignatures()
}

// DiscussionGroup returns linked chat, if any.
func (b Broadcast) DiscussionGroup(ctx context.Context) (Channel, bool, error) {
	full, err := b.FullRaw(ctx)
	if err != nil {
		return Channel{}, false, err
	}

	id, ok := full.GetLinkedChatID()
	if !ok {
		return Channel{}, false, nil
	}

	ch, err := b.m.GetChat(ctx, id)
	if err != nil {
		return Channel{}, false, err
	}

	actual, ok, err := ch.ActualChat(ctx)
	if err != nil {
		return Channel{}, false, err
	}
	if !ok {
		return Channel{}, false, errors.Errorf("chat %d is linked, but not migrated", id)
	}

	return actual, true, nil
}

// SetDiscussionGroup associates a group to a channel as discussion group for that channel.
func (b Broadcast) SetDiscussionGroup(ctx context.Context, p tg.InputChannelClass) error {
	if _, err := b.m.api.ChannelsSetDiscussionGroup(ctx, &tg.ChannelsSetDiscussionGroupRequest{
		Broadcast: b.InputChannel(),
		Group:     p,
	}); err != nil {
		return errors.Wrap(err, "set discussion group")
	}

	return nil
}

// ToggleSignatures enable/disable message signatures in channels.
func (b Broadcast) ToggleSignatures(ctx context.Context, enabled bool) error {
	if _, err := b.m.api.ChannelsToggleSignatures(ctx, &tg.ChannelsToggleSignaturesRequest{
		Channel: b.InputChannel(),
		Enabled: enabled,
	}); err != nil {
		return errors.Wrap(err, "toggle signatures")
	}

	return nil
}
