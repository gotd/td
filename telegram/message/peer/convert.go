package peer

import (
	"github.com/nnqq/td/tg"
)

// ToInputUser converts given peer to input user if possible.
func ToInputUser(user tg.InputPeerClass) (tg.InputUserClass, bool) {
	switch u := user.(type) {
	case *tg.InputPeerUser:
		v := new(tg.InputUser)
		v.FillFrom(u)
		return v, true
	case *tg.InputPeerUserFromMessage:
		v := new(tg.InputUserFromMessage)
		v.FillFrom(u)
		return v, true
	case *tg.InputPeerSelf:
		v := new(tg.InputUserSelf)
		return v, true
	default:
		return nil, false
	}
}

// ToInputChannel converts given peer to input channel if possible.
func ToInputChannel(channel tg.InputPeerClass) (tg.InputChannelClass, bool) {
	switch u := channel.(type) {
	case *tg.InputPeerChannel:
		v := new(tg.InputChannel)
		v.FillFrom(u)
		return v, true
	case *tg.InputPeerChannelFromMessage:
		v := new(tg.InputChannelFromMessage)
		v.FillFrom(u)
		return v, true
	default:
		return nil, false
	}
}
