package downloader

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// eofMock answers upload.getFileHashes past the end of the file with
// VOLUME_LOC_NOT_FOUND, as Telegram does.
type eofMock struct {
	*mock
	size int64
}

func (m eofMock) UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error) {
	if request.Offset >= m.size {
		return nil, tgerr.New(400, "VOLUME_LOC_NOT_FOUND")
	}
	return m.mock.UploadGetFileHashes(ctx, request)
}

// A file that ends on a hash window border is fully verified without a hash
// request at its end, found by the live check on a 30 MB document.
func TestBuilderVerifyAlignedEOF(t *testing.T) {
	ctx := context.Background()
	const part = 4096
	for _, windows := range []int{10, 20, 25} { // full and partial hash ranges
		data := make([]byte, part*windows)
		_, err := io.ReadFull(rand.Reader, data)
		require.NoError(t, err)
		m := eofMock{mock: &mock{data: data, hashes: mockHashes{ranges: countHashes(data, part)}}, size: int64(len(data))}

		got, err := NewDownloader().Download(m, nil).Verify(ctx, bytes.NewReader(data), int64(len(data)))
		require.NoError(t, err, "windows %d", windows)
		require.Equal(t, int64(len(data)), got)

		out := new(bytes.Buffer)
		_, err = NewDownloader().Download(m, nil).WithVerify(true).Stream(ctx, out)
		require.NoError(t, err, "verified download, windows %d", windows)
		require.Equal(t, data, out.Bytes())
	}
}

// VOLUME_LOC_NOT_FOUND before any hashes is a real error (bad location).
func TestBuilderVerifyLocationNotFound(t *testing.T) {
	data := make([]byte, 4096*10)
	m := eofMock{mock: &mock{data: data, hashes: mockHashes{ranges: countHashes(data, 4096)}}, size: 0}
	_, err := NewDownloader().Download(m, nil).Verify(context.Background(), bytes.NewReader(data), int64(len(data)))
	require.True(t, tgerr.Is(err, "VOLUME_LOC_NOT_FOUND"), "got %v", err)
}
