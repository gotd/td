package fileid

import (
	"encoding/base64"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

var testData = map[string]string{
	"Sticker":         "CAACAgIAAxkBAAM6YZlDEHCmaTKrUhCIjxAPtPtjVx4AAicAA4dXjx6dGLyHwXVNcCIE",
	"AnimatedSticker": "CAACAgIAAxkBAANCYZzsGL3c2jB4BE46_bD9-aaYH10AApEOAAJZQylKSstCqeiyJ5giBA",
	"GIF":             "CgACAgIAAxkBAAM7YZqVjhoGXOIk6qgVu7xd0QvyRVEAArQQAAK7XrBIi5xgKHPRFpQiBA",
	"GIFThumbnail":    "AAMCAgADGQEAAzthmpWOGgZc4iTqqBW7vF3RC_JFUQACtBAAArtesEiLnGAoc9EWlAEAB20AAyIE",
	"Photo":           "AgACAgIAAxkBAAM9YZqXG-B0WHEv7lFlQxOQDs6jrGQAAoa7MRvdfNlIhJa73cDxR0kBAAMCAAN4AAMiBA",
	"Video":           "BAACAgIAAxkBAANAYZzjSkCVY7Ttrp2l92eCQzYYxVEAAkoRAAJIYKFIRionwJTz4kIiBA",
	"VideoThumbnail":  "AAMCAgADGQEAA0BhnONKQJVjtO2unaX3Z4JDNhjFUQACShEAAkhgoUhGKifAlPPiQgEAB20AAyIE",
	"ChatPhoto":       "AQADAgAD7a8xG75QcEkACAMAA2jAIuIW____cd7THMWjNdIiBA",
	"Voice":           "AwACAgIAAxkBAANDYZzsXw55-6fljCSeQXEP3dX5_egAAlkSAAJStulIAYO3JdIypKQiBA",
	"Audio":           "CQACAgIAAxkBAANEYZzt3rDAw5CkHSU8RZA8AzTTsyMAAvACAAKoAAF4SjhQUd8y3lIoIgQ",
}

var wantData = map[string]FileID{
	"Sticker": {
		Type:          Sticker,
		DC:            2,
		ID:            2202074980139663399,
		AccessHash:    8092253579521038493,
		FileReference: []byte("\x01\x00\x00\x00:a\x99C\x10p\xa6i2\xabR\x10\x88\x8f\x10\x0f\xb4\xfbcW\x1e"),
	},
	"AnimatedSticker": {
		Type:          Sticker,
		DC:            2,
		ID:            5343876482382958225,
		AccessHash:    -7482815543510906038,
		FileReference: []byte("\x01\x00\x00\x00Ba\x9c\xec\x18\xbd\xdc\xda0x\x04N:\xfd\xb0\xfd\xf9\xa6\x98\x1f]"),
	},
	"GIF": {
		Type:          Animation,
		DC:            2,
		ID:            5237790523883786420,
		AccessHash:    -7775797414079718261,
		FileReference: []byte("\x01\x00\x00\x00;a\x9a\x95\x8e\x1a\x06\\\xe2$\xea\xa8\x15\xbb\xbc]\xd1\v\xf2EQ"),
	},
	"GIFThumbnail": {
		Type:          Thumbnail,
		DC:            2,
		ID:            5237790523883786420,
		AccessHash:    -7775797414079718261,
		FileReference: []byte("\x01\x00\x00\x00;a\x9a\x95\x8e\x1a\x06\\\xe2$\xea\xa8\x15\xbb\xbc]\xd1\v\xf2EQ"),
		PhotoSizeSource: PhotoSizeSource{
			Type:          PhotoSizeSourceThumbnail,
			ThumbnailType: 109,
		},
	},
	"Photo": {
		Type:          Photo,
		DC:            2,
		ID:            5249364129762884486,
		AccessHash:    5280454898771269252,
		FileReference: []byte("\x01\x00\x00\x00=a\x9a\x97\x1b\xe0tXq/\xeeQeC\x13\x90\x0eΣ\xacd"),
		PhotoSizeSource: PhotoSizeSource{
			Type:          PhotoSizeSourceThumbnail,
			FileType:      Photo,
			ThumbnailType: 120,
		},
	},
	"Video": {
		Type:          Video,
		DC:            2,
		ID:            5233570104335143242,
		AccessHash:    4819682371444353606,
		FileReference: []byte("\x01\x00\x00\x00@a\x9c\xe3J@\x95c\xb4\xed\xae\x9d\xa5\xf7g\x82C6\x18\xc5Q"),
	},
	"VideoThumbnail": {
		Type:          Thumbnail,
		DC:            2,
		ID:            5233570104335143242,
		AccessHash:    4819682371444353606,
		FileReference: []byte("\x01\x00\x00\x00@a\x9c\xe3J@\x95c\xb4\xed\xae\x9d\xa5\xf7g\x82C6\x18\xc5Q"),
		PhotoSizeSource: PhotoSizeSource{
			Type:          PhotoSizeSourceThumbnail,
			ThumbnailType: 109,
		},
	},
	"ChatPhoto": {
		Type: ProfilePhoto,
		DC:   2,
		ID:   5291818339590582253,
		PhotoSizeSource: PhotoSizeSource{
			Type:             PhotoSizeSourceDialogPhotoBig,
			DialogID:         -1001228418968,
			DialogAccessHash: -3299551084991488399,
		},
	},
	"Voice": {
		Type:          Voice,
		DC:            2,
		ID:            5253930903607972441,
		AccessHash:    -6583080877151517951,
		FileReference: []byte("\x01\x00\x00\x00Ca\x9c\xec_\x0ey\xfb\xa7\xe5\x8c$\x9eAq\x0f\xdd\xd5\xf9\xfd\xe8"),
	},
	"Audio": {
		Type:          Audio,
		DC:            2,
		ID:            5366039677566452464,
		AccessHash:    2905629019683770424,
		FileReference: []byte("\x01\x00\x00\x00Da\x9c\xedް\xc0Ð\xa4\x1d%<E\x90<\x034ӳ#"),
	},
}

func TestDecodeFileID(t *testing.T) {
	type testCase struct {
		name    string
		input   string
		want    FileID
		wantErr bool
	}
	tests := []testCase{
		{
			"Empty",
			"",
			FileID{},
			true,
		},
		{
			"InvalidBase64",
			"/-*-/--+",
			FileID{},
			true,
		},
		{
			"TooSmallLength",
			base64.RawURLEncoding.EncodeToString([]byte{1}),
			FileID{},
			true,
		},
		{
			"UnsupportedVersion",
			base64.RawURLEncoding.EncodeToString([]byte{1, persistentIDVersionOld}),
			FileID{},
			true,
		},
		{
			"UnknownVersion",
			base64.RawURLEncoding.EncodeToString([]byte{1, 20}),
			FileID{},
			true,
		},
	}
	for name, input := range testData {
		expect, ok := wantData[name]
		if !ok {
			t.Fatalf("Update wantData[%q]", name)
		}
		tests = append(tests, testCase{
			name:    name,
			input:   input,
			want:    expect,
			wantErr: false,
		})
	}
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].name < tests[j].name
	})

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
