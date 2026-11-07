// Package stories contains helpers to iterate over stories.
//
// See https://core.telegram.org/api/stories.
package stories

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// QueryBuilder is a helper to create stories queries.
type QueryBuilder struct {
	raw *tg.Client
}

// NewQueryBuilder creates new QueryBuilder.
func NewQueryBuilder(raw *tg.Client) *QueryBuilder {
	return &QueryBuilder{raw: raw}
}

const (
	defaultBatchSize = 100
	// maxIDsPerRequest is the limit of story IDs in stories.getStoriesByID.
	maxIDsPerRequest = 100
)

// Elem is a story with entities of the response.
type Elem struct {
	Story    *tg.StoryItem
	Entities peer.Entities
}

// AllStoriesQueryBuilder is a query builder of stories.getAllStories.
type AllStoriesQueryBuilder struct {
	raw    *tg.Client
	hidden bool
	state  string
}

// GetAllStories creates query of active stories of all peers the user
// follows (contacts and channels).
func (q *QueryBuilder) GetAllStories() *AllStoriesQueryBuilder {
	return &AllStoriesQueryBuilder{raw: q.raw}
}

// Hidden sets whether to fetch stories of hidden peers (the "archive" of
// stories list) instead of visible ones.
func (b *AllStoriesQueryBuilder) Hidden(hidden bool) *AllStoriesQueryBuilder {
	b.hidden = hidden
	return b
}

// State sets the state returned by previous iteration (see ForEach), to
// fetch only changes since then.
func (b *AllStoriesQueryBuilder) State(state string) *AllStoriesQueryBuilder {
	b.state = state
	return b
}

// ForEach calls cb on each peer with active stories, following pagination,
// and returns the state to pass to State for the next call.
//
// If nothing changed since given state, cb is not called.
func (b *AllStoriesQueryBuilder) ForEach(
	ctx context.Context,
	cb func(ctx context.Context, stories tg.PeerStories, e peer.Entities) error,
) (string, error) {
	state := b.state
	next := false
	for {
		req := &tg.StoriesGetAllStoriesRequest{
			Hidden: b.hidden,
			State:  state,
			Next:   next,
		}
		res, err := b.raw.StoriesGetAllStories(ctx, req)
		if err != nil {
			return state, errors.Wrap(err, "get all stories")
		}

		switch r := res.(type) {
		case *tg.StoriesAllStoriesNotModified:
			return r.State, nil
		case *tg.StoriesAllStories:
			e := peer.EntitiesFromResult(r)
			for _, s := range r.PeerStories {
				if err := cb(ctx, s, e); err != nil {
					return state, err
				}
			}
			state = r.State
			if !r.HasMore {
				return state, nil
			}
			next = true
		default:
			return state, errors.Errorf("unexpected type %T", res)
		}
	}
}

// storiesFunc fetches a page of stories.
type storiesFunc func(ctx context.Context, offsetID, limit int) (*tg.StoriesStories, error)

// StoriesQueryBuilder is a query builder of paginated stories of a peer
// (pinned or archive).
type StoriesQueryBuilder struct {
	fetch     storiesFunc
	batchSize int
	offsetID  int
}

// GetPinnedStories creates query of stories pinned to the profile of peer.
func (q *QueryBuilder) GetPinnedStories(p tg.InputPeerClass) *StoriesQueryBuilder {
	return &StoriesQueryBuilder{
		batchSize: defaultBatchSize,
		fetch: func(ctx context.Context, offsetID, limit int) (*tg.StoriesStories, error) {
			return q.raw.StoriesGetPinnedStories(ctx, &tg.StoriesGetPinnedStoriesRequest{
				Peer:     p,
				OffsetID: offsetID,
				Limit:    limit,
			})
		},
	}
}

// GetStoriesArchive creates query of archived (expired) stories of peer,
// available to its owner or admins.
func (q *QueryBuilder) GetStoriesArchive(p tg.InputPeerClass) *StoriesQueryBuilder {
	return &StoriesQueryBuilder{
		batchSize: defaultBatchSize,
		fetch: func(ctx context.Context, offsetID, limit int) (*tg.StoriesStories, error) {
			return q.raw.StoriesGetStoriesArchive(ctx, &tg.StoriesGetStoriesArchiveRequest{
				Peer:     p,
				OffsetID: offsetID,
				Limit:    limit,
			})
		},
	}
}

// BatchSize sets number of stories per request.
func (b *StoriesQueryBuilder) BatchSize(batchSize int) *StoriesQueryBuilder {
	if batchSize > 0 {
		b.batchSize = batchSize
	}
	return b
}

// OffsetID sets story ID to start from (exclusive), e.g. to resume.
func (b *StoriesQueryBuilder) OffsetID(offsetID int) *StoriesQueryBuilder {
	b.offsetID = offsetID
	return b
}

// ForEach calls cb on each story, from newest to oldest.
//
// Deleted and skipped stories (storyItemDeleted, storyItemSkipped) are
// passed as well via raw, Elem.Story is nil for them; use GetStoriesByID to
// fetch skipped ones.
func (b *StoriesQueryBuilder) ForEach(ctx context.Context, cb func(ctx context.Context, raw tg.StoryItemClass, e Elem) error) error {
	offset := b.offsetID
	for {
		res, err := b.fetch(ctx, offset, b.batchSize)
		if err != nil {
			return errors.Wrap(err, "get stories")
		}
		e := peer.EntitiesFromResult(res)
		for _, s := range res.Stories {
			item, _ := s.(*tg.StoryItem)
			if err := cb(ctx, s, Elem{Story: item, Entities: e}); err != nil {
				return err
			}
			offset = s.GetID()
		}
		if len(res.Stories) < b.batchSize {
			return nil
		}
	}
}

// GetStoriesByID fetches stories of peer by IDs, in chunks of 100.
//
// Returned stories may include storyItemDeleted for deleted ones.
func (q *QueryBuilder) GetStoriesByID(ctx context.Context, p tg.InputPeerClass, ids ...int) ([]tg.StoryItemClass, peer.Entities, error) {
	var (
		stories  []tg.StoryItemClass
		entities = peer.NewEntities(map[int64]*tg.User{}, map[int64]*tg.Chat{}, map[int64]*tg.Channel{})
	)
	for len(ids) > 0 {
		n := min(len(ids), maxIDsPerRequest)
		res, err := q.raw.StoriesGetStoriesByID(ctx, &tg.StoriesGetStoriesByIDRequest{
			Peer: p,
			ID:   ids[:n],
		})
		if err != nil {
			return nil, entities, errors.Wrap(err, "get stories by id")
		}
		ids = ids[n:]
		stories = append(stories, res.Stories...)
		entities.FillFromResult(res)
	}
	return stories, entities, nil
}

// ViewsQueryBuilder is a query builder of stories.getStoryViewsList.
type ViewsQueryBuilder struct {
	raw *tg.Client
	req tg.StoriesGetStoryViewsListRequest
}

// GetStoryViewsList creates query of views of own story id of peer (self
// or a channel where the user is an admin).
func (q *QueryBuilder) GetStoryViewsList(p tg.InputPeerClass, id int) *ViewsQueryBuilder {
	return &ViewsQueryBuilder{
		raw: q.raw,
		req: tg.StoriesGetStoryViewsListRequest{
			Peer:  p,
			ID:    id,
			Limit: defaultBatchSize,
		},
	}
}

// BatchSize sets number of views per request.
func (b *ViewsQueryBuilder) BatchSize(batchSize int) *ViewsQueryBuilder {
	if batchSize > 0 {
		b.req.Limit = batchSize
	}
	return b
}

// JustContacts sets whether to return only views of contacts.
func (b *ViewsQueryBuilder) JustContacts(v bool) *ViewsQueryBuilder {
	b.req.JustContacts = v
	return b
}

// ReactionsFirst sets whether to return views with reactions first.
func (b *ViewsQueryBuilder) ReactionsFirst(v bool) *ViewsQueryBuilder {
	b.req.ReactionsFirst = v
	return b
}

// ForwardsFirst sets whether to return forwards and reposts first.
func (b *ViewsQueryBuilder) ForwardsFirst(v bool) *ViewsQueryBuilder {
	b.req.ForwardsFirst = v
	return b
}

// Query sets search query for views.
func (b *ViewsQueryBuilder) Query(q string) *ViewsQueryBuilder {
	b.req.SetQ(q)
	return b
}

// Offset sets offset to start from (NextOffset of a previous page), e.g. to
// resume.
func (b *ViewsQueryBuilder) Offset(offset string) *ViewsQueryBuilder {
	b.req.Offset = offset
	return b
}

// ForEach calls cb on each story view, following pagination.
func (b *ViewsQueryBuilder) ForEach(ctx context.Context, cb func(ctx context.Context, view tg.StoryViewClass, e peer.Entities) error) error {
	req := b.req
	for {
		res, err := b.raw.StoriesGetStoryViewsList(ctx, &req)
		if err != nil {
			return errors.Wrap(err, "get story views list")
		}
		e := peer.EntitiesFromResult(res)
		for _, v := range res.Views {
			if err := cb(ctx, v, e); err != nil {
				return err
			}
		}
		next, ok := res.GetNextOffset()
		if !ok || next == "" || len(res.Views) == 0 {
			return nil
		}
		req.Offset = next
	}
}
