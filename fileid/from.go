package fileid

import (
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

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

// ChatPhoto is interface for user profile photo and chat photo structures.
type ChatPhoto interface {
	GetDCID() int
	GetPhotoID() int64
}

var _ = []ChatPhoto{
	(*tg.ChatPhoto)(nil),
	(*tg.UserProfilePhoto)(nil),
}

// FromChatPhoto creates new FileID from ChatPhoto.
func FromChatPhoto(id constant.TDLibPeerID, accessHash int64, photo ChatPhoto, big bool) FileID {
	typ := PhotoSizeSourceDialogPhotoSmall
	if big {
		typ = PhotoSizeSourceDialogPhotoBig
	}
	return FileID{
		Type: ProfilePhoto,
		DC:   photo.GetDCID(),
		ID:   photo.GetPhotoID(),
		PhotoSizeSource: PhotoSizeSource{
			Type:             typ,
			DialogID:         id,
			DialogAccessHash: accessHash,
		},
	}
}
