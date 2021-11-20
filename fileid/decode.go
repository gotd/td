package fileid

import (
	"bytes"
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
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return FileID{}, errors.Wrap(err, "base64")
	}
	data = rleDecode(data)

	switch version := data[len(data)-1]; version {
	case persistentIDVersionOld, persistentIDVersionMap:
		return FileID{}, errors.Errorf("%v is unsupported now", version)
	case persistentIDVersion:
		data = data[:len(data)-1]
		return decodeLatestFileID(&bin.Buffer{Buf: data})
	default:
		return FileID{}, errors.Errorf("unknown file_id version %x", version)
	}
}

func rleDecode(s []byte) (r []byte) {
	var last []byte
	for _, cur := range s {
		if string(last) == string(rune(0)) {
			r = append(r, bytes.Repeat(last, int(cur))...)
			last = nil
		} else {
			r = append(r, last...)
			last = []byte{cur}
		}
	}
	r = append(r, last...)
	return r
}
