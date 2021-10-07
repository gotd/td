package participants

import (
	"github.com/nnqq/td/telegram/query/photos"
	"github.com/nnqq/td/tg"
)

// UserPhotos returns new user photo query builder for participant.
func (e Elem) UserPhotos(raw *tg.Client) (*photos.GetUserPhotosQueryBuilder, bool) {
	user, ok := e.User()
	if !ok {
		return nil, false
	}
	return photos.NewQueryBuilder(raw).GetUserPhotos(user.AsInput()), true
}

// User tries to get participant user object.
func (e Elem) User() (*tg.User, bool) {
	switch part := e.Participant.(type) {
	case interface{ GetUserID() int64 }:
		return e.Entities.User(part.GetUserID())
	case interface{ GetPeer() tg.PeerClass }:
		user, ok := part.GetPeer().(*tg.PeerUser)
		if !ok {
			return nil, false
		}

		return e.Entities.User(user.GetUserID())
	default:
		return nil, false
	}
}

// Creator returns participant user object and meta info if participant is a creator of channel.
func (e Elem) Creator() (*tg.User, *tg.ChannelParticipantCreator, bool) {
	part, ok := e.Participant.(*tg.ChannelParticipantCreator)
	if !ok {
		return nil, nil, false
	}

	user, ok := e.User()
	if !ok {
		return nil, nil, false
	}

	return user, part, true
}

// Admin returns participant user object and meta info if participant is admin of channel.
func (e Elem) Admin() (*tg.User, *tg.ChannelParticipantAdmin, bool) {
	part, ok := e.Participant.(*tg.ChannelParticipantAdmin)
	if !ok {
		return nil, nil, false
	}

	user, ok := e.User()
	if !ok {
		return nil, nil, false
	}

	return user, part, true
}
