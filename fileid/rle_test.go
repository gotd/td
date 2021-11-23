package fileid

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_rleEncode(t *testing.T) {
	for name, input := range testData {
		t.Run(name, func(t *testing.T) {
			a := require.New(t)
			original, err := base64Decode(input)
			a.NoError(err)

			decoded := rleDecode(original)
			a.Equal(original, rleEncode(decoded))
		})
	}
}

func Test_rleDecode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{
			"Valid",
			testData["Sticker"],
			[]uint8{
				0x08, 0x00, 0x00, 0x02, 0x02, 0x00, 0x00, 0x00, 0x19, 0x01, 0x00, 0x00, 0x00, 0x3a, 0x61, 0x99,
				0x43, 0x10, 0x70, 0xa6, 0x69, 0x32, 0xab, 0x52, 0x10, 0x88, 0x8f, 0x10, 0x0f, 0xb4, 0xfb, 0x63,
				0x57, 0x1e, 0x00, 0x00, 0x27, 0x00, 0x00, 0x00, 0x87, 0x57, 0x8f, 0x1e, 0x9d, 0x18, 0xbc, 0x87,
				0xc1, 0x75, 0x4d, 0x70, 0x22, 0x04,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			decoded, err := base64.URLEncoding.DecodeString(tt.input)
			a.NoError(err)

			a.Equal(tt.want, rleDecode(decoded))
		})
	}
}
