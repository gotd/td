package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message/entity"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

func performTextOptions(media *tg.InputSingleMedia, opts []StyledTextOption) error {
	if len(opts) > 0 {
		tb := entity.Builder{}
		if err := styling.Perform(&tb, opts...); err != nil {
			return err
		}
		media.Message, media.Entities = tb.Complete()
	}

	return nil
}

// Media adds given media attachment to message.
func Media(media tg.InputMediaClass, caption ...StyledTextOption) MediaOption {
	return mediaOptionFunc(func(ctx context.Context, b *multiMediaBuilder) error {
		singleMedia := tg.InputSingleMedia{
			Media: media,
		}
		if err := performTextOptions(&singleMedia, caption); err != nil {
			return err
		}

		b.media = append(b.media, singleMedia)
		return nil
	})
}

// Album sends message with multiple media attachments.
func (b *Builder) Album(ctx context.Context, media MultiMediaOption, album ...MultiMediaOption) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	if len(album) < 1 {
		return b.Media(ctx, media)
	}

	mb, err := b.applyMedia(ctx, p, media, album...)
	if err != nil {
		return nil, err
	}

	upd, err := b.sender.sendMultiMedia(ctx, &tg.MessagesSendMultiMediaRequest{
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		Peer:         p,
		ReplyToMsgID: b.replyToMsgID,
		MultiMedia:   mb,
		ScheduleDate: b.scheduleDate,
	})
	if err != nil {
		return nil, xerrors.Errorf("send album: %w", err)
	}

	return upd, nil
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

// Media sends message with media attachment.
func (b *Builder) Media(ctx context.Context, media MediaOption) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	attachment, err := b.applySingleMedia(ctx, p, media)
	if err != nil {
		return nil, err
	}

	upd, err := b.sender.sendMedia(ctx, &tg.MessagesSendMediaRequest{
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
	})
	if err != nil {
		return nil, xerrors.Errorf("send media: %w", err)
	}

	return upd, nil
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

	r, err := b.sender.uploadMedia(ctx, &tg.MessagesUploadMediaRequest{
		Peer:  p,
		Media: attachment.Media,
	})
	if err != nil {
		return nil, xerrors.Errorf("upload media: %w", err)
	}

	return r, nil
}
