package fileref

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestMessage(t *testing.T) {
	ctx := context.Background()
	docLoc := &tg.InputDocumentFileLocation{ID: 10, AccessHash: 1, FileReference: []byte{1}, ThumbSize: "m"}
	fresh := &tg.Document{ID: 10, AccessHash: 2, FileReference: []byte{2}}
	msg := &tg.Message{ID: 5, PeerID: &tg.PeerUser{UserID: 1}, Media: &tg.MessageMediaDocument{
		Document:     &tg.Document{ID: 11},
		AltDocuments: []tg.DocumentClass{fresh},
	}}

	t.Run("Channel", func(t *testing.T) {
		m := tgmock.NewRequire(t)
		m.ExpectCall(&tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: 1, AccessHash: 1},
			ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: 5}},
		}).ThenResult(&tg.MessagesChannelMessages{Messages: []tg.MessageClass{msg}})

		loc, err := Message(tg.NewClient(m), &tg.InputPeerChannel{ChannelID: 1, AccessHash: 1}, 5, docLoc)(ctx)
		require.NoError(t, err)
		require.Equal(t, &tg.InputDocumentFileLocation{
			ID: 10, AccessHash: 2, FileReference: []byte{2}, ThumbSize: "m",
		}, loc)
	})
	t.Run("ChannelFromMessage", func(t *testing.T) {
		// Min channel seen in a message: channels.getMessages must be used
		// with inputChannelFromMessage, not messages.getMessages.
		chatPeer := &tg.InputPeerChat{ChatID: 7}
		m := tgmock.NewRequire(t)
		m.ExpectCall(&tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannelFromMessage{Peer: chatPeer, MsgID: 3, ChannelID: 1},
			ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: 5}},
		}).ThenResult(&tg.MessagesChannelMessages{Messages: []tg.MessageClass{msg}})

		peer := &tg.InputPeerChannelFromMessage{Peer: chatPeer, MsgID: 3, ChannelID: 1}
		loc, err := Message(tg.NewClient(m), peer, 5, docLoc)(ctx)
		require.NoError(t, err)
		require.Equal(t, []byte{2}, loc.(*tg.InputDocumentFileLocation).FileReference)
	})
	t.Run("User", func(t *testing.T) {
		m := tgmock.NewRequire(t)
		m.ExpectCall(&tg.MessagesGetMessagesRequest{
			ID: []tg.InputMessageClass{&tg.InputMessageID{ID: 5}},
		}).ThenResult(&tg.MessagesMessages{Messages: []tg.MessageClass{msg}})

		loc, err := Message(tg.NewClient(m), &tg.InputPeerUser{UserID: 1}, 5, docLoc)(ctx)
		require.NoError(t, err)
		require.Equal(t, []byte{2}, loc.(*tg.InputDocumentFileLocation).FileReference)
	})
	t.Run("Deleted", func(t *testing.T) {
		m := tgmock.NewRequire(t)
		m.ExpectCall(&tg.MessagesGetMessagesRequest{
			ID: []tg.InputMessageClass{&tg.InputMessageID{ID: 5}},
		}).ThenResult(&tg.MessagesMessages{Messages: []tg.MessageClass{&tg.MessageEmpty{ID: 5}}})

		_, err := Message(tg.NewClient(m), &tg.InputPeerSelf{}, 5, docLoc)(ctx)
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestStory(t *testing.T) {
	ctx := context.Background()
	m := tgmock.NewRequire(t)
	m.ExpectCall(&tg.StoriesGetStoriesByIDRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   []int{3},
	}).ThenResult(&tg.StoriesStories{Stories: []tg.StoryItemClass{
		&tg.StoryItem{ID: 3, Media: &tg.MessageMediaPhoto{
			Photo: &tg.Photo{ID: 7, AccessHash: 2, FileReference: []byte{2}},
		}},
	}})

	loc, err := Story(tg.NewClient(m), &tg.InputPeerSelf{}, 3,
		&tg.InputPhotoFileLocation{ID: 7, FileReference: []byte{1}, ThumbSize: "y"})(ctx)
	require.NoError(t, err)
	require.Equal(t, &tg.InputPhotoFileLocation{
		ID: 7, AccessHash: 2, FileReference: []byte{2}, ThumbSize: "y",
	}, loc)
}

func TestUserPhoto(t *testing.T) {
	ctx := context.Background()
	page := make([]tg.PhotoClass, userPhotosLimit)
	for i := range page {
		page[i] = &tg.Photo{ID: int64(100 + i)}
	}

	m := tgmock.NewRequire(t)
	m.ExpectCall(&tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{}, Offset: 0, Limit: userPhotosLimit,
	}).ThenResult(&tg.PhotosPhotosSlice{Photos: page})
	m.ExpectCall(&tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{}, Offset: userPhotosLimit, Limit: userPhotosLimit,
	}).ThenResult(&tg.PhotosPhotosSlice{Photos: []tg.PhotoClass{
		&tg.Photo{ID: 7, FileReference: []byte{2}},
	}})

	loc, err := UserPhoto(tg.NewClient(m), &tg.InputUserSelf{},
		&tg.InputPhotoFileLocation{ID: 7, FileReference: []byte{1}})(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte{2}, loc.(*tg.InputPhotoFileLocation).FileReference)
}

func TestFromMedia(t *testing.T) {
	page := &tg.MessageMediaWebPage{Webpage: &tg.WebPage{
		Photo:    &tg.Photo{ID: 1, FileReference: []byte{1}},
		Document: &tg.Document{ID: 2, FileReference: []byte{2}},
	}}
	loc, err := FromMedia(&tg.InputPhotoFileLocation{ID: 1}, page)
	require.NoError(t, err)
	require.Equal(t, []byte{1}, loc.(*tg.InputPhotoFileLocation).FileReference)

	loc, err = FromMedia(&tg.InputDocumentFileLocation{ID: 2}, page)
	require.NoError(t, err)
	require.Equal(t, []byte{2}, loc.(*tg.InputDocumentFileLocation).FileReference)

	_, err = FromMedia(&tg.InputPeerPhotoFileLocation{}, page)
	require.Error(t, err)
	_, err = FromMedia(&tg.InputDocumentFileLocation{ID: 3}, page)
	require.ErrorIs(t, err, ErrNotFound)
}
