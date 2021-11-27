package fileid

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestFromDocument(t *testing.T) {
	doc := func(attrs ...tg.DocumentAttributeClass) *tg.Document {
		return &tg.Document{
			ID:            1,
			AccessHash:    2,
			FileReference: []byte{3},
			DCID:          4,
			Attributes:    attrs,
		}
	}
	fileID := func(typ Type) FileID {
		return FileID{
			Type:          typ,
			ID:            1,
			AccessHash:    2,
			FileReference: []byte{3},
			DC:            4,
		}
	}

	tests := []struct {
		name string
		doc  *tg.Document
		want FileID
	}{
		{"File", doc(), fileID(DocumentAsFile)},
		{"Animation", doc(&tg.DocumentAttributeAnimated{}), fileID(Animation)},
		{"Sticker", doc(&tg.DocumentAttributeSticker{}), fileID(Sticker)},
		{"Video", doc(&tg.DocumentAttributeVideo{}), fileID(Video)},
		{"VideoNote", doc(&tg.DocumentAttributeVideo{RoundMessage: true}), fileID(VideoNote)},
		{"Audio", doc(&tg.DocumentAttributeAudio{}), fileID(Audio)},
		{"Voice", doc(&tg.DocumentAttributeAudio{Voice: true}), fileID(Voice)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, FromDocument(tt.doc))
		})
	}
}

func TestFromPhoto(t *testing.T) {
	tests := []struct {
		name  string
		photo *tg.Photo
		size  rune
		want  FileID
	}{
		{
			"Photo",
			&tg.Photo{
				ID:            1,
				AccessHash:    2,
				FileReference: []byte{3},
				DCID:          4,
			},
			'x',
			FileID{
				Type:          Photo,
				ID:            1,
				AccessHash:    2,
				FileReference: []byte{3},
				DC:            4,
				PhotoSizeSource: PhotoSizeSource{
					Type:          PhotoSizeSourceThumbnail,
					FileType:      Photo,
					ThumbnailType: 'x',
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, FromPhoto(tt.photo, tt.size))
		})
	}
}
