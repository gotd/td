// Package query contains generic pagination helpers.
package query

import (
	"github.com/nnqq/td/telegram/query/channels/participants"
	"github.com/nnqq/td/telegram/query/contacts/blocked"
	"github.com/nnqq/td/telegram/query/dialogs"
	"github.com/nnqq/td/telegram/query/messages"
	"github.com/nnqq/td/telegram/query/messages/stickers/featured"
	"github.com/nnqq/td/telegram/query/photos"
	"github.com/nnqq/td/tg"
)

// Query is common struct to create query builders.
type Query struct {
	raw *tg.Client
}

// NewQuery creates Query.
func NewQuery(raw *tg.Client) *Query {
	return &Query{raw: raw}
}

// Participants creates participants.QueryBuilder
func (q *Query) Participants() *participants.QueryBuilder {
	return participants.NewQueryBuilder(q.raw)
}

// Blocked creates blocked.QueryBuilder
func (q *Query) Blocked() *blocked.QueryBuilder {
	return blocked.NewQueryBuilder(q.raw)
}

// Photos creates photos.QueryBuilder
func (q *Query) Photos() *photos.QueryBuilder {
	return photos.NewQueryBuilder(q.raw)
}

// Dialogs creates dialogs.QueryBuilder
func (q *Query) Dialogs() *dialogs.QueryBuilder {
	return dialogs.NewQueryBuilder(q.raw)
}

// Messages creates messages.QueryBuilder.
func (q *Query) Messages() *messages.QueryBuilder {
	return messages.NewQueryBuilder(q.raw)
}

// Featured creates featured.QueryBuilder
func (q *Query) Featured() *featured.QueryBuilder {
	return featured.NewQueryBuilder(q.raw)
}

// GetParticipants creates participants.GetParticipantsQueryBuilder.
func (q *Query) GetParticipants(channel tg.InputChannelClass) *participants.GetParticipantsQueryBuilder {
	return participants.NewQueryBuilder(q.raw).GetParticipants(channel)
}

// GetParticipants creates participants.GetParticipantsQueryBuilder.
// Shorthand for
//
//	query.NewQuery(raw).GetParticipants(channel)
//
func GetParticipants(raw *tg.Client, channel tg.InputChannelClass) *participants.GetParticipantsQueryBuilder {
	return NewQuery(raw).GetParticipants(channel)
}

// GetBlocked creates blocked.GetBlockedQueryBuilder.
func (q *Query) GetBlocked() *blocked.GetBlockedQueryBuilder {
	return blocked.NewQueryBuilder(q.raw).GetBlocked()
}

// GetBlocked creates blocked.GetBlockedQueryBuilder.
// Shorthand for
//
//	query.NewQuery(raw).GetBlocked()
//
func GetBlocked(raw *tg.Client) *blocked.GetBlockedQueryBuilder {
	return NewQuery(raw).GetBlocked()
}

// GetUserPhotos creates photos.GetUserPhotosQueryBuilder.
func (q *Query) GetUserPhotos(user tg.InputUserClass) *photos.GetUserPhotosQueryBuilder {
	return photos.NewQueryBuilder(q.raw).GetUserPhotos(user)
}

// GetUserPhotos creates photos.GetUserPhotosQueryBuilder.
// Shorthand for
//
//	query.NewQuery(raw).GetUserPhotos(user)
//
func GetUserPhotos(raw *tg.Client, user tg.InputUserClass) *photos.GetUserPhotosQueryBuilder {
	return NewQuery(raw).GetUserPhotos(user)
}

// GetDialogs creates dialogs.GetDialogsQueryBuilder.
func (q *Query) GetDialogs() *dialogs.GetDialogsQueryBuilder {
	return dialogs.NewQueryBuilder(q.raw).GetDialogs()
}

// Messages creates messages.QueryBuilder.
// Shorthand for
//
//	query.NewQuery(raw).Messages()
//
func Messages(raw *tg.Client) *messages.QueryBuilder {
	return NewQuery(raw).Messages()
}

// GetDialogs creates dialogs.GetDialogsQueryBuilder.
// Shorthand for
//
//	query.NewQuery(raw).GetDialogs()
//
func GetDialogs(raw *tg.Client) *dialogs.GetDialogsQueryBuilder {
	return NewQuery(raw).GetDialogs()
}

// GetOldFeaturedStickers creates featured.QueryBuilder.
func (q *Query) GetOldFeaturedStickers() *featured.GetOldFeaturedStickersQueryBuilder {
	return featured.NewQueryBuilder(q.raw).GetOldFeaturedStickers()
}

// GetOldFeaturedStickers creates featured.QueryBuilder.
// Shorthand for
//
//	query.NewQuery(raw).GetOldFeaturedStickers()
//
func GetOldFeaturedStickers(raw *tg.Client) *featured.GetOldFeaturedStickersQueryBuilder {
	return NewQuery(raw).GetOldFeaturedStickers()
}
