package fileid

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func TestPhotoSizeSourceEncodeDecode(t *testing.T) {
	tests := []PhotoSizeSource{
		{
			Type:   PhotoSizeSourceLegacy,
			Secret: 10,
		},
		{
			Type:                 PhotoSizeSourceStickerSetThumbnail,
			StickerSetID:         12,
			StickerSetAccessHash: 13,
		},
		{
			Type:     PhotoSizeSourceFullLegacy,
			VolumeID: 13,
			LocalID:  14,
			Secret:   15,
		},
		{
			Type:             PhotoSizeSourceDialogPhotoBigLegacy,
			VolumeID:         13,
			LocalID:          14,
			DialogID:         -1001228418968,
			DialogAccessHash: 15,
		},
		{
			Type:                 PhotoSizeSourceStickerSetThumbnailLegacy,
			VolumeID:             10,
			LocalID:              11,
			StickerSetID:         12,
			StickerSetAccessHash: 13,
		},
		{
			Type:                 PhotoSizeSourceStickerSetThumbnailVersion,
			StickerSetID:         12,
			StickerSetAccessHash: 13,
			StickerVersion:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Type.String(), func(t *testing.T) {
			a := require.New(t)
			var b bin.Buffer

			tt.encode(&b)
			var got PhotoSizeSource
			a.NoError(got.decode(&b, latestSubVersion))
			a.Equal(tt, got)
		})
	}
}
