package updates

import (
	"github.com/nnqq/td/tg"
)

// State is the user state.
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

// StateStorage is the users state storage.
//
// Note:
// SetPts, SetQts, SetDate, SetSeq, SetDateSeq
// should return error if user state does not exist.
type StateStorage interface {
	GetState(userID int64) (state State, found bool, err error)
	SetState(userID int64, state State) error
	SetPts(userID int64, pts int) error
	SetQts(userID int64, qts int) error
	SetDate(userID int64, date int) error
	SetSeq(userID int64, seq int) error
	SetDateSeq(userID int64, date, seq int) error
	GetChannelPts(userID, channelID int64) (pts int, found bool, err error)
	SetChannelPts(userID, channelID int64, pts int) error
	ForEachChannels(userID int64, f func(channelID int64, pts int) error) error
}

// ChannelAccessHasher stores users channel access hashes.
type ChannelAccessHasher interface {
	SetChannelAccessHash(userID, channelID, accessHash int64) error
	GetChannelAccessHash(userID, channelID int64) (accessHash int64, found bool, err error)
}
