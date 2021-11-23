package fileid

import (
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// FileID represents parsed Telegram Bot API file_id.
type FileID struct {
	Type            Type
	DC              int
	ID              int64
	AccessHash      int64
	FileReference   []byte
	URL             string
	PhotoSizeSource PhotoSizeSource
}

// AsInputWebFileLocation converts file ID to tg.InputWebFileLocationClass.
func (f FileID) AsInputWebFileLocation() (tg.InputWebFileLocationClass, bool) {
	if f.URL != "" {
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
	case Photo:
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

const (
	webLocationFlag   = 1 << 24
	fileReferenceFlag = 1 << 25
)

func (f *FileID) decodeLatestFileID(b *bin.Buffer) error {
	if len(b.Buf) < 1 {
		return io.ErrUnexpectedEOF
	}
	var subVersion = b.Buf[len(b.Buf)-1]

	typeID, err := b.Uint32()
	if err != nil {
		return errors.Wrap(err, "read type_id")
	}

	hasWebLocation := typeID&webLocationFlag != 0
	hasReference := typeID&fileReferenceFlag != 0

	typeID &^= webLocationFlag
	typeID &^= fileReferenceFlag
	if typeID >= uint32(lastType) {
		return errors.Errorf("unknown type %d", typeID)
	}
	f.Type = Type(typeID)

	{
		dcID, err := b.Uint32()
		if err != nil {
			return errors.Wrap(err, "read dc_id")
		}
		f.DC = int(dcID)
	}

	if hasReference {
		reference, err := b.Bytes()
		if err != nil {
			return errors.Wrap(err, "read file_reference")
		}
		f.FileReference = reference
	}
	if hasWebLocation {
		url, err := b.String()
		if err != nil {
			return errors.Wrap(err, "read url")
		}
		f.URL = url
		return nil
	}

	{
		id, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read id")
		}
		f.ID = id
	}

	{
		accessHash, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read access_hash")
		}
		f.AccessHash = accessHash
	}

	switch Type(typeID) {
	case Thumbnail, Photo, ProfilePhoto:
	default:
		return nil
	}

	if err := f.PhotoSizeSource.decode(b, subVersion); err != nil {
		return errors.Wrap(err, "decode photo_size")
	}
	return nil
}
