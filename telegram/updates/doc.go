// Package updates provides a Telegram's state synchronization manager.
//
// It guarantees that all state-sensitive updates will be performed
// in correct order.
//
// Limitations:
//  1. Manager cannot verify stateless types of updates
//     (tg.UpdatesClass without Seq, or tg.UpdateClass without Pts or Qts).
//
//  2. Due to the fact that updates.getDifference and updates.getChannelDifference
//     do not return event sequences, manager cannot guarantee the correctness
//     of these operations. We rely on the server here.
//
//  3. Manager cannot recover the channel gap if there is a ChannelDifferenceTooLong error.
//     Restoring the state in such situation is not the prerogative of this manager.
//     See: https://core.telegram.org/constructor/updates.channelDifferenceTooLong
//
// TODO: Write implementation details.
package updates
