package fileid

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

// PhotoSizeSource represents photo metadata stored in file_id.
type PhotoSizeSource struct {
	Type      PhotoSizeSourceType
	VolumeID  int64
	LocalID   int
	Secret    int64
	PhotoSize string

	FileType      Type
	ThumbnailType rune

	DialogID         int64
	DialogAccessHash int64

	StickerSetID         int64
	StickerSetAccessHash int64
	StickerVersion       int32
}

func (p *PhotoSizeSource) stickerSet() tg.InputStickerSetClass {
	return &tg.InputStickerSetID{
		ID:         p.StickerSetID,
		AccessHash: p.StickerSetAccessHash,
	}
}

func (p *PhotoSizeSource) dialogPeer() tg.InputPeerClass {
	switch id := p.DialogID; {
	case id > 0 && id <= constant.MaxUserID:
		return &tg.InputPeerUser{
			UserID:     id,
			AccessHash: p.DialogAccessHash,
		}
	case id < 0 && -constant.MaxChatID <= id:
		return &tg.InputPeerChat{
			ChatID: id,
		}
	case id < 0 && constant.ZeroChannelID-constant.MaxChannelID <= id && id != constant.ZeroChannelID:
		return &tg.InputPeerChannel{
			ChannelID:  id,
			AccessHash: p.DialogAccessHash,
		}
	}
	return &tg.InputPeerEmpty{}
}

func (p *PhotoSizeSource) readLocalIDVolumeID(b *bin.Buffer) error {
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

func (p *PhotoSizeSource) readDialog(b *bin.Buffer) error {
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

func (p *PhotoSizeSource) readStickerSet(b *bin.Buffer) error {
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

const latestSubVersion = 34

func (p *PhotoSizeSource) decode(b *bin.Buffer, subVersion byte) error {
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

	var photoSizeType PhotoSizeSourceType
	if subVersion >= 4 {
		v, err := b.Int()
		if err != nil {
			return errors.Wrap(err, "read photo_size_type")
		}
		photoSizeType = PhotoSizeSourceType(v)
	}
	if photoSizeType < 0 || photoSizeType >= lastPhotoSizeSourceType {
		return errors.Errorf("unknown photo size source type %d", photoSizeType)
	}
	p.Type = photoSizeType

	switch photoSizeType {
	case PhotoSizeSourceLegacy:
		v, err := b.Long()
		if err != nil {
			return errors.Wrap(err, "read secret")
		}
		p.Secret = v
	case PhotoSizeSourceThumbnail:
		{
			v, err := b.Uint32()
			if err != nil {
				return errors.Wrap(err, "read file_type")
			}
			p.FileType = Type(v)
		}
		{
			v, err := b.Int32()
			if err != nil {
				return errors.Wrap(err, "read thumbnail_type")
			}
			p.ThumbnailType = v
		}
	case PhotoSizeSourceDialogPhotoBig, PhotoSizeSourceDialogPhotoSmall:
		if err := p.readDialog(b); err != nil {
			return errors.Wrap(err, "read dialog")
		}
	case PhotoSizeSourceStickerSetThumbnail:
		if err := p.readStickerSet(b); err != nil {
			return errors.Wrap(err, "read sticker_set")
		}
	case PhotoSizeSourceFullLegacy:
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
	case PhotoSizeSourceDialogPhotoBigLegacy, PhotoSizeSourceDialogPhotoSmallLegacy:
		if err := p.readDialog(b); err != nil {
			return errors.Wrap(err, "read dialog")
		}
		if err := p.readLocalIDVolumeID(b); err != nil {
			return errors.Wrap(err, "read legacy photo")
		}

	case PhotoSizeSourceStickerSetThumbnailLegacy:
		if err := p.readStickerSet(b); err != nil {
			return errors.Wrap(err, "read sticker_set")
		}
		if err := p.readLocalIDVolumeID(b); err != nil {
			return errors.Wrap(err, "read legacy photo")
		}

	case PhotoSizeSourceStickerSetThumbnailVersion:
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

func (p *PhotoSizeSource) writeLocalIDVolumeID(b *bin.Buffer) {
	b.PutLong(p.VolumeID)
	b.PutInt(p.LocalID)
}

func (p *PhotoSizeSource) writeDialog(b *bin.Buffer) {
	b.PutLong(p.DialogID)
	b.PutLong(p.DialogAccessHash)
}

func (p *PhotoSizeSource) writeStickerSet(b *bin.Buffer) {
	b.PutLong(p.StickerSetID)
	b.PutLong(p.StickerSetAccessHash)
}

func (p *PhotoSizeSource) encode(b *bin.Buffer) {
	b.PutInt(int(p.Type))
	switch p.Type {
	case PhotoSizeSourceLegacy:
		b.PutLong(p.Secret)
	case PhotoSizeSourceThumbnail:
		b.PutUint32(uint32(p.FileType))
		b.PutInt32(p.ThumbnailType)
	case PhotoSizeSourceDialogPhotoBig, PhotoSizeSourceDialogPhotoSmall:
		p.writeDialog(b)
	case PhotoSizeSourceStickerSetThumbnail:
		p.writeStickerSet(b)
	case PhotoSizeSourceFullLegacy:
		b.PutLong(p.VolumeID)
		b.PutLong(p.Secret)
		b.PutInt(p.LocalID)
	case PhotoSizeSourceDialogPhotoBigLegacy, PhotoSizeSourceDialogPhotoSmallLegacy:
		p.writeDialog(b)
		p.writeLocalIDVolumeID(b)
	case PhotoSizeSourceStickerSetThumbnailLegacy:
		p.writeStickerSet(b)
		p.writeLocalIDVolumeID(b)
	case PhotoSizeSourceStickerSetThumbnailVersion:
		p.writeStickerSet(b)
		b.PutInt32(p.StickerVersion)
	}
}
