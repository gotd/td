package updates

import (
	"github.com/gotd/td/tg"
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
	GetState(userID int) (state State, found bool, err error)
	SetState(userID int, state State) error
	SetPts(userID, pts int) error
	SetQts(userID, qts int) error
	SetDate(userID, date int) error
	SetSeq(userID, seq int) error
	SetDateSeq(userID, date, seq int) error
	GetChannelPts(userID, channelID int) (pts int, found bool, err error)
	SetChannelPts(userID, channelID, pts int) error
	ForEachChannels(userID int, f func(channelID, pts int) error) error
}

// ChannelAccessHasher stores users channel access hashes.
type ChannelAccessHasher interface {
	SetChannelAccessHash(userID, channelID int, accessHash int64) error
	GetChannelAccessHash(userID, channelID int) (accessHash int64, found bool, err error)
}
