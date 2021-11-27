package fileid

import (
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
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

func (f *FileID) encodeLatestFileID(b *bin.Buffer) {
	hasWebLocation := f.URL != ""
	hasReference := len(f.FileReference) != 0

	{
		typeID := f.Type
		if hasWebLocation {
			typeID |= webLocationFlag
		}
		if hasReference {
			typeID |= fileReferenceFlag
		}
		b.PutUint32(uint32(typeID))
	}
	b.PutUint32(uint32(f.DC))
	if hasReference {
		b.PutBytes(f.FileReference)
	}
	if hasWebLocation {
		b.PutString(f.URL)
		return
	}
	b.PutLong(f.ID)
	b.PutLong(f.AccessHash)

	switch f.Type {
	case Thumbnail, Photo, ProfilePhoto:
		f.PhotoSizeSource.encode(b)
	}

	b.Buf = append(b.Buf, latestSubVersion)
}
