package messages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func Test_getDocFilename(t *testing.T) {
	date := time.Now()
	f := date.Format(dateLayout)

	tests := []struct {
		name string
		args *tg.Document
		want string
	}{
		{
			"Doc",
			&tg.Document{
				Date: int(date.Unix()),
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeFilename{FileName: "10.jpg"},
				},
			},
			"10.jpg",
		},
		{
			"Gif",
			&tg.Document{
				Date: int(date.Unix()),
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeAnimated{},
				},
			},
			"doc0_" + f + ".gif",
		},
		{
			"Video",
			&tg.Document{
				Date: int(date.Unix()),
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeVideo{},
				},
			},
			"doc0_" + f + ".mp4",
		},
		{
			"Photo",
			&tg.Document{
				Date: int(date.Unix()),
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeImageSize{},
				},
			},
			"doc0_" + f + ".jpg",
		},
		{
			"Audio",
			&tg.Document{
				Date: int(date.Unix()),
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeAudio{},
				},
			},
			"doc0_" + f + ".mp3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, getDocFilename(tt.args))
		})
	}
}

func TestElem_File(t *testing.T) {
	type results struct {
		file, doc, photo bool
	}
	tests := []struct {
		Name string
		Msg  tg.NotEmptyMessage
		results
	}{
		{"EmptyMessage", &tg.Message{}, results{}},
		{"ServiceMessage", &tg.MessageService{}, results{}},
		{"EmptyPhoto", &tg.Message{
			Media: &tg.MessageMediaPhoto{
				Photo: &tg.PhotoEmpty{},
			},
		}, results{}},
		{"EmptyDoc", &tg.Message{
			Media: &tg.MessageMediaDocument{
				Document: &tg.DocumentEmpty{},
			},
		}, results{}},
		{"Photo", &tg.Message{
			Media: &tg.MessageMediaPhoto{
				Photo: &tg.Photo{
					Sizes: []tg.PhotoSizeClass{
						&tg.PhotoSize{
							Type: "cock",
							W:    10,
							H:    10,
						},
					},
				},
			},
		}, results{file: true, photo: true}},
		{"Document", &tg.Message{
			Media: &tg.MessageMediaDocument{
				Document: &tg.Document{},
			},
		}, results{file: true, doc: true}},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			a := require.New(t)
			var ok bool

			elem := Elem{Msg: test.Msg}
			_, ok = elem.File()
			a.Equal(test.file, ok)
			_, ok = elem.Document()
			a.Equal(test.doc, ok)
			_, ok = elem.Photo()
			a.Equal(test.photo, ok)
		})
	}
}
