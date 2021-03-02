package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

type multiMediaBuilder struct {
	sender *Sender
	// Attached media.
	media []tg.InputSingleMedia
}

// MediaOption is a option for setting media attachments.
type MediaOption interface {
	apply(ctx context.Context, b *multiMediaBuilder) error
}

// mediaOptionFunc is a function adapter for MediaOption.
type mediaOptionFunc func(ctx context.Context, b *multiMediaBuilder) error

// apply implements MediaOption.
func (m mediaOptionFunc) apply(ctx context.Context, b *multiMediaBuilder) error {
	return m(ctx, b)
}

func performTextOptions(media *tg.InputSingleMedia, opts []StyledTextOption) {
	if len(opts) > 0 {
		captionBuilder := textBuilder{}
		for _, opt := range opts {
			opt(&captionBuilder)
		}

		media.Message, media.Entities = captionBuilder.Complete()
	}
}

// Media adds given media attachment to message.
func Media(media tg.InputMediaClass, caption ...StyledTextOption) MediaOption {
	return mediaOptionFunc(func(ctx context.Context, b *multiMediaBuilder) error {
		singleMedia := tg.InputSingleMedia{
			Media: media,
		}
		performTextOptions(&singleMedia, caption)

		b.media = append(b.media, singleMedia)
		return nil
	})
}

// GeoPoint adds geo point attachment.
// NB: parameter accuracy may be zero and will not be used.
func GeoPoint(lat, long float64, accuracy int, caption ...StyledTextOption) MediaOption {
	return Media(&tg.InputMediaGeoPoint{
		GeoPoint: &tg.InputGeoPoint{
			Lat:            lat,
			Long:           long,
			AccuracyRadius: accuracy,
		},
	}, caption...)
}

// Contact adds contact attachment.
func Contact(contact tg.InputMediaContact, caption ...StyledTextOption) MediaOption {
	return Media(&contact, caption...)
}

func (b *Builder) applyMedia(ctx context.Context, media MediaOption, album ...MediaOption) ([]tg.InputSingleMedia, error) {
	mb := multiMediaBuilder{
		sender: b.sender,
		media:  make([]tg.InputSingleMedia, 0, len(album)+1),
	}

	if err := media.apply(ctx, &mb); err != nil {
		return nil, xerrors.Errorf("media first option: %w", err)
	}

	if len(album) > 0 {
		for i, opt := range album {
			if err := opt.apply(ctx, &mb); err != nil {
				return nil, xerrors.Errorf("media option %d: %w", i+2, err)
			}
		}
	}

	return mb.media, nil
}

func (b *Builder) applySingleMedia(ctx context.Context, media MediaOption) (tg.InputSingleMedia, error) {
	r, err := b.applyMedia(ctx, media)
	if err != nil {
		return tg.InputSingleMedia{}, err
	}

	if len(r) < 1 {
		panic("unreachable: there are should be at least one media attachment")
	}

	return r[0], nil
}

// Media sends message with media attachments.
func (b *Builder) Media(ctx context.Context, media MediaOption, album ...MediaOption) error {
	peer, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	if len(album) > 0 {
		mb, err := b.applyMedia(ctx, media, album...)
		if err != nil {
			return err
		}

		return b.sender.SendMultiMedia(ctx, &tg.MessagesSendMultiMediaRequest{
			Silent:       b.silent,
			Background:   b.background,
			ClearDraft:   b.clearDraft,
			Peer:         peer,
			ReplyToMsgID: b.replyToMsgID,
			MultiMedia:   mb,
			ScheduleDate: b.scheduleDate,
		})
	}

	attachment, err := b.applySingleMedia(ctx, media)
	if err != nil {
		return err
	}

	return b.sender.SendMedia(ctx, &tg.MessagesSendMediaRequest{
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		Peer:         peer,
		ReplyToMsgID: b.replyToMsgID,
		Media:        attachment.Media,
		Message:      attachment.Message,
		ReplyMarkup:  b.replyMarkup,
		Entities:     attachment.Entities,
		ScheduleDate: b.scheduleDate,
	})
}

// UploadMedia uses messages.uploadMedia to upload a file and associate it to
// a chat (without actually sending it to the chat).
//
// See https://core.telegram.org/method/messages.uploadMedia.
func (b *Builder) UploadMedia(ctx context.Context, media MediaOption) (tg.MessageMediaClass, error) {
	peer, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	attachment, err := b.applySingleMedia(ctx, media)
	if err != nil {
		return nil, err
	}

	return b.sender.UploadMedia(ctx, &tg.MessagesUploadMediaRequest{
		Peer:  peer,
		Media: attachment.Media,
	})
}
