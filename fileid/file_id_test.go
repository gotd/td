package fileid

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestFileID_AsInputFileLocation(t *testing.T) {
	type testCase struct {
		name   string
		fileID FileID
		want   tg.InputFileLocationClass
		wantOk bool
	}
	tests := []testCase{
		{
			"Sticker",
			wantData["Sticker"],
			&tg.InputDocumentFileLocation{
				ID:            2202074980139663399,
				AccessHash:    8092253579521038493,
				FileReference: []byte("\x01\x00\x00\x00:a\x99C\x10p\xa6i2\xabR\x10\x88\x8f\x10\x0f\xb4\xfbcW\x1e"),
			},
			true,
		},
		{
			"AnimatedSticker",
			wantData["AnimatedSticker"],
			&tg.InputDocumentFileLocation{
				ID:            5343876482382958225,
				AccessHash:    -7482815543510906038,
				FileReference: []byte("\x01\x00\x00\x00Ba\x9c\xec\x18\xbd\xdc\xda0x\x04N:\xfd\xb0\xfd\xf9\xa6\x98\x1f]"),
			},
			true,
		},
		{
			"GIF",
			wantData["GIF"],
			&tg.InputDocumentFileLocation{
				ID:            5237790523883786420,
				AccessHash:    -7775797414079718261,
				FileReference: []byte("\x01\x00\x00\x00;a\x9a\x95\x8e\x1a\x06\\\xe2$\xea\xa8\x15\xbb\xbc]\xd1\v\xf2EQ"),
			},
			true,
		},
		{
			"GIFThumbnail",
			wantData["GIFThumbnail"],
			&tg.InputPhotoFileLocation{
				ID:            5237790523883786420,
				AccessHash:    -7775797414079718261,
				FileReference: []byte("\x01\x00\x00\x00;a\x9a\x95\x8e\x1a\x06\\\xe2$\xea\xa8\x15\xbb\xbc]\xd1\v\xf2EQ"),
				ThumbSize:     "m",
			},
			true,
		},
		{
			"Photo",
			wantData["Photo"],
			&tg.InputPhotoFileLocation{
				ID:            5249364129762884486,
				AccessHash:    5280454898771269252,
				FileReference: []byte("\x01\x00\x00\x00=a\x9a\x97\x1b\xe0tXq/\xeeQeC\x13\x90\x0eΣ\xacd"),
				ThumbSize:     "x",
			},
			true,
		},
		{
			"Video",
			wantData["Video"],
			&tg.InputDocumentFileLocation{
				ID:            5233570104335143242,
				AccessHash:    4819682371444353606,
				FileReference: []byte("\x01\x00\x00\x00@a\x9c\xe3J@\x95c\xb4\xed\xae\x9d\xa5\xf7g\x82C6\x18\xc5Q"),
			},
			true,
		},
		{
			"VideoThumbnail",
			wantData["VideoThumbnail"],
			&tg.InputPhotoFileLocation{
				ID:            5233570104335143242,
				AccessHash:    4819682371444353606,
				FileReference: []byte("\x01\x00\x00\x00@a\x9c\xe3J@\x95c\xb4\xed\xae\x9d\xa5\xf7g\x82C6\x18\xc5Q"),
				ThumbSize:     "m",
			},
			true,
		},
		{
			"ChatPhoto",
			wantData["ChatPhoto"],
			&tg.InputPeerPhotoFileLocation{
				Big: true,
				Peer: &tg.InputPeerChannel{
					ChannelID:  -1001228418968,
					AccessHash: -3299551084991488399,
				},
				PhotoID: 5291818339590582253,
			},
			true,
		},
		{
			"Voice",
			wantData["Voice"],
			&tg.InputDocumentFileLocation{
				ID:            5253930903607972441,
				AccessHash:    -6583080877151517951,
				FileReference: []byte("\x01\x00\x00\x00Ca\x9c\xec_\x0ey\xfb\xa7\xe5\x8c$\x9eAq\x0f\xdd\xd5\xf9\xfd\xe8"),
			},
			true,
		},
		{
			"Audio",
			wantData["Audio"],
			&tg.InputDocumentFileLocation{
				ID:            5366039677566452464,
				AccessHash:    2905629019683770424,
				FileReference: []byte("\x01\x00\x00\x00Da\x9c\xedް\xc0Ð\xa4\x1d%<E\x90<\x034ӳ#"),
			},
			true,
		},
		{
			"Temp",
			FileID{Type: Temp},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			got, ok := tt.fileID.AsInputFileLocation()
			a.Equal(tt.wantOk, ok)
			a.Equal(tt.want, got)
		})
	}
}
