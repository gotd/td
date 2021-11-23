package fileid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPhotoSizeSourceType_String(t *testing.T) {
	for i := PhotoSizeSourceLegacy; i <= lastPhotoSizeSourceType+1; i++ {
		require.NotEmpty(t, i.String())
	}
}
