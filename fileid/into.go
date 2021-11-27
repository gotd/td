package fileid

import "github.com/gotd/td/tg"

// AsInputWebFileLocation converts file ID to tg.InputWebFileLocationClass.
func (f FileID) AsInputWebFileLocation() (tg.InputWebFileLocationClass, bool) {
	if f.URL == "" {
		return nil, false
	}

	return &tg.InputWebFileLocation{
		URL:        f.URL,
		AccessHash: f.AccessHash,
	}, true
}

func (f FileID) asPhotoLocation() (tg.InputFileLocationClass, bool) {
	switch src := f.PhotoSizeSource; src.Type {
	case PhotoSizeSourceLegacy:
	case PhotoSizeSourceThumbnail:
		switch src.FileType {
		case Photo, Thumbnail:
			return &tg.InputPhotoFileLocation{
				ID:            f.ID,
				AccessHash:    f.AccessHash,
				FileReference: f.FileReference,
				ThumbSize:     string(f.PhotoSizeSource.ThumbnailType),
			}, true
		}
	case PhotoSizeSourceDialogPhotoSmall,
		PhotoSizeSourceDialogPhotoBig:
		return &tg.InputPeerPhotoFileLocation{
			Big:     src.Type == PhotoSizeSourceDialogPhotoBig,
			Peer:    src.dialogPeer(),
			PhotoID: f.ID,
		}, true
	case PhotoSizeSourceStickerSetThumbnail:
	case PhotoSizeSourceFullLegacy:
		return &tg.InputPhotoLegacyFileLocation{
			ID:            f.ID,
			AccessHash:    f.AccessHash,
			FileReference: f.FileReference,
			VolumeID:      f.PhotoSizeSource.VolumeID,
			LocalID:       f.PhotoSizeSource.LocalID,
			Secret:        f.PhotoSizeSource.Secret,
		}, true
	case PhotoSizeSourceDialogPhotoSmallLegacy,
		PhotoSizeSourceDialogPhotoBigLegacy:
		return &tg.InputPeerPhotoFileLocationLegacy{
			Big:      src.Type == PhotoSizeSourceDialogPhotoBigLegacy,
			Peer:     src.dialogPeer(),
			VolumeID: f.PhotoSizeSource.VolumeID,
			LocalID:  f.PhotoSizeSource.LocalID,
		}, true
	case PhotoSizeSourceStickerSetThumbnailLegacy:
		return &tg.InputStickerSetThumbLegacy{
			Stickerset: f.PhotoSizeSource.stickerSet(),
			VolumeID:   f.PhotoSizeSource.VolumeID,
			LocalID:    f.PhotoSizeSource.LocalID,
		}, true
	case PhotoSizeSourceStickerSetThumbnailVersion:
		return &tg.InputStickerSetThumb{
			Stickerset:   f.PhotoSizeSource.stickerSet(),
			ThumbVersion: int(f.PhotoSizeSource.StickerVersion),
		}, true
	}

	return nil, false
}

// AsInputFileLocation converts file ID to tg.InputFileLocationClass.
func (f FileID) AsInputFileLocation() (tg.InputFileLocationClass, bool) {
	switch f.Type {
	case Thumbnail, ProfilePhoto, Photo:
		return f.asPhotoLocation()
	case Encrypted:
		return &tg.InputEncryptedFileLocation{
			ID:         f.ID,
			AccessHash: f.AccessHash,
		}, true
	case SecureRaw,
		Secure:
		return &tg.InputSecureFileLocation{
			ID:         f.ID,
			AccessHash: f.AccessHash,
		}, true
	case Video,
		Voice,
		Document,
		Sticker,
		Audio,
		Animation,
		VideoNote,
		Background,
		DocumentAsFile:
		return &tg.InputDocumentFileLocation{
			ID:            f.ID,
			AccessHash:    f.AccessHash,
			FileReference: f.FileReference,
			ThumbSize:     "", // ?
		}, true
	}

	return nil, false
}
