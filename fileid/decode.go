package fileid

import (
	"encoding/base64"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

const (
	persistentIDVersionOld = 2
	persistentIDVersionMap = 3
	persistentIDVersion    = 4
)

// DecodeFileID parses FileID from a string.
func DecodeFileID(s string) (fileID FileID, _ error) {
	if s == "" {
		return FileID{}, errors.New("input is empty")
	}
	data, err := base64Decode(s)
	if err != nil {
		return FileID{}, errors.Wrap(err, "base64")
	}
	data = rleDecode(data)
	if len(data) < 2 {
		return FileID{}, errors.New("RLE-decoded data is too small")
	}

	switch version := data[len(data)-1]; version {
	case persistentIDVersionOld, persistentIDVersionMap:
		return FileID{}, errors.Errorf("%v is unsupported now", version)
	case persistentIDVersion:
		data = data[:len(data)-1]
		err := fileID.decodeLatestFileID(&bin.Buffer{Buf: data})
		return fileID, err
	default:
		return FileID{}, errors.Errorf("unknown file_id version %x", version)
	}
}

func base64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
