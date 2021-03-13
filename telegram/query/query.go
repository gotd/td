// Package query contains generic pagination helpers.
package query

import (
	"github.com/gotd/td/telegram/query/channels/participants"
	"github.com/gotd/td/telegram/query/contacts/blocked"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/telegram/query/messages/stickers/featured"
	"github.com/gotd/td/telegram/query/photos"
	"github.com/gotd/td/tg"
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

// Messages creates messages.QueryBuilder
func (q *Query) Messages() *messages.QueryBuilder {
	return messages.NewQueryBuilder(q.raw)
}

// Featured creates featured.QueryBuilder
func (q *Query) Featured() *featured.QueryBuilder {
	return featured.NewQueryBuilder(q.raw)
}
