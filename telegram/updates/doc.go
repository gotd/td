// Package updates provides a Telegram's internalState synchronization manager.
//
// It guarantees that all internalState-sensitive updates will be performed
// in correct order.
//
// Limitations:
//
//  1. Manager cannot verify stateless types of updates
//     (tg.UpdatesClass without Seq, or tg.UpdateClass without Pts or Qts).
//
//  2. Due to the fact that updates.getDifference and updates.getChannelDifference
//     do not return event sequences, manager cannot guarantee the correctness
//     of these operations. We rely on the server here.
//
//  3. Manager cannot fully recover the channel gap if there is a ChannelDifferenceTooLong error.
//     The gap range and the latest messages returned by the server are reported via
//     Config.OnChannelGap (not the handler, as they may have been delivered already),
//     so missed messages can be fetched manually.
//     See: https://core.telegram.org/constructor/updates.channelDifferenceTooLong
//
// # Failed differences
//
// A failed updates.getDifference or updates.getChannelDifference is retried
// with backoff (1 s doubling up to 60 s, then every 60–80 s), like TDLib and
// Telegram Desktop; FLOOD_WAIT is waited out. A lost authorization is not
// retried. PERSISTENT_TIMESTAMP_INVALID and PERSISTENT_TIMESTAMP_EMPTY on
// the common sequence reset the state from updates.getState and call
// Config.OnTooLong, as the updates in between are lost. CHANNEL_INVALID
// stops tracking the channel, like CHANNEL_PRIVATE. Config.Clock drives the
// retries.
//
// # Observability
//
// Events are logged via Config.Logger: debug logs for every difference
// request, gap, tracked channel and update received before Run, warnings for
// abnormal ones (e.g. a channel gap whose pts did not advance).
//
// With Config.MeterProvider, the manager records these OpenTelemetry metrics
// (meter "github.com/gotd/td/telegram/updates"). Attributes never carry
// channel or user IDs.
//
//   - gotd.updates.difference.requests (counter): getDifference (kind=common)
//     and getChannelDifference (kind=channel) calls by result: empty, slice,
//     full, too_long, error, inaccessible.
//   - gotd.updates.difference.duration (histogram, s): duration of those
//     calls, same attributes.
//   - gotd.updates.difference.triggers (counter): why the difference was
//     requested, by kind and reason (e.g. startup, pts-gap-timeout,
//     idle-timeout, channel-too-long, and retry or channel-retry for a
//     retry of a failed difference).
//   - gotd.updates.gaps (counter): sequence gaps by kind, seq (pts, qts, seq)
//     and action: detected, resolved_by_waiting, resolved_by_pending. Gaps
//     not resolved in time trigger a difference with a *-gap-timeout reason.
//   - gotd.updates.channel.too_long (counter): channelDifferenceTooLong, i.e.
//     Config.OnChannelGap calls, by dialog: ok, stale (pts did not advance),
//     invalid (no dialog pts).
//   - gotd.updates.channel.gap.size (histogram, {pts}): RemotePts - LocalPts
//     of those gaps.
//   - gotd.updates.channel.gap.messages (histogram, {message}): length of
//     ChannelGap.Messages.
//   - gotd.updates.before_run (counter, {batch}): updates received before Run
//     by action: passed, buffered, dropped, replayed, passed_raw (replayed to
//     the handler as-is because the state was fetched from the server).
//   - gotd.updates.channel.track (counter): channel tracking by source (api
//     for Manager.TrackChannel, stored, update, too_long) and result
//     (started, already, pending, stopped, canceled, error).
//   - gotd.updates.channels.tracked (up-down counter): channels tracked by
//     running managers.
//   - gotd.updates.channel.inaccessible (counter): channels removed from
//     tracking after CHANNEL_PRIVATE or CHANNEL_INVALID.
//   - gotd.updates.stopped (counter): calls that returned ErrStopped, by op:
//     handle, handle_affected, track_channel.
//   - gotd.updates.dispatched (counter, {update}): updates passed to the
//     handler, by source: live or difference.
package updates
