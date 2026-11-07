package downloader

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestWithCDNProvider(t *testing.T) {
	ctx := context.Background()
	data := []byte("cdn-through-plain-client")

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	_, err := io.ReadFull(rand.Reader, key)
	require.NoError(t, err)
	_, err = io.ReadFull(rand.Reader, iv)
	require.NoError(t, err)

	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	// Plain client (e.g. tg.NewClient(pool)) with CDN provider of the client.
	c := WithCDNProvider(&noCDNClient{base: m}, m)
	_, ok := c.(CDNProvider)
	require.True(t, ok)

	output := new(bytes.Buffer)
	_, err = NewDownloader().
		WithAllowCDN(true).
		Download(c, nil).
		WithVerify(true).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.Positive(t, m.cdnGetCalls.Load(), "CDN must be used")

	require.Equal(t, Client(&noCDNClient{base: m}), WithCDNProvider(&noCDNClient{base: m}, nil))
}
