package fileid

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// FileID represents parsed Telegram Bot API file_id.
type FileID struct {
	Type          Type
	DC            int
	ID            int64
	AccessHash    int64
	FileReference string
	URL           string
	PhotoSize     PhotoSize
}

const (
	webLocationFlag   = 1 << 24
	fileReferenceFlag = 1 << 25
)

func decodeLatestFileID(b *bin.Buffer) (fileID FileID, _ error) {
	var subVersion = b.Buf[len(b.Buf)-1]

	typeID, err := b.Uint32()
	if err != nil {
		return FileID{}, errors.Wrap(err, "read type_id")
	}

	hasWebLocation := typeID&webLocationFlag != 0
	hasReference := typeID&fileReferenceFlag != 0

	typeID &^= webLocationFlag
	typeID &^= fileReferenceFlag
	if typeID >= uint32(lastType) {
		return fileID, errors.Errorf("unknown type %d", typeID)
	}
	fileID.Type = Type(typeID)

	{
		dcID, err := b.Uint32()
		if err != nil {
			return FileID{}, errors.Wrap(err, "read dc_id")
		}
		fileID.DC = int(dcID)
	}

	if hasReference {
		reference, err := b.String()
		if err != nil {
			return FileID{}, errors.Wrap(err, "read file_reference")
		}
		fileID.FileReference = reference
	}
	if hasWebLocation {
		url, err := b.String()
		if err != nil {
			return FileID{}, errors.Wrap(err, "read url")
		}
		fileID.URL = url
		return fileID, nil
	}

	{
		id, err := b.Long()
		if err != nil {
			return FileID{}, errors.Wrap(err, "read id")
		}
		fileID.ID = id
	}

	{
		accessHash, err := b.Long()
		if err != nil {
			return FileID{}, errors.Wrap(err, "read access_hash")
		}
		fileID.AccessHash = accessHash
	}

	switch Type(typeID) {
	case Thumbnail, Photo, ProfilePhoto:
	default:
		return fileID, nil
	}

	if err := fileID.PhotoSize.decode(b, subVersion); err != nil {
		return FileID{}, errors.Wrap(err, "decode photo_size")
	}
	return fileID, nil
}
