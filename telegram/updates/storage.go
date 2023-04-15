package updates

import (
	"context"

	"github.com/gotd/td/tg"
)

// State is the user internalState.
type State struct {
	Pts, Qts, Date, Seq int
}

func (s State) fromRemote(remote *tg.UpdatesState) State {
	return State{
		Pts:  remote.Pts,
		Qts:  remote.Qts,
		Date: remote.Date,
		Seq:  remote.Seq,
	}
}

// StateStorage is the users internalState storage.
//
// Note:
// SetPts, SetQts, SetDate, SetSeq, SetDateSeq
// should return error if user internalState does not exist.
type StateStorage interface {
	GetState(ctx context.Context, userID int64) (state State, found bool, err error)
	SetState(ctx context.Context, userID int64, state State) error
	SetPts(ctx context.Context, userID int64, pts int) error
	SetQts(ctx context.Context, userID int64, qts int) error
	SetDate(ctx context.Context, userID int64, date int) error
	SetSeq(ctx context.Context, userID int64, seq int) error
	SetDateSeq(ctx context.Context, userID int64, date, seq int) error
	GetChannelPts(ctx context.Context, userID, channelID int64) (pts int, found bool, err error)
	SetChannelPts(ctx context.Context, userID, channelID int64, pts int) error
	ForEachChannels(ctx context.Context, userID int64, f func(ctx context.Context, channelID int64, pts int) error) error
}

// ChannelAccessHasher stores users channel access hashes.
type ChannelAccessHasher interface {
	SetChannelAccessHash(ctx context.Context, userID, channelID, accessHash int64) error
	GetChannelAccessHash(ctx context.Context, userID, channelID int64) (accessHash int64, found bool, err error)
}
