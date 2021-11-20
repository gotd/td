package fileid

import (
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// PhotoSize represents photo metadata stored in file_id.
type PhotoSize struct {
	VolumeID  int64
	LocalID   int
	Secret    int64
	PhotoSize string

	FileType      uint32
	ThumbnailType [4]byte

	DialogID         int64
	DialogAccessHash int64

	StickerSetID         int64
	StickerSetAccessHash int64
	StickerVersion       int32
}

const (
	photoSizeSourceLegacy = iota
	photoSizeSourceThumbnail
	photoSizeSourceDialogPhotoSmall
	photoSizeSourceDialogPhotoBig
	photoSizeSourceStickerSetThumbnail
	photoSizeSourceFullLegacy
	photoSizeSourceDialogPhotoSmallLegacy
	photoSizeSourceDialogPhotoBigLegacy
	photoSizeSourceStickerSetThumbnailLegacy
	photoSizeSourceStickerSetThumbnailVersion
)

func (p *PhotoSize) readLocalIDVolumeID(b *bin.Buffer) error {
	{
		v, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read volume_id")
		}
		p.VolumeID = v
	}
	{
		v, err := b.Int()
		if err != nil {
			return errors.Wrap(err, "read local_id")
		}
		p.LocalID = v
	}
	return nil
}

func (p *PhotoSize) readDialog(b *bin.Buffer) error {
	{
		v, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read dialog_id")
		}
		p.DialogID = v
	}
	{
		v, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read dialog_access_hash")
		}
		p.DialogAccessHash = v
	}
	return nil
}

func (p *PhotoSize) readStickerSet(b *bin.Buffer) error {
	{
		v, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read sticker_set_id")
		}
		p.StickerSetID = v
	}
	{
		v, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read sticker_set_access_hash")
		}
		p.StickerSetAccessHash = v
	}
	return nil
}

func (p *PhotoSize) decode(b *bin.Buffer, subVersion byte) error {
	if subVersion < 32 {
		{
			v, err := b.Long()
			if err != nil {
				return errors.Wrap(err, "read volume_id")
			}
			p.VolumeID = v
		}

		if subVersion < 22 {
			{
				v, err := b.Long()
				if err != nil {
					return errors.Wrap(err, "read secret")
				}
				p.Secret = v
			}
			{
				v, err := b.Int()
				if err != nil {
					return errors.Wrap(err, "read local_id")
				}
				p.LocalID = v
			}

			return nil
		}
	}

	var photoSizeType int
	if subVersion >= 4 {
		v, err := b.Int()
		if err != nil {
			return errors.Wrap(err, "read photo_size_type")
		}
		photoSizeType = v
	}

	switch photoSizeType {
	case photoSizeSourceLegacy:
		v, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read secret")
		}
		p.Secret = v
	case photoSizeSourceThumbnail:
		{
			v, err := b.Uint32()
			if err != nil {
				return errors.Wrap(err, "read file_type")
			}
			p.FileType = v
		}
		{
			if len(b.Buf) < 4 {
				return io.ErrUnexpectedEOF
			}
			copy(p.ThumbnailType[:], b.Buf)
			b.Buf = b.Buf[4:]
		}
	case photoSizeSourceDialogPhotoBig, photoSizeSourceDialogPhotoSmall:
		if err := p.readDialog(b); err != nil {
			return errors.Wrap(err, "read dialog")
		}
	case photoSizeSourceStickerSetThumbnail:
		if err := p.readStickerSet(b); err != nil {
			return errors.Wrap(err, "read sticker_set")
		}
	case photoSizeSourceFullLegacy:
		{
			v, err := b.Long()
			if err != nil {
				return errors.Wrap(err, "read volume_id")
			}
			p.VolumeID = v
		}
		{
			v, err := b.Long()
			if err != nil {
				return errors.Wrap(err, "read secret")
			}
			p.Secret = v
		}
		{
			v, err := b.Int()
			if err != nil {
				return errors.Wrap(err, "read local_id")
			}
			p.LocalID = v
		}
	case photoSizeSourceDialogPhotoBigLegacy, photoSizeSourceDialogPhotoSmallLegacy:
		if err := p.readDialog(b); err != nil {
			return errors.Wrap(err, "read dialog")
		}
		if err := p.readLocalIDVolumeID(b); err != nil {
			return errors.Wrap(err, "read legacy photo")
		}

	case photoSizeSourceStickerSetThumbnailLegacy:
		if err := p.readStickerSet(b); err != nil {
			return errors.Wrap(err, "read sticker_set")
		}
		if err := p.readLocalIDVolumeID(b); err != nil {
			return errors.Wrap(err, "read legacy photo")
		}

	case photoSizeSourceStickerSetThumbnailVersion:
		if err := p.readStickerSet(b); err != nil {
			return errors.Wrap(err, "read sticker_set")
		}
		{
			v, err := b.Int32()
			if err != nil {
				return errors.Wrap(err, "read sticker_version")
			}
			p.StickerVersion = v
		}
	}

	if subVersion < 32 && subVersion >= 22 {
		v, err := b.Int()
		if err != nil {
			return errors.Wrap(err, "read local_id")
		}
		p.LocalID = v
	}
	return nil
}
