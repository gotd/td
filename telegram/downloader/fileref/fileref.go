// Package fileref contains helpers to refresh file references of file
// locations for downloader.Builder.WithRefresh.
//
// File references expire, and a download fails with FILE_REFERENCE_EXPIRED.
// The fresh reference is obtained by fetching the object that contains the
// file again (message, story, profile photo), which is what these helpers do.
//
// See https://core.telegram.org/api/file_reference.
package fileref

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// API is a subset of tg.Client used by refresh helpers.
type API interface {
	MessagesGetMessages(ctx context.Context, id []tg.InputMessageClass) (tg.MessagesMessagesClass, error)
	ChannelsGetMessages(ctx context.Context, request *tg.ChannelsGetMessagesRequest) (tg.MessagesMessagesClass, error)
	StoriesGetStoriesByID(ctx context.Context, request *tg.StoriesGetStoriesByIDRequest) (*tg.StoriesStories, error)
	PhotosGetUserPhotos(ctx context.Context, request *tg.PhotosGetUserPhotosRequest) (tg.PhotosPhotosClass, error)
}

var _ API = (*tg.Client)(nil)

// ErrNotFound means that the object or the file is not found, e.g. the
// message was deleted or its media changed.
var ErrNotFound = errors.New("file not found")

// Message returns refresh function for a file of message msgID in peer:
// document or photo of its media, including alternative documents, video
// cover and web page preview.
func Message(api API, peer tg.InputPeerClass, msgID int, loc tg.InputFileLocationClass) downloader.RefreshFunc {
	return func(ctx context.Context) (tg.InputFileLocationClass, error) {
		ids := []tg.InputMessageClass{&tg.InputMessageID{ID: msgID}}
		var (
			res tg.MessagesMessagesClass
			err error
		)
		if ch, ok := inputChannel(peer); ok {
			res, err = api.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
				Channel: ch,
				ID:      ids,
			})
		} else {
			res, err = api.MessagesGetMessages(ctx, ids)
		}
		if err != nil {
			return nil, errors.Wrap(err, "get message")
		}
		modified, ok := res.AsModified()
		if !ok {
			return nil, errors.Errorf("unexpected result %T", res)
		}

		var media []tg.MessageMediaClass
		for _, m := range modified.GetMessages() {
			msg, ok := m.(*tg.Message)
			if !ok || msg.ID != msgID {
				continue
			}
			if v, ok := msg.GetMedia(); ok {
				media = append(media, v)
			}
		}
		return FromMedia(loc, media...)
	}
}

// inputChannel returns input channel of peer, if peer is a channel.
func inputChannel(peer tg.InputPeerClass) (tg.InputChannelClass, bool) {
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		return &tg.InputChannel{ChannelID: p.ChannelID, AccessHash: p.AccessHash}, true
	case *tg.InputPeerChannelFromMessage:
		return &tg.InputChannelFromMessage{Peer: p.Peer, MsgID: p.MsgID, ChannelID: p.ChannelID}, true
	default:
		return nil, false
	}
}

// Story returns refresh function for a file of story storyID of peer.
func Story(api API, peer tg.InputPeerClass, storyID int, loc tg.InputFileLocationClass) downloader.RefreshFunc {
	return func(ctx context.Context) (tg.InputFileLocationClass, error) {
		res, err := api.StoriesGetStoriesByID(ctx, &tg.StoriesGetStoriesByIDRequest{
			Peer: peer,
			ID:   []int{storyID},
		})
		if err != nil {
			return nil, errors.Wrap(err, "get story")
		}

		var media []tg.MessageMediaClass
		for _, s := range res.Stories {
			if story, ok := s.(*tg.StoryItem); ok && story.ID == storyID {
				media = append(media, story.Media)
			}
		}
		return FromMedia(loc, media...)
	}
}

// userPhotosLimit is the page size of photos.getUserPhotos.
const userPhotosLimit = 100

// UserPhoto returns refresh function for a profile photo of user, from
// photos.getUserPhotos.
func UserPhoto(api API, user tg.InputUserClass, loc tg.InputFileLocationClass) downloader.RefreshFunc {
	return func(ctx context.Context) (tg.InputFileLocationClass, error) {
		for offset := 0; ; {
			res, err := api.PhotosGetUserPhotos(ctx, &tg.PhotosGetUserPhotosRequest{
				UserID: user,
				Offset: offset,
				Limit:  userPhotosLimit,
			})
			if err != nil {
				return nil, errors.Wrap(err, "get user photos")
			}
			photos := res.GetPhotos()
			updated, err := FromPhotos(loc, photos...)
			if err == nil || !errors.Is(err, ErrNotFound) {
				return updated, err
			}
			if len(photos) < userPhotosLimit {
				return nil, ErrNotFound
			}
			offset += len(photos)
		}
	}
}

// FromMedia returns copy of loc with file reference of the matching document
// or photo (by ID) from media.
func FromMedia(loc tg.InputFileLocationClass, media ...tg.MessageMediaClass) (tg.InputFileLocationClass, error) {
	var (
		docs   []tg.DocumentClass
		photos []tg.PhotoClass
	)
	for _, m := range media {
		switch m := m.(type) {
		case *tg.MessageMediaDocument:
			docs = append(docs, m.Document)
			docs = append(docs, m.AltDocuments...)
			photos = append(photos, m.VideoCover)
		case *tg.MessageMediaPhoto:
			photos = append(photos, m.Photo)
		case *tg.MessageMediaWebPage:
			if page, ok := m.Webpage.(*tg.WebPage); ok {
				docs = append(docs, page.Document)
				photos = append(photos, page.Photo)
			}
		}
	}

	switch loc.(type) {
	case *tg.InputDocumentFileLocation:
		return FromDocuments(loc, docs...)
	case *tg.InputPhotoFileLocation:
		return FromPhotos(loc, photos...)
	default:
		return nil, errors.Errorf("unsupported location %T", loc)
	}
}

// FromDocuments returns copy of document loc with file reference of the
// document with the same ID.
func FromDocuments(loc tg.InputFileLocationClass, docs ...tg.DocumentClass) (tg.InputFileLocationClass, error) {
	l, ok := loc.(*tg.InputDocumentFileLocation)
	if !ok {
		return nil, errors.Errorf("unsupported location %T", loc)
	}
	for _, d := range docs {
		doc, ok := d.(*tg.Document)
		if !ok || doc.ID != l.ID {
			continue
		}
		updated := *l
		updated.AccessHash = doc.AccessHash
		updated.FileReference = doc.FileReference
		return &updated, nil
	}
	return nil, ErrNotFound
}

// FromPhotos returns copy of photo loc with file reference of the photo with
// the same ID.
func FromPhotos(loc tg.InputFileLocationClass, photos ...tg.PhotoClass) (tg.InputFileLocationClass, error) {
	l, ok := loc.(*tg.InputPhotoFileLocation)
	if !ok {
		return nil, errors.Errorf("unsupported location %T", loc)
	}
	for _, p := range photos {
		photo, ok := p.(*tg.Photo)
		if !ok || photo.ID != l.ID {
			continue
		}
		updated := *l
		updated.AccessHash = photo.AccessHash
		updated.FileReference = photo.FileReference
		return &updated, nil
	}
	return nil, ErrNotFound
}
