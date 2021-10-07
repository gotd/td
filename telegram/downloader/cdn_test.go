package downloader

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func Test_cdn_decrypt(t *testing.T) {
	testdata := make([]byte, 32)
	tests := []struct {
		name    string
		key, iv []byte
		err     bool
	}{
		{"Bad key", []byte{10}, nil, true},
		{"Bad IV", make([]byte, 32), nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &cdn{
				redirect: &tg.UploadFileCDNRedirect{
					EncryptionKey: test.key,
					EncryptionIv:  test.iv,
				},
			}
			_, err := c.decrypt(testdata, 0)
			if test.err {
				require.Error(t, err)
			}
		})
	}
}
