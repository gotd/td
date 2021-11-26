//go:build go1.18
// +build go1.18

package fileid

import (
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func FuzzDecodeEncodeDecode(f *testing.F) {
	for name, input := range testData {
		data, err := base64Decode(input)
		if err != nil {
			f.Fatal(name, err)
		}
		data = rleDecode(data)
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		a := require.New(t)
		input := bin.Buffer{Buf: data}

		var fileID FileID
		if err := fileID.decodeLatestFileID(&input); err != nil {
			t.Skip(err)
		}
		if data[len(data)-1] < 32 {
			t.Skip("Legacy file_id encoding is not supported")
		}
		if fileID.PhotoSizeSource.Type >= PhotoSizeSourceStickerSetThumbnail {
			t.Log(pp.Sprint(input))
		}
		input.Reset()

		a.NoError(fileID.encodeLatestFileID(&input))

		var decoded FileID
		a.NoError(decoded.decodeLatestFileID(&input))
		a.Equal(fileID, decoded)
	})
}
