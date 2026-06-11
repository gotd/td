package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// StickerSource resolves a list of stickers (documents) to pick from.
//
// Use one of the built-in sources: FavedStickers, RecentStickers,
// AttachedRecentStickers or StickerSet.
type StickerSource interface {
	// Stickers returns stickers (documents) of the source.
	Stickers(ctx context.Context, raw *tg.Client) ([]tg.DocumentClass, error)
}

type favedStickerSource struct{}

func (favedStickerSource) Stickers(ctx context.Context, raw *tg.Client) ([]tg.DocumentClass, error) {
	r, err := raw.MessagesGetFavedStickers(ctx, 0)
	if err != nil {
		return nil, errors.Wrap(err, "get faved stickers")
	}

	m, ok := r.AsModified()
	if !ok {
		return nil, errors.Errorf("unexpected type %T", r)
	}
	return m.Stickers, nil
}

// FavedStickers returns a StickerSource using the user's favorite stickers.
//
// See https://core.telegram.org/method/messages.getFavedStickers.
func FavedStickers() StickerSource {
	return favedStickerSource{}
}

type recentStickerSource struct {
	attached bool
}

func (s recentStickerSource) Stickers(ctx context.Context, raw *tg.Client) ([]tg.DocumentClass, error) {
	r, err := raw.MessagesGetRecentStickers(ctx, &tg.MessagesGetRecentStickersRequest{
		Attached: s.attached,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get recent stickers")
	}

	m, ok := r.AsModified()
	if !ok {
		return nil, errors.Errorf("unexpected type %T", r)
	}
	return m.Stickers, nil
}

// RecentStickers returns a StickerSource using the user's recently used stickers.
//
// See https://core.telegram.org/method/messages.getRecentStickers.
func RecentStickers() StickerSource {
	return recentStickerSource{}
}

// AttachedRecentStickers returns a StickerSource using stickers recently
// attached to photo or video files.
//
// See https://core.telegram.org/method/messages.getRecentStickers.
func AttachedRecentStickers() StickerSource {
	return recentStickerSource{attached: true}
}

type stickerSetSource struct {
	set tg.InputStickerSetClass
}

func (s stickerSetSource) Stickers(ctx context.Context, raw *tg.Client) ([]tg.DocumentClass, error) {
	r, err := raw.MessagesGetStickerSet(ctx, &tg.MessagesGetStickerSetRequest{
		Stickerset: s.set,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get sticker set")
	}

	m, ok := r.AsModified()
	if !ok {
		return nil, errors.Errorf("unexpected type %T", r)
	}
	return m.Documents, nil
}

// StickerSet returns a StickerSource using stickers of the given set.
//
// See https://core.telegram.org/method/messages.getStickerSet.
func StickerSet(set tg.InputStickerSetClass) StickerSource {
	return stickerSetSource{set: set}
}

// StickerSetName returns a StickerSource using stickers of the set with given
// short name.
//
// See https://core.telegram.org/method/messages.getStickerSet.
func StickerSetName(shortName string) StickerSource {
	return StickerSet(&tg.InputStickerSetShortName{ShortName: shortName})
}

// StickerSetBuilder selects and sends a sticker from a StickerSource.
//
// It is created using the Builder.Sticker method.
type StickerSetBuilder struct {
	builder *Builder
	source  StickerSource
}

// Sticker creates a StickerSetBuilder to pick and send a sticker from the
// given source.
//
//	sender.Self().Sticker(message.FavedStickers()).ByIndex(ctx, 0)
//	sender.Self().Sticker(message.RecentStickers()).ByEmoji(ctx, "😎")
func (b *Builder) Sticker(source StickerSource) *StickerSetBuilder {
	return &StickerSetBuilder{
		builder: b,
		source:  source,
	}
}

// Stickers returns all stickers (documents) of the source.
func (s *StickerSetBuilder) Stickers(ctx context.Context) ([]tg.DocumentClass, error) {
	return s.source.Stickers(ctx, s.builder.sender.raw)
}

func (s *StickerSetBuilder) send(
	ctx context.Context,
	doc tg.DocumentClass, caption []StyledTextOption,
) (tg.UpdatesClass, error) {
	d, ok := doc.AsNotEmpty()
	if !ok {
		return nil, errors.New("sticker document is empty")
	}
	return s.builder.Media(ctx, Document(d, caption...))
}

// ByIndex sends the sticker at the given index in the source.
func (s *StickerSetBuilder) ByIndex(
	ctx context.Context,
	index int, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	stickers, err := s.Stickers(ctx)
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= len(stickers) {
		return nil, errors.Errorf("sticker index %d out of range [0, %d)", index, len(stickers))
	}
	return s.send(ctx, stickers[index], caption)
}

// ByEmoji sends the first sticker of the source whose alternative emoji
// representation matches the given emoji.
func (s *StickerSetBuilder) ByEmoji(
	ctx context.Context,
	emoji string, caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	stickers, err := s.Stickers(ctx)
	if err != nil {
		return nil, err
	}

	for _, sticker := range stickers {
		doc, ok := sticker.AsNotEmpty()
		if !ok {
			continue
		}
		for _, attr := range doc.Attributes {
			attr, ok := attr.(*tg.DocumentAttributeSticker)
			if ok && attr.Alt == emoji {
				return s.send(ctx, sticker, caption)
			}
		}
	}

	return nil, errors.Errorf("no sticker with emoji %q found", emoji)
}

// First sends the first sticker of the source.
func (s *StickerSetBuilder) First(
	ctx context.Context,
	caption ...StyledTextOption,
) (tg.UpdatesClass, error) {
	return s.ByIndex(ctx, 0, caption...)
}
