package uploader

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/constant"
)

// https://core.telegram.org/api/files#uploading-files
const (
	// Use upload.saveBigFilePart in case the full size of the file is more than 10 MB
	// and upload.saveFilePart for smaller files.
	bigFileLimit = constant.UploadMaxSmallSize

	// Each part should have a sequence number, file_part, with a value ranging from 0 to 3,999.
	partsLimit = constant.UploadMaxParts

	defaultPartSize = 128 * 1024 // 128 KB
	// The file’s binary content is then split into parts. All parts must have the same size (part_size)
	// and the following conditions must be met:

	// `part_size % 1024 = 0` (divisible by 1KB)
	paddingPartSize = constant.UploadPadding

	// MaximumPartSize is maximum size of single part.
	MaximumPartSize = constant.UploadMaxPartSize
)

func checkPartSize(partSize int) error {
	switch {
	case partSize == 0:
		return errors.New("is equal to zero")
	case partSize%paddingPartSize != 0:
		return errors.Errorf("%d is not divisible by %d", partSize, paddingPartSize)
	case MaximumPartSize%partSize != 0:
		return errors.Errorf("%d is not divisible by %d", MaximumPartSize, partSize)
	}

	return nil
}

func computeParts(partSize, total int) int {
	if total <= 0 {
		return 0
	}

	parts := total / partSize
	if total%partSize != 0 {
		parts++
	}
	return parts
}

func (u *Uploader) initUpload(upload *Upload) error {
	big := upload.totalBytes > bigFileLimit
	totalParts := computeParts(u.partSize, int(upload.totalBytes))
	if !big && totalParts > partsLimit {
		return errors.Errorf(
			"part size is too small: total size = %d, part size = %d, %d / %d > %d",
			upload.totalBytes, u.partSize, upload.totalBytes, u.partSize, partsLimit,
		)
	}

	if upload.id == 0 {
		id, err := u.id()
		if err != nil {
			return errors.Wrap(err, "id generation")
		}

		upload.id = id
		upload.partSize = u.partSize
	} else if upload.partSize != u.partSize {
		return errors.Errorf(
			"previous upload has part size %d, but uploader size is %d",
			upload.partSize, u.partSize,
		)
	}

	upload.big = big
	upload.totalParts = totalParts
	return nil
}
