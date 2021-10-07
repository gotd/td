package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
)

func TestBuilder_Album(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	loc := &tg.InputPhoto{
		ID: 10,
	}

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMultiMediaRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Len(t, req.MultiMedia, 2)
		for i := range req.MultiMedia {
			require.Equal(t, req.MultiMedia[i].Media, &tg.InputMediaPhoto{ID: loc})
			require.NotZero(t, req.MultiMedia[i].RandomID)
		}
	}).ThenResult(&tg.Updates{})
	_, err := sender.Self().Album(ctx, Photo(loc), Photo(loc))
	require.NoError(t, err)

	doc := &tg.InputDocument{
		ID: 10,
	}
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMultiMediaRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Len(t, req.MultiMedia, 2)
		for i := range req.MultiMedia {
			require.Equal(t, req.MultiMedia[i].Media, &tg.InputMediaDocument{ID: doc})
			require.NotZero(t, req.MultiMedia[i].RandomID)
		}
	}).ThenResult(&tg.Updates{})
	_, err = sender.Self().Album(ctx, Document(doc), Document(doc))
	require.NoError(t, err)
}

func TestBuilder_UploadMedia(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	file := &tg.InputFile{
		ID: 10,
	}
	expected := &tg.MessageMediaEmpty{}

	mock.ExpectCall(&tg.MessagesUploadMediaRequest{
		Peer: &tg.InputPeerSelf{},
		Media: &tg.InputMediaUploadedPhoto{
			File: file,
		},
	}).ThenResult(expected)

	r, err := sender.Self().UploadMedia(ctx, UploadedPhoto(file))
	require.NoError(t, err)
	require.Equal(t, expected, r)

	mock.ExpectCall(&tg.MessagesUploadMediaRequest{
		Peer: &tg.InputPeerSelf{},
		Media: &tg.InputMediaUploadedPhoto{
			File: file,
		},
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().UploadMedia(ctx, UploadedPhoto(file))
	require.Error(t, err)
}
