package message

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestBuilder_Album(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	loc := &tg.InputPhoto{
		ID: 10,
	}

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMultiMediaRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Len(req.MultiMedia, 2)
		for i := range req.MultiMedia {
			mock.Equal(req.MultiMedia[i].Media, &tg.InputMediaPhoto{ID: loc})
			mock.NotZero(req.MultiMedia[i].RandomID)
		}
	}).ThenResult(&tg.Updates{})
	_, err := sender.Self().Album(ctx, Photo(loc), Photo(loc))
	mock.NoError(err)

	doc := &tg.InputDocument{
		ID: 10,
	}
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMultiMediaRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Len(req.MultiMedia, 2)
		for i := range req.MultiMedia {
			mock.Equal(req.MultiMedia[i].Media, &tg.InputMediaDocument{ID: doc})
			mock.NotZero(req.MultiMedia[i].RandomID)
		}
	}).ThenResult(&tg.Updates{})
	_, err = sender.Self().Album(ctx, Document(doc), Document(doc))
	mock.NoError(err)
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
	mock.NoError(err)
	mock.Equal(expected, r)

	mock.ExpectCall(&tg.MessagesUploadMediaRequest{
		Peer: &tg.InputPeerSelf{},
		Media: &tg.InputMediaUploadedPhoto{
			File: file,
		},
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().UploadMedia(ctx, UploadedPhoto(file))
	mock.Error(err)
}
