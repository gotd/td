package uploader

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/syncio"
	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/telegram/uploader/source"
	"github.com/nnqq/td/tg"
)

type mockClient struct {
	err bool

	// Upload state.
	buf   *syncio.BufWriterAt
	parts []atomic.Int64

	partSize    int
	partSizeMux sync.Mutex
}

func newMockClient(err bool) *mockClient {
	return &mockClient{
		err:   err,
		buf:   &syncio.BufWriterAt{},
		parts: make([]atomic.Int64, partsLimit+1),
	}
}

var testErr = testutil.TestError()

func (m *mockClient) write(part int, data []byte) error {
	m.partSizeMux.Lock()
	if m.partSize == 0 {
		m.partSize = len(data)
	} else if m.partSize != len(data) {
		m.partSizeMux.Unlock()

		return xerrors.Errorf(
			"invalid part size, expected %d, got %d",
			m.partSize, len(data),
		)
	}
	partSize := m.partSize
	m.partSizeMux.Unlock()

	// Every part have ID which is offset in partSize from start of file.
	// But maximal ID is 3999, so part ID for big files can overflow.
	// We use parts array to count received parts by ID to compute the offset.
	rangeOffset := int(m.parts[part].Inc() - 1)

	// If rangeOffset is zero, so offset will be zero, part ID received first time.
	// Otherwise, we count next range offset.
	offset := rangeOffset * partsLimit * partSize
	_, err := m.buf.WriteAt(data, int64(part*partSize+offset))
	return err
}

func (m *mockClient) UploadSaveFilePart(ctx context.Context, request *tg.UploadSaveFilePartRequest) (bool, error) {
	if m.err {
		return false, testErr
	}

	if err := m.write(request.FilePart, request.Bytes); err != nil {
		return false, err
	}

	return true, nil
}

func (m *mockClient) UploadSaveBigFilePart(ctx context.Context, request *tg.UploadSaveBigFilePartRequest) (bool, error) {
	if m.err {
		return false, testErr
	}

	if err := m.write(request.FilePart, request.Bytes); err != nil {
		return false, err
	}

	return true, nil
}

type mockSource struct {
	name string
	data *bytes.Reader
}

func (m mockSource) Open(ctx context.Context, u *url.URL) (source.RemoteFile, error) {
	return m, nil
}

func (m mockSource) Read(p []byte) (n int, err error) {
	return m.data.Read(p)
}

func (m mockSource) Close() error {
	return nil
}

func (m mockSource) Name() string {
	return m.name
}

func (m mockSource) Size() int64 {
	return m.data.Size()
}

func TestUploader(t *testing.T) {
	ctx := context.Background()

	testData := make([]byte, 15*1024*1024)
	if _, err := io.ReadFull(rand.Reader, testData); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		data []byte
		err  bool
	}{
		{"5b", []byte{1, 2, 3, 4, 5}, false},
		{strconv.Itoa(defaultPartSize) + "b", bytes.Repeat([]byte{1}, defaultPartSize), false},
		{strconv.Itoa(len(testData)) + "b", testData, false},
		{"Error", []byte{1, 2, 3, 4, 5}, true},
	}

	ways := []struct {
		name   string
		action func(b *Uploader, data []byte) error
	}{
		{"FromReader", func(b *Uploader, data []byte) error {
			if len(data) == len(testData) {
				b = b.WithPartSize(16384)
			}
			_, err := b.FromReader(ctx, "10.jpg", bytes.NewReader(data))
			return err
		}},
		{"FromBytes", func(b *Uploader, data []byte) error {
			if len(data) == len(testData) {
				b = b.WithPartSize(MaximumPartSize)
			}

			_, err := b.FromBytes(ctx, "10.jpg", data)
			return err
		}},
		{"FromFS", func(b *Uploader, data []byte) error {
			if len(data) == len(testData) {
				b = b.WithPartSize(MaximumPartSize)
			}

			_, err := b.FromFS(ctx, fstest.MapFS{
				"10.jpg": &fstest.MapFile{
					Data: data,
				},
			}, "10.jpg")
			return err
		}},
		{"FromURL", func(b *Uploader, data []byte) error {
			if len(data) == len(testData) {
				b = b.WithPartSize(MaximumPartSize)
			}
			b = b.WithSource(mockSource{
				name: "img.jpg",
				data: bytes.NewReader(data),
			})

			_, err := b.FromURL(ctx, "http://example.com")
			return err
		}},
	}

	options := []struct {
		name   string
		action func(b *Uploader) *Uploader
	}{
		{"OneThread", func(b *Uploader) *Uploader {
			return b.WithThreads(1)
		}},
		{"ManyThread", func(b *Uploader) *Uploader {
			return b.WithThreads(runtime.GOMAXPROCS(0))
		}},
	}

	for _, way := range ways {
		t.Run(way.name, func(t *testing.T) {
			for _, option := range options {
				t.Run(option.name, func(t *testing.T) {
					for _, test := range tests {
						t.Run(test.name, func(t *testing.T) {
							client := newMockClient(test.err)
							u := NewUploader(client)

							err := way.action(option.action(u), test.data)
							if test.err {
								require.Error(t, err)
								return
							}

							require.NoError(t, err)
							require.Truef(
								t, bytes.Equal(test.data, client.buf.Bytes()),
								"expected uploaded and given equal",
							)
						})
					}
				})
			}
		})
	}
}
