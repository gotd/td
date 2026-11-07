package stories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func story(id int) *tg.StoryItem {
	return &tg.StoryItem{ID: id, Media: &tg.MessageMediaEmpty{}}
}

func TestGetAllStories(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	q := NewQueryBuilder(tg.NewClient(mock))

	mock.ExpectCall(&tg.StoriesGetAllStoriesRequest{Hidden: true}).ThenResult(&tg.StoriesAllStories{
		HasMore: true,
		State:   "s1",
		PeerStories: []tg.PeerStories{
			{Peer: &tg.PeerUser{UserID: 1}, Stories: []tg.StoryItemClass{story(1)}},
		},
		Users:       []tg.UserClass{&tg.User{ID: 1, AccessHash: 2}},
		StealthMode: tg.StoriesStealthMode{},
	})
	mock.ExpectCall(&tg.StoriesGetAllStoriesRequest{Hidden: true, Next: true, State: "s1"}).ThenResult(&tg.StoriesAllStories{
		State: "s2",
		PeerStories: []tg.PeerStories{
			{Peer: &tg.PeerUser{UserID: 3}, Stories: []tg.StoryItemClass{story(2)}},
		},
		StealthMode: tg.StoriesStealthMode{},
	})

	var peers []int64
	state, err := q.GetAllStories().Hidden(true).ForEach(ctx, func(ctx context.Context, s tg.PeerStories, e peer.Entities) error {
		peers = append(peers, s.Peer.(*tg.PeerUser).UserID)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, "s2", state)
	require.Equal(t, []int64{1, 3}, peers)

	mock.ExpectCall(&tg.StoriesGetAllStoriesRequest{State: "s2"}).ThenResult(&tg.StoriesAllStoriesNotModified{State: "s3"})
	state, err = q.GetAllStories().State("s2").ForEach(ctx, func(context.Context, tg.PeerStories, peer.Entities) error {
		t.Fatal("must not be called")
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, "s3", state)
}

func TestGetPinnedStories(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	p := &tg.InputPeerSelf{}

	mock.ExpectCall(&tg.StoriesGetPinnedStoriesRequest{Peer: p, Limit: 2}).
		ThenResult(&tg.StoriesStories{Stories: []tg.StoryItemClass{story(10), &tg.StoryItemSkipped{ID: 9}}})
	mock.ExpectCall(&tg.StoriesGetPinnedStoriesRequest{Peer: p, OffsetID: 9, Limit: 2}).
		ThenResult(&tg.StoriesStories{Stories: []tg.StoryItemClass{story(8)}})

	var got []int
	var skipped int
	err := NewQueryBuilder(tg.NewClient(mock)).GetPinnedStories(p).BatchSize(2).
		ForEach(ctx, func(ctx context.Context, raw tg.StoryItemClass, e Elem) error {
			got = append(got, raw.GetID())
			if e.Story == nil {
				skipped++
			}
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, []int{10, 9, 8}, got)
	require.Equal(t, 1, skipped)
}

func TestGetStoriesArchive(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	p := &tg.InputPeerSelf{}
	mock.ExpectCall(&tg.StoriesGetStoriesArchiveRequest{Peer: p, OffsetID: 5, Limit: defaultBatchSize}).
		ThenResult(&tg.StoriesStories{Stories: []tg.StoryItemClass{story(4)}})

	var got []int
	err := NewQueryBuilder(tg.NewClient(mock)).GetStoriesArchive(p).OffsetID(5).
		ForEach(ctx, func(ctx context.Context, raw tg.StoryItemClass, e Elem) error {
			got = append(got, e.Story.ID)
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, []int{4}, got)
}

func TestGetStoriesByID(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	p := &tg.InputPeerSelf{}
	ids := make([]int, 150)
	for i := range ids {
		ids[i] = i + 1
	}
	mock.ExpectCall(&tg.StoriesGetStoriesByIDRequest{Peer: p, ID: ids[:100]}).
		ThenResult(&tg.StoriesStories{Stories: []tg.StoryItemClass{story(1)}, Users: []tg.UserClass{&tg.User{ID: 7}}})
	mock.ExpectCall(&tg.StoriesGetStoriesByIDRequest{Peer: p, ID: ids[100:]}).
		ThenResult(&tg.StoriesStories{Stories: []tg.StoryItemClass{&tg.StoryItemDeleted{ID: 150}}})

	stories, e, err := NewQueryBuilder(tg.NewClient(mock)).GetStoriesByID(ctx, p, ids...)
	require.NoError(t, err)
	require.Len(t, stories, 2)
	_, ok := e.User(7)
	require.True(t, ok)
}

func TestGetStoryViewsList(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	p := &tg.InputPeerSelf{}
	first := &tg.StoriesGetStoryViewsListRequest{Peer: p, ID: 3, Limit: 2, JustContacts: true}
	first.SetQ("bob")
	mock.ExpectCall(first).ThenResult(&tg.StoriesStoryViewsList{
		Views:      []tg.StoryViewClass{&tg.StoryView{UserID: 1}, &tg.StoryView{UserID: 2}},
		NextOffset: "next",
	})
	second := *first
	second.Offset = "next"
	mock.ExpectCall(&second).ThenResult(&tg.StoriesStoryViewsList{
		Views: []tg.StoryViewClass{&tg.StoryView{UserID: 3}},
	})

	var got []int64
	err := NewQueryBuilder(tg.NewClient(mock)).GetStoryViewsList(p, 3).
		BatchSize(2).JustContacts(true).Query("bob").
		ForEach(ctx, func(ctx context.Context, v tg.StoryViewClass, e peer.Entities) error {
			got = append(got, v.(*tg.StoryView).UserID)
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, []int64{1, 2, 3}, got)
}
