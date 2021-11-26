package fileid

import (
	"encoding/base64"

	"github.com/gotd/td/bin"
)

// EncodeFileID parses FileID to a string.
func EncodeFileID(id FileID) (string, error) {
	var buf bin.Buffer
	id.encodeLatestFileID(&buf)
	buf.Buf = append(buf.Buf, persistentIDVersion)
	buf.Buf = rleEncode(buf.Buf)
	return base64Encode(buf.Buf), nil
}

func base64Encode(s []byte) string {
	return base64.RawURLEncoding.EncodeToString(s)
}
