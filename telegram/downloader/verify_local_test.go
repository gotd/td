package downloader

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilderVerify(t *testing.T) {
	ctx := context.Background()
	const part = 4096
	data := make([]byte, part*25+100) // 3 hash ranges of up to 10 hashes
	_, err := io.ReadFull(rand.Reader, data)
	require.NoError(t, err)

	builder := func() *Builder {
		m := &mock{data: data, hashes: mockHashes{ranges: countHashes(data, part)}}
		return NewDownloader().Download(m, nil)
	}
	corrupt := func(at int) []byte {
		c := bytes.Clone(data)
		c[at] ^= 0xff
		return c
	}

	for _, tt := range []struct {
		name  string
		local []byte
		want  int64
	}{
		{"Empty", nil, 0},
		{"Full", data, int64(len(data))},
		{"Prefix", data[:part*12], part * 12},
		{"PartialWindow", data[:part*12+10], part * 12},
		{"CorruptFirst", corrupt(0), 0},
		{"CorruptMiddle", corrupt(part*11 + 5), part * 11},
		{"CorruptTail", corrupt(len(data) - 1), part * 25},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := builder().Verify(ctx, bytes.NewReader(tt.local), int64(len(tt.local)))
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	t.Run("HashesError", func(t *testing.T) {
		m := &mock{data: data, hashesErr: true}
		_, err := NewDownloader().Download(m, nil).Verify(ctx, bytes.NewReader(data), int64(len(data)))
		require.Error(t, err)
	})
}
