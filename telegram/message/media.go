package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

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

func (b *Builder) applyMedia(
	ctx context.Context,
	p tg.InputPeerClass,
	media MultiMediaOption, album ...MultiMediaOption,
) ([]tg.InputSingleMedia, error) {
	mb := multiMediaBuilder{
		sender: b.sender,
		peer:   p,
		media:  make([]tg.InputSingleMedia, 0, len(album)+1),
	}

	if err := media.applyMulti(ctx, &mb); err != nil {
		return nil, xerrors.Errorf("media first option: %w", err)
	}

	if len(album) > 0 {
		for i, opt := range album {
			if err := opt.applyMulti(ctx, &mb); err != nil {
				return nil, xerrors.Errorf("media option %d: %w", i+2, err)
			}
		}
	}

	return mb.media, nil
}

func (b *Builder) applySingleMedia(
	ctx context.Context,
	p tg.InputPeerClass,
	media MediaOption,
) (tg.InputSingleMedia, error) {
	mb := multiMediaBuilder{
		sender: b.sender,
		peer:   p,
		media:  make([]tg.InputSingleMedia, 0, 1),
	}

	if err := media.apply(ctx, &mb); err != nil {
		return tg.InputSingleMedia{}, xerrors.Errorf("media first option: %w", err)
	}

	return mb.media[0], nil
}

// Album sends message with multiple media attachments.
func (b *Builder) Album(ctx context.Context, media MultiMediaOption, album ...MultiMediaOption) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	if len(album) < 1 {
		return b.Media(ctx, media)
	}

	mb, err := b.applyMedia(ctx, p, media, album...)
	if err != nil {
		return err
	}

	if err := b.sender.sendMultiMedia(ctx, &tg.MessagesSendMultiMediaRequest{
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		Peer:         p,
		ReplyToMsgID: b.replyToMsgID,
		MultiMedia:   mb,
		ScheduleDate: b.scheduleDate,
	}); err != nil {
		return xerrors.Errorf("send album: %w", err)
	}

	return nil
}

// Media sends message with media attachment.
func (b *Builder) Media(ctx context.Context, media MediaOption) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	attachment, err := b.applySingleMedia(ctx, p, media)
	if err != nil {
		return err
	}

	if err := b.sender.sendMedia(ctx, &tg.MessagesSendMediaRequest{
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		Peer:         p,
		ReplyToMsgID: b.replyToMsgID,
		Media:        attachment.Media,
		Message:      attachment.Message,
		ReplyMarkup:  b.replyMarkup,
		Entities:     attachment.Entities,
		ScheduleDate: b.scheduleDate,
	}); err != nil {
		return xerrors.Errorf("send media: %w", err)
	}

	return nil
}

// UploadMedia uses messages.uploadMedia to upload a file and associate it to
// a chat (without actually sending it to the chat).
//
// See https://core.telegram.org/method/messages.uploadMedia.
func (b *Builder) UploadMedia(ctx context.Context, media MediaOption) (tg.MessageMediaClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	attachment, err := b.applySingleMedia(ctx, p, media)
	if err != nil {
		return nil, err
	}

	return b.sender.uploadMedia(ctx, &tg.MessagesUploadMediaRequest{
		Peer:  p,
		Media: attachment.Media,
	})
}
