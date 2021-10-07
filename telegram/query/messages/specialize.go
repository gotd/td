package messages

import (
	"fmt"
	"time"

	"github.com/nnqq/td/tg"
)

// Document returns document object if message has a document attachment (video, voice, audio,
// basically every type except photo).
func (e Elem) Document() (*tg.Document, bool) {
	msg, ok := e.Msg.(*tg.Message)
	if !ok {
		return nil, false
	}

	media, ok := msg.Media.(*tg.MessageMediaDocument)
	if !ok {
		return nil, false
	}

	return media.Document.AsNotEmpty()
}

// Photo returns photo object if message has a photo attachment.
func (e Elem) Photo() (*tg.Photo, bool) {
	msg, ok := e.Msg.(*tg.Message)
	if !ok {
		return nil, false
	}

	media, ok := msg.Media.(*tg.MessageMediaPhoto)
	if !ok {
		return nil, false
	}

	return media.Photo.AsNotEmpty()
}

// File represents file attachment.
type File struct {
	Name     string
	MIMEType string
	Location tg.InputFileLocationClass
}

const dateLayout = "2006-01-02_15-04-05"

func getDocFilename(doc *tg.Document) string {
	var filename, ext string
	for _, attr := range doc.Attributes {
		switch v := attr.(type) {
		case *tg.DocumentAttributeImageSize:
			switch doc.MimeType {
			case "image/png":
				ext = ".png"
			case "image/webp":
				ext = ".webp"
			case "image/tiff":
				ext = ".tif"
			default:
				ext = ".jpg"
			}
		case *tg.DocumentAttributeAnimated:
			ext = ".gif"
		case *tg.DocumentAttributeSticker:
			ext = ".webp"
		case *tg.DocumentAttributeVideo:
			switch doc.MimeType {
			case "video/mpeg":
				ext = ".mpeg"
			case "video/webm":
				ext = ".webm"
			case "video/ogg":
				ext = ".ogg"
			default:
				ext = ".mp4"
			}
		case *tg.DocumentAttributeAudio:
			switch doc.MimeType {
			case "audio/webm":
				ext = ".webm"
			case "audio/aac":
				ext = ".aac"
			case "audio/ogg":
				ext = ".ogg"
			default:
				ext = ".mp3"
			}
		case *tg.DocumentAttributeFilename:
			filename = v.FileName
		}
	}

	if filename == "" {
		filename = fmt.Sprintf(
			"doc%d_%s%s", doc.GetID(),
			time.Unix(int64(doc.Date), 0).Format(dateLayout),
			ext,
		)
	}

	return filename
}

type sizedPhoto interface {
	GetW() int
	GetH() int
	GetType() string
}

var (
	_ sizedPhoto = (*tg.PhotoSize)(nil)
	_ sizedPhoto = (*tg.PhotoCachedSize)(nil)
	_ sizedPhoto = (*tg.PhotoSizeProgressive)(nil)
)

// File returns file location if message has a file attachment.
func (e Elem) File() (File, bool) {
	msg, ok := e.Msg.(*tg.Message)
	if !ok {
		return File{}, false
	}

	switch media := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := media.Photo.AsNotEmpty()
		if !ok {
			return File{}, false
		}

		filename := fmt.Sprintf(
			"photo%d_%s.jpg", photo.GetID(),
			time.Unix(int64(photo.Date), 0).Format(dateLayout),
		)

		var (
			thumbSize  string
			maxW, maxH int
		)
		for _, g := range photo.Sizes {
			// TODO(tdakkota): add helpers to choose photo size.
			if sz, ok := g.(sizedPhoto); ok && maxW < sz.GetW() && maxH < sz.GetH() {
				thumbSize = sz.GetType()
			}
		}

		if thumbSize == "" {
			return File{}, false
		}

		return File{
			Name:     filename,
			MIMEType: "image/jpeg",
			Location: &tg.InputPhotoFileLocation{
				ID:            photo.ID,
				AccessHash:    photo.AccessHash,
				FileReference: photo.FileReference,
				ThumbSize:     thumbSize,
			},
		}, true
	case *tg.MessageMediaDocument:
		doc, ok := media.Document.AsNotEmpty()
		if !ok {
			return File{}, false
		}

		return File{
			Name:     getDocFilename(doc),
			MIMEType: doc.MimeType,
			Location: doc.AsInputDocumentFileLocation(),
		}, true
	default:
		return File{}, false
	}
}
