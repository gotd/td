package hook

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// AffectedHandler applies the pts increment from a messages.affected* RPC result.
//
// It is implemented by *updates.Manager via its HandleAffected method.
type AffectedHandler interface {
	HandleAffected(ctx context.Context, channelID int64, pts, ptsCount int) error
}

// AffectedHook middleware keeps the updates manager's local pts in sync after
// self-initiated reads and deletes.
//
// Methods such as messages.readHistory, messages.deleteMessages and
// channels.deleteMessages return messages.affectedMessages /
// messages.affectedHistory, which carry a pts increment the client must apply.
// These results are not tg.UpdatesClass, so they bypass the regular update hook;
// without applying them the server pts advances while the local pts stays behind,
// making the next genuine update look like a gap (see issue #1382).
//
// Place it in telegram.Options.Middlewares, like UpdateHook.
func AffectedHook(handler AffectedHandler) telegram.Middleware {
	return affectedHook{handler: handler}
}

type affectedHook struct {
	handler AffectedHandler
}

// Handle implements telegram.Middleware.
func (h affectedHook) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		if err := next.Invoke(ctx, input, output); err != nil {
			return err
		}

		pts, ptsCount, ok := affectedPts(output)
		if !ok {
			return nil
		}

		if err := h.handler.HandleAffected(ctx, channelIDFromRequest(input), pts, ptsCount); err != nil {
			return errors.Wrap(err, "affected hook")
		}
		return nil
	}
}

// affectedPts extracts the pts increment from a messages.affected* result.
func affectedPts(output bin.Decoder) (pts, ptsCount int, ok bool) {
	switch o := output.(type) {
	case *tg.MessagesAffectedMessages:
		return o.Pts, o.PtsCount, true
	case *tg.MessagesAffectedHistory:
		return o.Pts, o.PtsCount, true
	default:
		return 0, 0, false
	}
}

// channelIDFromRequest determines which pts sequence an affected result belongs
// to. A request carrying an InputChannel (channels.*) or an InputPeerChannel
// (messages.* targeting a channel) routes to that channel; anything else routes
// to the common sequence (channelID 0).
func channelIDFromRequest(input bin.Encoder) int64 {
	if r, ok := input.(interface {
		GetChannel() tg.InputChannelClass
	}); ok {
		if ch, ok := r.GetChannel().(*tg.InputChannel); ok {
			return ch.ChannelID
		}
		return 0
	}
	if r, ok := input.(interface {
		GetPeer() tg.InputPeerClass
	}); ok {
		if p, ok := r.GetPeer().(*tg.InputPeerChannel); ok {
			return p.ChannelID
		}
	}
	return 0
}
