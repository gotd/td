// Package updates provides a Telegram's state synchronization engine.
//
// It guarantees that all state-sensitive updates will be performed
// in correct order.
//
// Limitations:
//  1. Engine cannot verify stateless types of updates
//     (tg.UpdatesClass without Seq, or tg.UpdateClass without Pts or Qts).
//
//  2. Due to the fact that updates.getDifference and updates.getChannelDifference
//     do not return event sequences, the engine cannot guarantee the correctness
//     of these operations. We rely on the server here.
//
//  3. Engine cannot recover the channel gap if there is a ChannelDifferenceTooLong error.
//     Restoring the state in such situation is not the prerogative of this engine.
//     See: https://core.telegram.org/constructor/updates.channelDifferenceTooLong
//
// TODO: Write implementation details.
package updates
