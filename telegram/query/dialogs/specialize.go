package dialogs

import (
	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/telegram/query/channels/participants"
	"github.com/nnqq/td/telegram/query/messages"
	"github.com/nnqq/td/telegram/query/photos"
	"github.com/nnqq/td/tg"
)

// Messages returns new messages history query builder for current dialog.
func (e Elem) Messages(raw *tg.Client) *messages.GetHistoryQueryBuilder {
	return messages.NewQueryBuilder(raw).GetHistory(e.Peer)
}

// Search returns new search query builder for current dialog.
func (e Elem) Search(raw *tg.Client) *messages.SearchQueryBuilder {
	return messages.NewQueryBuilder(raw).Search(e.Peer)
}

// Replies returns new replies query builder for current dialog.
func (e Elem) Replies(raw *tg.Client) *messages.GetRepliesQueryBuilder {
	return messages.NewQueryBuilder(raw).GetReplies(e.Peer)
}

// UnreadMentions returns new unread mentions query builder for current dialog.
func (e Elem) UnreadMentions(raw *tg.Client) *messages.GetUnreadMentionsQueryBuilder {
	return messages.NewQueryBuilder(raw).GetUnreadMentions(e.Peer)
}

// RecentLocations returns new live location history query builder for current dialog.
func (e Elem) RecentLocations(raw *tg.Client) *messages.GetRecentLocationsQueryBuilder {
	return messages.NewQueryBuilder(raw).GetRecentLocations(e.Peer)
}

// UserPhotos returns new user photo query builder for current dialog.
// If peer is not user, returns false.
func (e Elem) UserPhotos(raw *tg.Client) (*photos.GetUserPhotosQueryBuilder, bool) {
	user, ok := peer.ToInputUser(e.Peer)
	if !ok {
		return nil, false
	}
	return photos.NewQueryBuilder(raw).GetUserPhotos(user), true
}

// Participants returns new channel participants query builder for current dialog.
// If peer is not channel, returns false.
func (e Elem) Participants(raw *tg.Client) (*participants.GetParticipantsQueryBuilder, bool) {
	channel, ok := peer.ToInputChannel(e.Peer)
	if !ok {
		return nil, false
	}
	return participants.NewQueryBuilder(raw).GetParticipants(channel), true
}

// Deleted denotes that dialog is deleted.
func (e Elem) Deleted() bool {
	_, ok := e.Peer.(*tg.InputPeerEmpty)
	return ok
}
