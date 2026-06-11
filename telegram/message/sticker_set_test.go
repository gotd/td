package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func testStickers() []tg.DocumentClass {
	return []tg.DocumentClass{
		&tg.DocumentEmpty{ID: 1},
		&tg.Document{
			ID:            10,
			AccessHash:    20,
			FileReference: []byte{1, 2, 3},
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeSticker{Alt: "😎", Stickerset: &tg.InputStickerSetEmpty{}},
			},
		},
		&tg.Document{
			ID:            11,
			AccessHash:    21,
			FileReference: []byte{4, 5, 6},
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeSticker{Alt: "😢", Stickerset: &tg.InputStickerSetEmpty{}},
			},
		},
	}
}

func TestStickerByIndex(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesGetRecentStickersRequest{}).ThenResult(&tg.MessagesRecentStickers{
		Stickers: testStickers(),
	})
	expectSendMedia(t, &tg.InputMediaDocument{ID: &tg.InputDocument{
		ID:            10,
		AccessHash:    20,
		FileReference: []byte{1, 2, 3},
	}}, mock)

	_, err := sender.Self().Sticker(RecentStickers()).ByIndex(ctx, 1)
	require.NoError(t, err)
}

func TestStickerByIndexOutOfRange(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesGetFavedStickersRequest{}).ThenResult(&tg.MessagesFavedStickers{
		Stickers: testStickers(),
	})

	_, err := sender.Self().Sticker(FavedStickers()).ByIndex(ctx, 10)
	require.Error(t, err)
}

func TestStickerByEmoji(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesGetFavedStickersRequest{}).ThenResult(&tg.MessagesFavedStickers{
		Stickers: testStickers(),
	})
	expectSendMedia(t, &tg.InputMediaDocument{ID: &tg.InputDocument{
		ID:            11,
		AccessHash:    21,
		FileReference: []byte{4, 5, 6},
	}}, mock)

	_, err := sender.Self().Sticker(FavedStickers()).ByEmoji(ctx, "😢")
	require.NoError(t, err)
}

func TestStickerByEmojiNotFound(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesGetFavedStickersRequest{}).ThenResult(&tg.MessagesFavedStickers{
		Stickers: testStickers(),
	})

	_, err := sender.Self().Sticker(FavedStickers()).ByEmoji(ctx, "🤡")
	require.Error(t, err)
}

func TestStickerSetFirst(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesGetStickerSetRequest{
		Stickerset: &tg.InputStickerSetShortName{ShortName: "gotd"},
	}).ThenResult(&tg.MessagesStickerSet{
		Documents: testStickers()[1:],
	})
	expectSendMedia(t, &tg.InputMediaDocument{ID: &tg.InputDocument{
		ID:            10,
		AccessHash:    20,
		FileReference: []byte{1, 2, 3},
	}}, mock)

	_, err := sender.Self().Sticker(StickerSetName("gotd")).First(ctx)
	require.NoError(t, err)
}

func TestStickerNotModified(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesGetFavedStickersRequest{}).
		ThenResult(&tg.MessagesFavedStickersNotModified{})

	_, err := sender.Self().Sticker(FavedStickers()).First(ctx)
	require.Error(t, err)
}
