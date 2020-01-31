package downloader

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

type mock struct {
	data         []byte
	migrate, err bool
	redirect     *tg.UploadFileCdnRedirect
}

var testErr = xerrors.New("test err")

func (m mock) getPart(offset, limit int) []byte {
	length := len(m.data)
	if offset >= length {
		return []byte{}
	}

	size := length - offset
	if size > limit {
		size = limit
	}

	r := make([]byte, size)
	copy(r, m.data[offset:])
	return r
}

func (m mock) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	if m.err {
		return nil, testErr
	}

	if m.migrate {
		return m.redirect, nil
	}

	return &tg.UploadFile{
		Bytes: m.getPart(request.Offset, request.Limit),
	}, nil
}

func (m mock) UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error) {
	panic("implement me")
}

func (m mock) UploadReuploadCdnFile(ctx context.Context, request *tg.UploadReuploadCdnFileRequest) ([]tg.FileHash, error) {
	panic("implement me")
}

func (m mock) UploadGetCdnFile(ctx context.Context, request *tg.UploadGetCdnFileRequest) (tg.UploadCdnFileClass, error) {
	if m.err {
		return nil, testErr
	}

	if m.migrate {
		return &tg.UploadCdnFileReuploadNeeded{
			RequestToken: []byte{1, 2, 3},
		}, nil
	}

	block, err := aes.NewCipher(m.redirect.EncryptionKey)
	if err != nil {
		return nil, xerrors.Errorf("CDN mock cipher creation: %w", err)
	}

	iv := make([]byte, len(m.redirect.EncryptionIv))
	copy(iv, m.redirect.EncryptionIv)
	binary.BigEndian.PutUint32(iv[len(iv)-4:], uint32(request.Offset/16))

	part := m.getPart(request.Offset, request.Limit)
	r := make([]byte, len(part))
	cipher.NewCTR(block, iv).XORKeyStream(r, part)
	return &tg.UploadCdnFile{
		Bytes: r,
	}, nil
}

func (m mock) UploadGetCdnFileHashes(ctx context.Context, request *tg.UploadGetCdnFileHashesRequest) ([]tg.FileHash, error) {
	panic("implement me")
}

func (m mock) UploadGetWebFile(ctx context.Context, request *tg.UploadGetWebFileRequest) (*tg.UploadWebFile, error) {
	if m.err {
		return nil, testErr
	}

	return &tg.UploadWebFile{
		Bytes: m.getPart(request.Offset, request.Limit),
	}, nil
}

type bufWriterAt struct {
	buf []byte
}

func (b *bufWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	ends := len(p) + int(off)
	if len(b.buf) < ends {
		newBuf := make([]byte, ends)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}

	copy(b.buf[off:], p)
	return len(b.buf), nil
}

func TestDownloader(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}
	redirect := &tg.UploadFileCdnRedirect{
		DCID:          1,
		FileToken:     []byte{10},
		EncryptionKey: key,
		EncryptionIv:  iv,
	}

	testData := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, testData); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		data    []byte
		migrate bool
		err     bool
	}{
		{"5b", []byte{1, 2, 3, 4, 5}, false, false},
		{strconv.Itoa(len(testData)) + "b", testData, false, false},
		{"Error", []byte{}, false, true},
		{"Migrate", []byte{}, true, false},
	}
	schemas := []struct {
		name    string
		creator func(c Client) *Builder
	}{
		{"Master", func(c Client) *Builder {
			return NewDownloader().Download(c, true, nil)
		}},
		{"CDN", func(c Client) *Builder {
			return NewDownloader().CDN(c, redirect)
		}},
		{"Web", func(c Client) *Builder {
			return NewDownloader().Web(c, nil)
		}},
	}
	ways := []struct {
		name   string
		action func(b *Builder) ([]byte, error)
	}{
		{"Stream", func(b *Builder) ([]byte, error) {
			output := new(bytes.Buffer)
			_, err := b.Stream(ctx, output)
			return output.Bytes(), err
		}},
		{"Parallel", func(b *Builder) ([]byte, error) {
			output := &bufWriterAt{}
			_, err := b.Parallel(ctx, output)
			return output.buf, err
		}},
	}

	for _, schema := range schemas {
		t.Run(schema.name, func(t *testing.T) {
			for _, test := range tests {
				// Telegram can't redirect web file downloads.
				if schema.name == "Web" && test.migrate {
					continue
				}

				t.Run(test.name, func(t *testing.T) {
					for _, way := range ways {
						t.Run(way.name, func(t *testing.T) {
							a := require.New(t)
							m := &mock{
								data:     test.data,
								migrate:  test.migrate,
								err:      test.err,
								redirect: redirect,
							}

							data, err := way.action(schema.creator(m))
							switch {
							case test.migrate:
								a.Error(err)
							case test.err:
								a.Error(err)
							default:
								a.NoError(err)
								a.Equal(test.data, data)
							}
						})
					}
				})
			}
		})
	}
}
