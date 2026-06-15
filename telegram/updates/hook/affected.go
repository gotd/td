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

		channelID, ok := channelIDFromRequest(input)
		if !ok {
			// A channel-scoped request whose channel could not be identified
			// (e.g. InputChannelEmpty). Skip rather than misapply a channel pts
			// to the common sequence, which would desync update processing.
			return nil
		}

		if err := h.handler.HandleAffected(ctx, channelID, pts, ptsCount); err != nil {
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

// hasChannelID is implemented by every channel-bearing input type that carries a
// bare channel ID: *tg.InputChannel, *tg.InputChannelFromMessage,
// *tg.InputPeerChannel and *tg.InputPeerChannelFromMessage. The non-channel input
// peers (user/chat/self/empty) do not implement it.
type hasChannelID interface {
	GetChannelID() int64
}

// channelIDFromRequest determines which pts sequence an affected result belongs
// to. ok is false only for a channel-scoped request (channels.*) whose channel
// cannot be identified, in which case the caller must skip rather than route to
// the common sequence.
//
//   - channels.* request: routes to the request's channel (ok=false if it carries
//     no channel ID, e.g. InputChannelEmpty).
//   - messages.* request with an InputPeerChannel(FromMessage) peer: that channel.
//   - any other request (user/chat peer, or no peer at all): common sequence (0).
func channelIDFromRequest(input bin.Encoder) (channelID int64, ok bool) {
	if r, ok := input.(interface {
		GetChannel() tg.InputChannelClass
	}); ok {
		if ch, ok := r.GetChannel().(hasChannelID); ok {
			return ch.GetChannelID(), true
		}
		return 0, false
	}
	if r, ok := input.(interface {
		GetPeer() tg.InputPeerClass
	}); ok {
		if p, ok := r.GetPeer().(hasChannelID); ok {
			return p.GetChannelID(), true
		}
	}
	return 0, true
}
