package fileid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testData = map[string]string{
	"Sticker":      "CAACAgIAAxkBAAM6YZlDEHCmaTKrUhCIjxAPtPtjVx4AAicAA4dXjx6dGLyHwXVNcCIE",
	"GIF":          "CgACAgIAAxkBAAM7YZqVjhoGXOIk6qgVu7xd0QvyRVEAArQQAAK7XrBIi5xgKHPRFpQiBA",
	"GIFThumbnail": "AAMCAgADGQEAAzthmpWOGgZc4iTqqBW7vF3RC_JFUQACtBAAArtesEiLnGAoc9EWlAEAB20AAyIE",
	"Photo":        "AgACAgIAAxkBAAM9YZqXG-B0WHEv7lFlQxOQDs6jrGQAAoa7MRvdfNlIhJa73cDxR0kBAAMCAAN4AAMiBA",
}

func TestDecodeFileID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    FileID
		wantErr bool
	}{
		{
			"Sticker",
			testData["Sticker"],
			FileID{
				Type:          Sticker,
				DC:            2,
				ID:            2202074980139663399,
				AccessHash:    8092253579521038493,
				FileReference: "\x01\x00\x00\x00:a\x99C\x10p\xa6i2\xabR\x10\x88\x8f\x10\x0f\xb4\xfbcW\x1e",
			},
			false,
		},
		{
			"GIF",
			testData["GIF"],
			FileID{
				Type:          Animation,
				DC:            2,
				ID:            5237790523883786420,
				AccessHash:    -7775797414079718261,
				FileReference: "\x01\x00\x00\x00;a\x9a\x95\x8e\x1a\x06\\\xe2$\xea\xa8\x15\xbb\xbc]\xd1\v\xf2EQ",
			},
			false,
		},
		{
			"GIFThumbnail",
			testData["GIFThumbnail"],
			FileID{
				Type:          Thumbnail,
				DC:            2,
				ID:            5237790523883786420,
				AccessHash:    -7775797414079718261,
				FileReference: "\x01\x00\x00\x00;a\x9a\x95\x8e\x1a\x06\\\xe2$\xea\xa8\x15\xbb\xbc]\xd1\v\xf2EQ",
			},
			false,
		},
		{
			"Photo",
			testData["Photo"],
			FileID{
				Type:          Photo,
				DC:            2,
				ID:            5249364129762884486,
				AccessHash:    5280454898771269252,
				FileReference: "\x01\x00\x00\x00=a\x9a\x97\x1b\xe0tXq/\xeeQeC\x13\x90\x0eΣ\xacd",
				PhotoSize: PhotoSize{
					FileType: 0x2,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)

			got, err := DecodeFileID(tt.input)
			if tt.wantErr {
				a.Error(err)
			} else {
				a.NoError(err)
				a.Equal(tt.want, got)
			}
		})
	}
}

