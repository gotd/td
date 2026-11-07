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
//     The latest messages returned by the server are passed to the handler, and the gap
//     range is reported via Config.OnChannelGap, so older messages can be fetched manually.
//     See: https://core.telegram.org/constructor/updates.channelDifferenceTooLong
//
// TODO: Write implementation details.
package updates
