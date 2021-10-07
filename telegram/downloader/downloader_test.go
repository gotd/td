package downloader

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/syncio"
	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/tg"
)

type mock struct {
	data      []byte
	hashes    mockHashes
	migrate   bool
	err       bool
	hashesErr bool
	redirect  *tg.UploadFileCDNRedirect
}

var testErr = testutil.TestError()

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
	if m.hashesErr {
		return nil, testErr
	}

	return m.hashes.Hashes(ctx, request.Offset)
}

func (m mock) UploadReuploadCDNFile(ctx context.Context, request *tg.UploadReuploadCDNFileRequest) ([]tg.FileHash, error) {
	panic("implement me")
}

func (m mock) UploadGetCDNFile(ctx context.Context, request *tg.UploadGetCDNFileRequest) (tg.UploadCDNFileClass, error) {
	if m.err {
		return nil, testErr
	}

	if m.migrate {
		return &tg.UploadCDNFileReuploadNeeded{
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
	return &tg.UploadCDNFile{
		Bytes: r,
	}, nil
}

func (m mock) UploadGetCDNFileHashes(ctx context.Context, request *tg.UploadGetCDNFileHashesRequest) ([]tg.FileHash, error) {
	if m.hashesErr {
		return nil, testErr
	}

	return m.hashes.Hashes(ctx, request.Offset)
}

func (m mock) UploadGetWebFile(ctx context.Context, request *tg.UploadGetWebFileRequest) (*tg.UploadWebFile, error) {
	if m.err {
		return nil, testErr
	}

	return &tg.UploadWebFile{
		Bytes: m.getPart(request.Offset, request.Limit),
	}, nil
}

func countHashes(data []byte, partSize int) (r [][]tg.FileHash) {
	actions := data
	batchSize := partSize
	batches := make([][]byte, 0, (len(actions)+batchSize-1)/batchSize)

	for batchSize < len(actions) {
		actions, batches = actions[batchSize:], append(batches, actions[0:batchSize:batchSize])
	}
	batches = append(batches, actions)

	currentRange := make([]tg.FileHash, 0, 10)
	offset := 0
	for _, batch := range batches {
		if len(currentRange) >= 10 {
			r = append(r, currentRange)
			currentRange = make([]tg.FileHash, 0, 10)
		}
		currentRange = append(currentRange, tg.FileHash{
			Offset: offset,
			Limit:  partSize,
			Hash:   crypto.SHA256(batch),
		})
		offset += len(batch)

		if len(batch) < partSize {
			break
		}
	}
	r = append(r, currentRange)
	return
}

func Test_countHashes(t *testing.T) {
	a := require.New(t)
	data := bytes.Repeat([]byte{1, 2, 3, 4, 5}, 10)
	hashes := countHashes(data, 4)

	a.NotEmpty(hashes)
	for _, hashRange := range hashes {
		for _, hash := range hashRange {
			from := hash.Offset
			to := hash.Offset + hash.Limit
			if to > len(data) {
				to = len(data)
			}
			a.Equal(crypto.SHA256(data[from:to]), hash.Hash)
		}
	}
}

func TestDownloader(t *testing.T) {
	ctx := context.Background()

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}
	redirect := &tg.UploadFileCDNRedirect{
		DCID:          1,
		FileToken:     []byte{10},
		EncryptionKey: key,
		EncryptionIv:  iv,
	}

	testData := make([]byte, defaultPartSize*2)
	if _, err := io.ReadFull(rand.Reader, testData); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		data      []byte
		migrate   bool
		err       bool
		hashesErr bool
	}{
		{"5b", []byte{1, 2, 3, 4, 5}, false, false, false},
		{strconv.Itoa(len(testData)) + "b", testData, false, false, false},
		{"Error", []byte{}, false, true, false},
		{"HashesError", []byte{}, false, true, true},
		{"Migrate", []byte{}, true, false, false},
	}
	schemas := []struct {
		name    string
		creator func(c Client, cdn CDN) *Builder
	}{
		{"Master", func(c Client, cdn CDN) *Builder {
			return NewDownloader().Download(c, nil)
		}},
		{"Direct", func(c Client, cdn CDN) *Builder {
			return NewDownloader().DownloadDirect(c, nil)
		}},
		{"CDN", func(c Client, cdn CDN) *Builder {
			return NewDownloader().CDN(c, cdn, redirect)
		}},
		{"Web", func(c Client, cdn CDN) *Builder {
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
			output := new(syncio.BufWriterAt)
			_, err := b.WithThreads(runtime.GOMAXPROCS(0)).Parallel(ctx, output)
			return output.Bytes(), err
		}},
		{"Parallel-OneThread", func(b *Builder) ([]byte, error) {
			output := new(syncio.BufWriterAt)
			_, err := b.WithThreads(1).Parallel(ctx, output)
			return output.Bytes(), err
		}},
	}
	options := []struct {
		name   string
		action func(b *Builder) *Builder
	}{
		{"NoVerify", func(b *Builder) *Builder {
			return b.WithVerify(false)
		}},
		{"Verify", func(b *Builder) *Builder {
			return b.WithVerify(true)
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
					for _, option := range options {
						// Telegram can't return hashes for web files.
						if schema.name == "Web" && option.name == "Verify" {
							continue
						}

						t.Run(option.name, func(t *testing.T) {
							for _, way := range ways {
								t.Run(way.name, func(t *testing.T) {
									a := require.New(t)
									client := &mock{
										data: test.data,
										hashes: mockHashes{
											ranges: countHashes(test.data, 128*1024),
										},
										migrate:  test.migrate,
										err:      test.err,
										redirect: redirect,
									}

									b := schema.creator(client, client)
									b = option.action(b)
									data, err := way.action(b)
									switch {
									case test.migrate:
										a.Error(err)
									case test.err:
										a.Error(err)
									default:
										a.NoError(err)
										a.True(bytes.Equal(test.data, data))
									}
								})
							}
						})
					}
				})
			}
		})
	}
}
