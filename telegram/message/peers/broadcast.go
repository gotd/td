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
