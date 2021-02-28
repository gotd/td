package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

type multiMediaBuilder struct {
	msg *Builder
}

// MediaOption is a option for setting media attachments.
type MediaOption interface {
	apply(ctx context.Context, b multiMediaBuilder) error
}

// MediaOptionFunc is a function adapter for MediaOption.
type MediaOptionFunc func(ctx context.Context, b multiMediaBuilder) error

// apply implements MediaOption.
func (m MediaOptionFunc) apply(ctx context.Context, b multiMediaBuilder) error {
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
	return MediaOptionFunc(func(ctx context.Context, b multiMediaBuilder) error {
		singleMedia := tg.InputSingleMedia{
			Media: media,
		}
		performTextOptions(&singleMedia, caption)

		b.msg.media = append(b.msg.media, singleMedia)
		return nil
	})
}

// GeoPoint adds geo point attachment.
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

// Media sends message with media attachments.
func (b *Builder) Media(ctx context.Context, media MediaOption, album ...MediaOption) error {
	peer, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}
	mb := multiMediaBuilder{msg: b}

	if err := media.apply(ctx, mb); err != nil {
		return xerrors.Errorf("media option: %w", err)
	}

	if len(album) > 0 {
		for i, opt := range album {
			if err := opt.apply(ctx, mb); err != nil {
				return xerrors.Errorf("media option %d: %w", i, err)
			}
		}

		return b.sender.SendMultiMedia(ctx, &tg.MessagesSendMultiMediaRequest{
			Silent:       b.silent,
			Background:   b.background,
			ClearDraft:   b.clearDraft,
			Peer:         peer,
			ReplyToMsgID: b.replyToMsgID,
			MultiMedia:   b.media,
			ScheduleDate: b.scheduleDate,
		})
	}

	if len(b.media) < 1 {
		panic("unreachable: there are should be at least one media attachment")
	}

	attachment := b.media[0]
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
