package fileid

import "github.com/gotd/td/tg"

// FromDocument creates FileID from tg.Document.
func FromDocument(doc *tg.Document) FileID {
	fileID := FileID{
		Type:          DocumentAsFile,
		DC:            doc.DCID,
		ID:            doc.ID,
		AccessHash:    doc.AccessHash,
		FileReference: doc.FileReference,
	}
	for _, attr := range doc.Attributes {
		switch attr := attr.(type) {
		case *tg.DocumentAttributeAnimated:
			fileID.Type = Animation
		case *tg.DocumentAttributeSticker:
			fileID.Type = Sticker
		case *tg.DocumentAttributeVideo:
			fileID.Type = Video
			if attr.RoundMessage {
				fileID.Type = VideoNote
			}
		case *tg.DocumentAttributeAudio:
			fileID.Type = Audio
			if attr.Voice {
				fileID.Type = Voice
			}
		}
	}
	return fileID
}

// FromPhoto creates FileID from tg.Photo.
func FromPhoto(photo *tg.Photo, thumbType rune) FileID {
	return FileID{
		Type:          Photo,
		DC:            photo.DCID,
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,
		PhotoSizeSource: PhotoSizeSource{
			Type:          PhotoSizeSourceThumbnail,
			FileType:      Photo,
			ThumbnailType: thumbType,
		},
	}
}
