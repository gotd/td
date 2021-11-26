package fileid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileIDEncodeDecode(t *testing.T) {
	for name, input := range testData {
		t.Run(name, func(t *testing.T) {
			a := require.New(t)
			fileID, err := DecodeFileID(input)
			a.NoError(err)

			output, err := EncodeFileID(fileID)
			a.NoError(err)

			decoded, err := DecodeFileID(output)
			a.NoError(err)
			a.Equal(fileID, decoded)
		})
	}
}
