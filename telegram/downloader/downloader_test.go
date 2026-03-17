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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/syncio"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type mock struct {
	data   []byte
	hashes mockHashes
	// reupload emulates hashes returned by UploadReuploadCDNFile. When empty,
	// mock keeps old behavior and returns nil hashes.
	reupload  []tg.FileHash
	migrate   bool
	err       bool
	hashesErr bool
	redirect  *tg.UploadFileCDNRedirect
	// enforceCDNRequestRules enables strict Telegram CDN parameter checks
	// from docs: offset/limit are 4KB-aligned, limit divides 1MB and request
	// stays within a single 1MB window.
	enforceCDNRequestRules bool
	// If > 0, redirect starts from this offset (when cdn_supported is set).
	redirectAtOffset int64

	// trackWindow* is test-only instrumentation for full hash-window fetches
	// used by split-window verification path.
	trackWindowOffset int64
	trackWindowLimit  int
	trackWindowBlock  <-chan struct{}
	trackWindowCalls  atomic.Int32

	migrateOnce    atomic.Bool
	reuploadNeeded atomic.Bool
	cdnUploadTO    atomic.Bool
	tokenInvalid   atomic.Bool
	getTimeout     atomic.Bool
	cdnGetTimeout  atomic.Bool
	cdnHashTimeout atomic.Bool
	cdnFingerprint atomic.Bool
	cdnHashFP      atomic.Bool
	getFileCalls   atomic.Int32
	hashesCalls    atomic.Int32
	cdnGetCalls    atomic.Int32
	cdnReupCalls   atomic.Int32
	cdnHashCalls   atomic.Int32
}

var testErr = testutil.TestError()

func validCDNRequest(offset int64, limit int) bool {
	if limit <= 0 {
		return false
	}
	if offset < 0 {
		return false
	}
	if offset%4096 != 0 {
		return false
	}
	if limit%4096 != 0 {
		return false
	}
	const oneMB = 1024 * 1024
	if oneMB%limit != 0 {
		return false
	}
	end := offset + int64(limit) - 1
	if end < offset {
		return false
	}
	return (offset / oneMB) == (end / oneMB)
}

func (m *mock) getPart(offset int64, limit int) []byte {
	length := len(m.data)
	if offset >= int64(length) {
		return []byte{}
	}

	size := length - int(offset)
	if size > limit {
		size = limit
	}

	r := make([]byte, size)
	copy(r, m.data[offset:])
	return r
}

func (m *mock) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	m.getFileCalls.Add(1)
	if m.err {
		return nil, testErr
	}
	if m.getTimeout.CompareAndSwap(true, false) {
		return nil, tgerr.New(500, tg.ErrTimeout)
	}

	if request.GetCDNSupported() && m.migrateOnce.CompareAndSwap(true, false) {
		return m.redirect, nil
	}

	if request.GetCDNSupported() && m.redirectAtOffset > 0 && request.Offset >= m.redirectAtOffset {
		return m.redirect, nil
	}

	if request.GetCDNSupported() && m.migrate {
		return m.redirect, nil
	}

	return &tg.UploadFile{
		Bytes: m.getPart(request.Offset, request.Limit),
	}, nil
}

func (m *mock) UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error) {
	m.hashesCalls.Add(1)
	if m.hashesErr {
		return nil, testErr
	}

	return m.hashes.Hashes(ctx, request.Offset)
}

func (m *mock) UploadReuploadCDNFile(ctx context.Context, request *tg.UploadReuploadCDNFileRequest) ([]tg.FileHash, error) {
	m.cdnReupCalls.Add(1)
	if m.err {
		return nil, testErr
	}
	if m.cdnUploadTO.CompareAndSwap(true, false) {
		return nil, tgerr.New(500, "CDN_UPLOAD_TIMEOUT")
	}

	// Explicit copy avoids accidental aliasing between downloader cache and test
	// fixture slices.
	if len(m.reupload) == 0 {
		return nil, nil
	}

	r := make([]tg.FileHash, len(m.reupload))
	copy(r, m.reupload)
	return r, nil
}

func (m *mock) UploadGetCDNFile(ctx context.Context, request *tg.UploadGetCDNFileRequest) (tg.UploadCDNFileClass, error) {
	m.cdnGetCalls.Add(1)
	if m.err {
		return nil, testErr
	}
	if m.trackWindowLimit > 0 &&
		request.Offset == m.trackWindowOffset &&
		request.Limit == m.trackWindowLimit {
		m.trackWindowCalls.Add(1)
		if m.trackWindowBlock != nil {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-m.trackWindowBlock:
			}
		}
	}
	if m.enforceCDNRequestRules && !validCDNRequest(request.Offset, request.Limit) {
		return nil, tgerr.New(400, "LIMIT_INVALID")
	}
	if m.cdnGetTimeout.CompareAndSwap(true, false) {
		return nil, tgerr.New(500, tg.ErrTimeout)
	}
	if m.cdnFingerprint.CompareAndSwap(true, false) {
		return nil, exchange.ErrKeyFingerprintNotFound
	}

	if m.tokenInvalid.CompareAndSwap(true, false) {
		return nil, tgerr.New(400, "FILE_TOKEN_INVALID")
	}

	if m.reuploadNeeded.CompareAndSwap(true, false) {
		return &tg.UploadCDNFileReuploadNeeded{
			RequestToken: []byte{1, 2, 3},
		}, nil
	}

	block, err := aes.NewCipher(m.redirect.EncryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "CDN mock cipher creation")
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

func (m *mock) UploadGetCDNFileHashes(ctx context.Context, request *tg.UploadGetCDNFileHashesRequest) ([]tg.FileHash, error) {
	m.cdnHashCalls.Add(1)
	m.hashesCalls.Add(1)
	if m.cdnHashTimeout.CompareAndSwap(true, false) {
		return nil, tgerr.New(500, tg.ErrTimeout)
	}
	if m.cdnHashFP.CompareAndSwap(true, false) {
		return nil, exchange.ErrKeyFingerprintNotFound
	}
	if m.hashesErr {
		return nil, testErr
	}

	return m.hashes.Hashes(ctx, request.Offset)
}

func (m *mock) UploadGetWebFile(ctx context.Context, request *tg.UploadGetWebFileRequest) (*tg.UploadWebFile, error) {
	if m.err {
		return nil, testErr
	}

	return &tg.UploadWebFile{
		Bytes: m.getPart(int64(request.Offset), request.Limit),
	}, nil
}

type noopCloser struct{}

func (noopCloser) Close() error {
	return nil
}

func (m *mock) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	return m, noopCloser{}, nil
}

type noCDNClient struct {
	base *mock
}

func (c *noCDNClient) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	return c.base.UploadGetFile(ctx, request)
}

func (c *noCDNClient) UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error) {
	return c.base.UploadGetFileHashes(ctx, request)
}

func (c *noCDNClient) UploadReuploadCDNFile(ctx context.Context, request *tg.UploadReuploadCDNFileRequest) ([]tg.FileHash, error) {
	return c.base.UploadReuploadCDNFile(ctx, request)
}

func (c *noCDNClient) UploadGetCDNFileHashes(ctx context.Context, request *tg.UploadGetCDNFileHashesRequest) ([]tg.FileHash, error) {
	return c.base.UploadGetCDNFileHashes(ctx, request)
}

func (c *noCDNClient) UploadGetWebFile(ctx context.Context, request *tg.UploadGetWebFileRequest) (*tg.UploadWebFile, error) {
	return c.base.UploadGetWebFile(ctx, request)
}

type nilCDNProvider struct {
	*mock
}

func (c *nilCDNProvider) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	return nil, noopCloser{}, nil
}

type errCDNProvider struct {
	*mock
	err error
}

func (c *errCDNProvider) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	return nil, nil, c.err
}

// retryAttemptProvider emulates redirect refresh to another DC and one
// fingerprint error during client creation on refresh path.
type retryAttemptProvider struct {
	base *mock

	initialRedirect *tg.UploadFileCDNRedirect
	refreshRedirect *tg.UploadFileCDNRedirect

	masterCalls atomic.Int32
	cdnCalls    atomic.Int32
}

func (p *retryAttemptProvider) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	if request.GetCDNSupported() {
		if p.masterCalls.Add(1) == 1 {
			return p.initialRedirect, nil
		}
		return p.refreshRedirect, nil
	}
	return p.base.UploadGetFile(ctx, request)
}

func (p *retryAttemptProvider) UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error) {
	return p.base.UploadGetFileHashes(ctx, request)
}

func (p *retryAttemptProvider) UploadReuploadCDNFile(ctx context.Context, request *tg.UploadReuploadCDNFileRequest) ([]tg.FileHash, error) {
	return p.base.UploadReuploadCDNFile(ctx, request)
}

func (p *retryAttemptProvider) UploadGetCDNFileHashes(ctx context.Context, request *tg.UploadGetCDNFileHashesRequest) ([]tg.FileHash, error) {
	return p.base.UploadGetCDNFileHashes(ctx, request)
}

func (p *retryAttemptProvider) UploadGetWebFile(ctx context.Context, request *tg.UploadGetWebFileRequest) (*tg.UploadWebFile, error) {
	return p.base.UploadGetWebFile(ctx, request)
}

func (p *retryAttemptProvider) UploadGetCDNFile(ctx context.Context, request *tg.UploadGetCDNFileRequest) (tg.UploadCDNFileClass, error) {
	return p.base.UploadGetCDNFile(ctx, request)
}

func (p *retryAttemptProvider) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	if p.cdnCalls.Add(1) == 3 {
		return nil, nil, exchange.ErrKeyFingerprintNotFound
	}
	return p, noopCloser{}, nil
}

// refreshRetryProvider emulates one timeout from master while refreshing CDN
// redirect after token invalidation.
type refreshRetryProvider struct {
	*mock

	redirect           *tg.UploadFileCDNRedirect
	masterCalls        atomic.Int32
	refreshTimeoutOnce atomic.Bool
}

func (p *refreshRetryProvider) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	if request.GetCDNSupported() {
		if p.masterCalls.Add(1) == 2 && p.refreshTimeoutOnce.CompareAndSwap(true, false) {
			return nil, tgerr.New(500, tg.ErrTimeout)
		}
		return p.redirect, nil
	}
	return p.mock.UploadGetFile(ctx, request)
}

func (p *refreshRetryProvider) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	return p, noopCloser{}, nil
}

// reuploadRetryProvider emulates one retryable token error on reupload call.
type reuploadRetryProvider struct {
	*mock

	tokenInvalidOnce atomic.Bool
}

func (p *reuploadRetryProvider) UploadReuploadCDNFile(ctx context.Context, request *tg.UploadReuploadCDNFileRequest) ([]tg.FileHash, error) {
	if p.tokenInvalidOnce.CompareAndSwap(true, false) {
		return nil, tgerr.New(400, "REQUEST_TOKEN_INVALID")
	}
	return p.mock.UploadReuploadCDNFile(ctx, request)
}

func (p *reuploadRetryProvider) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	return p, noopCloser{}, nil
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
			Offset: int64(offset),
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
			to := int(hash.Offset) + hash.Limit
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
		name        string
		data        []byte
		migrate     bool
		cdnReupload bool
		cdnTokenErr bool
		err         bool
		hashesErr   bool
	}{
		{"5b", []byte{1, 2, 3, 4, 5}, false, false, false, false, false},
		{strconv.Itoa(len(testData)) + "b", testData, false, false, false, false, false},
		{"Error", []byte{}, false, false, false, true, false},
		{"HashesError", testData, false, false, false, false, true},
		{"Migrate", testData, true, false, false, false, false},
		{"MigrateReupload", testData, true, true, false, false, false},
		{"MigrateTokenInvalid", testData, true, false, true, false, false},
	}
	schemas := []struct {
		name    string
		creator func(c Client) *Builder
	}{
		{"Master", func(c Client) *Builder {
			return NewDownloader().WithAllowCDN(true).Download(c, nil)
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
										migrate:   test.migrate,
										err:       test.err,
										hashesErr: test.hashesErr,
										redirect:  redirect,
									}
									if test.cdnReupload {
										client.reuploadNeeded.Store(true)
									}
									if test.cdnTokenErr {
										client.tokenInvalid.Store(true)
									}

									b := schema.creator(client)
									b = option.action(b)
									data, err := way.action(b)
									shouldErr := test.err || (test.hashesErr && option.name == "Verify")
									if shouldErr {
										a.Error(err)
									} else {
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

func TestDownloader_CDNFallbackWithoutProvider(t *testing.T) {
	ctx := context.Background()
	data := []byte("fallback-without-cdn-provider")

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	redirect := &tg.UploadFileCDNRedirect{
		DCID:          203,
		FileToken:     []byte{10},
		EncryptionKey: key,
		EncryptionIv:  iv,
	}

	t.Run("NoProvider", func(t *testing.T) {
		m := &mock{
			data:    data,
			migrate: true,
			hashes: mockHashes{
				ranges: countHashes(data, 128*1024),
			},
			redirect: redirect,
		}
		output := new(bytes.Buffer)
		_, err := NewDownloader().
			WithAllowCDN(true).
			Download(&noCDNClient{base: m}, nil).
			WithVerify(true).
			Stream(ctx, output)
		require.NoError(t, err)
		require.Equal(t, data, output.Bytes())
		require.EqualValues(t, 1, m.getFileCalls.Load())
	})

	t.Run("NilProvider", func(t *testing.T) {
		m := &mock{
			data:    data,
			migrate: true,
			hashes: mockHashes{
				ranges: countHashes(data, 128*1024),
			},
			redirect: redirect,
		}
		output := new(bytes.Buffer)
		_, err := NewDownloader().
			WithAllowCDN(true).
			Download(&nilCDNProvider{mock: m}, nil).
			WithVerify(true).
			Stream(ctx, output)
		require.Error(t, err)
	})

	t.Run("ProviderErrorReturnsError", func(t *testing.T) {
		m := &mock{
			data:    data,
			migrate: true,
			hashes: mockHashes{
				ranges: countHashes(data, 128*1024),
			},
			redirect: redirect,
		}
		output := new(bytes.Buffer)
		_, err := NewDownloader().
			WithAllowCDN(true).
			Download(&errCDNProvider{
				mock: m,
				err:  testErr,
			}, nil).
			WithVerify(true).
			Stream(ctx, output)
		require.Error(t, err)
		require.ErrorIs(t, err, testErr)
	})

	t.Run("ProviderContextError", func(t *testing.T) {
		m := &mock{
			data:    data,
			migrate: true,
			hashes: mockHashes{
				ranges: countHashes(data, 128*1024),
			},
			redirect: redirect,
		}
		output := new(bytes.Buffer)
		_, err := NewDownloader().
			WithAllowCDN(true).
			Download(&errCDNProvider{
				mock: m,
				err:  context.Canceled,
			}, nil).
			WithVerify(true).
			Stream(ctx, output)
		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("TokenInvalidFallbackToMaster", func(t *testing.T) {
		m := &mock{
			data: data,
			hashes: mockHashes{
				ranges: countHashes(data, 128*1024),
			},
			redirect: redirect,
		}
		m.migrateOnce.Store(true)
		m.tokenInvalid.Store(true)

		output := new(bytes.Buffer)
		_, err := NewDownloader().
			WithAllowCDN(true).
			Download(m, nil).
			WithVerify(true).
			Stream(ctx, output)
		require.NoError(t, err)
		require.Equal(t, data, output.Bytes())
	})
}
func TestDownloader_CDNDisabledByDefault(t *testing.T) {
	// Default NewDownloader() must stay strictly backward compatible:
	// redirect-capable mock still should run through master-only flow.
	ctx := context.Background()
	data := []byte("cdn-policy-disabled")
	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}
	redirect := &tg.UploadFileCDNRedirect{
		DCID:          203,
		FileToken:     []byte{10},
		EncryptionKey: key,
		EncryptionIv:  iv,
	}
	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		redirect: redirect,
	}
	output := new(bytes.Buffer)
	_, err := NewDownloader().Download(m, nil).WithVerify(true).Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	// CDN is opt-in only.
	require.EqualValues(t, 1, m.getFileCalls.Load())
}

func TestDownloader_AllowCDNNoRedirectKeepsLegacyLoad(t *testing.T) {
	// Main compatibility check for explicit AllowCDN=true:
	// when server does not return redirect we should not issue extra hash RPCs.
	ctx := context.Background()
	const threads = 4
	data := make([]byte, defaultPartSize*2)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
	}

	output := new(syncio.BufWriterAt)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithThreads(threads).
		Parallel(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.Zero(t, m.hashesCalls.Load())
	require.Zero(t, m.cdnHashCalls.Load())
}

func TestDownloader_AllowCDNNoRedirectNoExtraProbeRequest(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize*2)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
	}

	output := new(syncio.BufWriterAt)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithThreads(1).
		Parallel(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	// 2 parts + EOF probe, no prepare-stage prefetch.
	require.EqualValues(t, 3, m.getFileCalls.Load())
}

func TestDownloader_NonCDNDefaultAvoidsExtraGetFile(t *testing.T) {
	ctx := context.Background()
	const threads = 4
	data := make([]byte, defaultPartSize*2)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
	}

	output := new(syncio.BufWriterAt)
	_, err := NewDownloader().Download(m, nil).WithThreads(threads).Parallel(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	// Unknown-size parallel downloads may have up to threads-1 in-flight tail probes.
	// Baseline here is 3 calls: 2 full parts + 1 EOF probe.
	calls := m.getFileCalls.Load()
	require.GreaterOrEqual(t, calls, int32(3))
	require.LessOrEqual(t, calls, int32(3+(threads-1)))
}

func TestDownloader_WithAllowCDNDisabledMatchesLegacy(t *testing.T) {
	ctx := context.Background()
	data := []byte("legacy-master-only")
	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          1,
			FileToken:     []byte{1},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	output := new(bytes.Buffer)
	_, err := NewDownloader().WithAllowCDN(false).Download(m, nil).Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.EqualValues(t, 1, m.getFileCalls.Load())
}

func TestDownloader_CDNLateRedirectDefaultEnablesVerify(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize*2)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		redirectAtOffset: defaultPartSize,
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	// Hash validation must be enabled by default when CDN flow is allowed,
	// even if redirect happens after the first chunk.
	require.Greater(t, m.hashesCalls.Load(), int32(0))
}

func TestDownloader_CDNDefaultVerifyDetectsHashMismatch(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	hashes := countHashes(data, 128*1024)
	hashes[0][0].Hash = bytes.Repeat([]byte{0x42}, len(hashes[0][0].Hash))

	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: hashes,
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, io.Discard)
	require.ErrorIs(t, err, ErrHashMismatch)
}

func TestDownloader_CDNVerifyCannotBeDisabled(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	hashes := countHashes(data, 128*1024)
	hashes[0][0].Hash = bytes.Repeat([]byte{0x42}, len(hashes[0][0].Hash))

	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: hashes,
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithVerify(false).
		Stream(ctx, io.Discard)
	require.ErrorIs(t, err, ErrHashMismatch)
}

func TestDownloader_CDNSplitWindowFullFetchDeduplicated(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := make([]byte, 128*1024)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	windowBlock := make(chan struct{})
	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{11},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
		trackWindowOffset: 0,
		trackWindowLimit:  128 * 1024,
		trackWindowBlock:  windowBlock,
	}

	writer := new(syncio.BufWriterAt)
	errCh := make(chan error, 1)
	go func() {
		_, err := NewDownloader().
			WithPartSize(64*1024).
			WithAllowCDN(true).
			Download(m, nil).
			WithThreads(2).
			Parallel(ctx, writer)
		errCh <- err
	}()

	require.Eventually(t, func() bool {
		return m.trackWindowCalls.Load() >= 1
	}, time.Second, 10*time.Millisecond)
	// Keep the first full-window request blocked for a short period to give
	// concurrent chunk verification a chance to request the same window.
	time.Sleep(50 * time.Millisecond)
	close(windowBlock)

	require.NoError(t, <-errCh)
	require.Equal(t, int32(1), m.trackWindowCalls.Load())
	require.Equal(t, data, writer.Bytes())
}

func TestDownloader_CDNDefaultVerifyAllowsShortFinalChunk(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, 131072+10093)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			// Last hash has nominal 128KB limit but hash bytes are for short tail.
			ranges: countHashes(data, 128*1024),
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithPartSize(128*1024).
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
}

func TestDownloader_WithAllowCDNDisabledNoCDNMethodsCalled(t *testing.T) {
	ctx := context.Background()
	const threads = 4
	data := make([]byte, defaultPartSize*2)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data:    data,
		migrate: true, // Server would redirect if cdn_supported is set.
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

	output := new(syncio.BufWriterAt)
	_, err := NewDownloader().
		WithAllowCDN(false).
		Download(m, nil).
		WithThreads(threads).
		Parallel(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.Zero(t, m.cdnGetCalls.Load())
	require.Zero(t, m.cdnReupCalls.Load())
	require.Zero(t, m.cdnHashCalls.Load())
}

func TestDownloader_CDNFingerprintMissRetriesGetFile(t *testing.T) {
	ctx := context.Background()
	data := []byte("cdn-fingerprint-retry")
	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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
	m.cdnFingerprint.Store(true)

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithVerify(true).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.GreaterOrEqual(t, m.cdnGetCalls.Load(), int32(2))
}

func TestDownloader_CDNFingerprintMissRetriesHashes(t *testing.T) {
	ctx := context.Background()
	data := []byte("cdn-hash-fingerprint-retry")
	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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
	m.cdnHashFP.Store(true)

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithVerify(true).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.GreaterOrEqual(t, m.cdnHashCalls.Load(), int32(2))
}

func TestDownloader_CDNParallelMultiThread(t *testing.T) {
	ctx := context.Background()
	const threads = 8
	data := make([]byte, defaultPartSize*6+777)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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

	output := new(syncio.BufWriterAt)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithThreads(threads).
		Parallel(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.Greater(t, m.cdnGetCalls.Load(), int32(0))
	require.Greater(t, m.cdnHashCalls.Load(), int32(0))
}

func TestDownloader_ConcurrentMixedCDNAndNonCDN(t *testing.T) {
	ctx := context.Background()

	mustRandom := func(size int) []byte {
		b := make([]byte, size)
		if _, err := io.ReadFull(rand.Reader, b); err != nil {
			t.Fatal(err)
		}
		return b
	}
	newRedirect := func() *tg.UploadFileCDNRedirect {
		key := make([]byte, 32)
		iv := make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			t.Fatal(err)
		}
		return &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		}
	}

	cdnData := mustRandom(defaultPartSize*2 + 111)
	legacyData := mustRandom(defaultPartSize*2 + 222)
	noRedirectData := mustRandom(defaultPartSize*2 + 333)

	cdnMock := &mock{
		data:     cdnData,
		migrate:  true,
		hashes:   mockHashes{ranges: countHashes(cdnData, 128*1024)},
		redirect: newRedirect(),
	}
	legacyMock := &mock{
		data:     legacyData,
		migrate:  true,
		hashes:   mockHashes{ranges: countHashes(legacyData, 128*1024)},
		redirect: newRedirect(),
	}
	noRedirectMock := &mock{
		data:     noRedirectData,
		hashes:   mockHashes{ranges: countHashes(noRedirectData, 128*1024)},
		redirect: newRedirect(),
	}

	type result struct {
		name string
		data []byte
		err  error
	}
	results := make(chan result, 3)
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		out := new(syncio.BufWriterAt)
		_, err := NewDownloader().
			WithAllowCDN(true).
			Download(cdnMock, nil).
			WithThreads(4).
			Parallel(ctx, out)
		results <- result{name: "cdn", data: out.Bytes(), err: err}
	}()

	go func() {
		defer wg.Done()
		out := new(syncio.BufWriterAt)
		_, err := NewDownloader().
			WithAllowCDN(false).
			Download(legacyMock, nil).
			WithThreads(4).
			Parallel(ctx, out)
		results <- result{name: "legacy", data: out.Bytes(), err: err}
	}()

	go func() {
		defer wg.Done()
		out := new(syncio.BufWriterAt)
		_, err := NewDownloader().
			WithAllowCDN(true).
			Download(noRedirectMock, nil).
			WithThreads(4).
			Parallel(ctx, out)
		results <- result{name: "no-redirect", data: out.Bytes(), err: err}
	}()

	wg.Wait()
	close(results)

	got := map[string]result{}
	for r := range results {
		got[r.name] = r
		require.NoError(t, r.err)
	}

	require.Equal(t, cdnData, got["cdn"].data)
	require.Equal(t, legacyData, got["legacy"].data)
	require.Equal(t, noRedirectData, got["no-redirect"].data)
	require.Greater(t, cdnMock.cdnGetCalls.Load(), int32(0))
	require.Zero(t, legacyMock.cdnGetCalls.Load())
	require.Zero(t, noRedirectMock.cdnGetCalls.Load())
}

func TestDownloader_CDNRetriesOnTimeout(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize*2)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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
	m.cdnGetTimeout.Store(true)
	m.cdnHashTimeout.Store(true)

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.GreaterOrEqual(t, m.cdnGetCalls.Load(), int32(2))
	require.GreaterOrEqual(t, m.cdnHashCalls.Load(), int32(2))
}

func TestDownloader_RetryHandlerCDNPath(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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
	// Cover two retry sources:
	// - reader/verifier retry on timeout from CDN hashes
	// - CDN state machine retry on fingerprint miss from getCdnFile
	m.cdnHashTimeout.Store(true)
	m.cdnFingerprint.Store(true)

	output := new(bytes.Buffer)
	var (
		mu     sync.Mutex
		events []RetryEvent
	)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithRetryHandler(func(event RetryEvent) {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
		}).
		WithVerify(true).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, events)

	ops := make(map[string]bool, len(events))
	for _, event := range events {
		require.NotEmpty(t, event.Operation)
		require.GreaterOrEqual(t, event.Attempt, 1)
		require.Error(t, event.Err)
		ops[event.Operation] = true
	}

	require.True(t, ops[RetryOperationReaderHashes])
	require.True(t, ops[RetryOperationGetFile])
}

func TestDownloader_RetryHandlerLegacyPath(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize+13)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
	}
	m.getTimeout.Store(true)

	output := new(bytes.Buffer)
	var (
		mu     sync.Mutex
		events []RetryEvent
	)
	_, err := NewDownloader().
		Download(m, nil).
		WithRetryHandler(func(event RetryEvent) {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
		}).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, events)

	hasReaderChunk := false
	for _, event := range events {
		if event.Operation == RetryOperationReaderChunk {
			hasReaderChunk = true
			require.GreaterOrEqual(t, event.Attempt, 1)
			require.Error(t, event.Err)
		}
	}
	require.True(t, hasReaderChunk)
}

func TestDownloader_RetryHandlerBuilderIsolationCDNPath(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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
	// Force retries from both verifier and CDN state machine.
	m.cdnHashTimeout.Store(true)
	m.cdnFingerprint.Store(true)

	downloader := NewDownloader().WithAllowCDN(true)

	var (
		firstMu  sync.Mutex
		first    []RetryEvent
		secondMu sync.Mutex
		second   []RetryEvent
	)

	firstBuilder := downloader.
		Download(m, nil).
		WithRetryHandler(func(event RetryEvent) {
			firstMu.Lock()
			first = append(first, event)
			firstMu.Unlock()
		}).
		WithVerify(true)

	// Build second request on same downloader but do not run it. Its handler
	// must not receive retries from first builder.
	_ = downloader.Download(m, nil).WithRetryHandler(func(event RetryEvent) {
		secondMu.Lock()
		second = append(second, event)
		secondMu.Unlock()
	})

	output := new(bytes.Buffer)
	_, err := firstBuilder.Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	firstMu.Lock()
	firstEvents := len(first)
	firstMu.Unlock()
	secondMu.Lock()
	secondEvents := len(second)
	secondMu.Unlock()

	require.NotZero(t, firstEvents)
	require.Zero(t, secondEvents)
}

func TestDownloader_RetryHandlerCreateClientAttemptFromRefresh(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	redirectOne := &tg.UploadFileCDNRedirect{
		DCID:          203,
		FileToken:     []byte{10},
		EncryptionKey: key,
		EncryptionIv:  iv,
	}
	redirectTwo := &tg.UploadFileCDNRedirect{
		DCID:          204,
		FileToken:     []byte{11},
		EncryptionKey: key,
		EncryptionIv:  iv,
	}

	base := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		redirect: redirectOne,
	}
	// First CDN request: fingerprint miss (outer retry #1).
	// Second CDN request: token invalid -> refresh redirect.
	base.cdnFingerprint.Store(true)
	base.tokenInvalid.Store(true)

	provider := &retryAttemptProvider{
		base:            base,
		initialRedirect: redirectOne,
		refreshRedirect: redirectTwo,
	}

	output := new(bytes.Buffer)
	var (
		mu     sync.Mutex
		events []RetryEvent
	)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(provider, nil).
		WithRetryHandler(func(event RetryEvent) {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
		}).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	mu.Lock()
	defer mu.Unlock()

	createAttempts := make([]int, 0, 1)
	for _, event := range events {
		if event.Operation == RetryOperationCreateClient {
			createAttempts = append(createAttempts, event.Attempt)
		}
	}
	require.Equal(t, []int{3}, createAttempts)
}

func TestDownloader_RetryHandlerRefreshRedirect(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	redirect := &tg.UploadFileCDNRedirect{
		DCID:          203,
		FileToken:     []byte{10},
		EncryptionKey: key,
		EncryptionIv:  iv,
	}

	base := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		migrate:  true,
		redirect: redirect,
	}
	// Trigger refresh path from CDN state machine.
	base.tokenInvalid.Store(true)
	provider := &refreshRetryProvider{
		mock:     base,
		redirect: redirect,
	}
	provider.refreshTimeoutOnce.Store(true)

	output := new(bytes.Buffer)
	var (
		mu     sync.Mutex
		events []RetryEvent
	)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(provider, nil).
		WithRetryHandler(func(event RetryEvent) {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
		}).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	mu.Lock()
	defer mu.Unlock()

	var refreshAttempts []int
	for _, event := range events {
		if event.Operation == RetryOperationRefreshRedirect {
			refreshAttempts = append(refreshAttempts, event.Attempt)
		}
	}
	require.Equal(t, []int{1}, refreshAttempts)
}

func TestDownloader_RetryHandlerReupload(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	base := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		migrate: true,
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}
	base.reuploadNeeded.Store(true)
	provider := &reuploadRetryProvider{mock: base}
	provider.tokenInvalidOnce.Store(true)

	output := new(bytes.Buffer)
	var (
		mu     sync.Mutex
		events []RetryEvent
	)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(provider, nil).
		WithRetryHandler(func(event RetryEvent) {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
		}).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	mu.Lock()
	defer mu.Unlock()

	var reuploadAttempts []int
	for _, event := range events {
		if event.Operation == RetryOperationReupload {
			reuploadAttempts = append(reuploadAttempts, event.Attempt)
		}
	}
	require.Len(t, reuploadAttempts, 1)
	require.GreaterOrEqual(t, reuploadAttempts[0], 1)
}

func TestDownloader_RetryHandlerGetFileHashes(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
		migrate: true,
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}
	m.cdnHashFP.Store(true)

	output := new(bytes.Buffer)
	var (
		mu     sync.Mutex
		events []RetryEvent
	)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		WithRetryHandler(func(event RetryEvent) {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
		}).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	mu.Lock()
	defer mu.Unlock()

	var hashAttempts []int
	for _, event := range events {
		if event.Operation == RetryOperationGetFileHashes {
			hashAttempts = append(hashAttempts, event.Attempt)
		}
	}
	require.Equal(t, []int{1}, hashAttempts)
}

func TestDownloader_LegacyRetriesOnTimeout(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize+13)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data: data,
		hashes: mockHashes{
			ranges: countHashes(data, 128*1024),
		},
	}
	m.getTimeout.Store(true)

	output := new(bytes.Buffer)
	_, err := NewDownloader().Download(m, nil).Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.GreaterOrEqual(t, m.getFileCalls.Load(), int32(2))
}

// Verifies split-window happy path for small custom part size (64KB):
// download must not fail with "hash window exceeds remaining chunk".
func TestDownloader_CDNSmallPartSizeNoHashWindowError(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, 131072*3+10093)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithPartSize(64*1024).
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
}

// Verifies split-window happy path for unaligned part size (160KB) where CDN
// hash windows (128KB) are crossed by request boundaries.
func TestDownloader_CDNUnalignedPartSizeNoHashWindowError(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, 131072*5+7777)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithPartSize(160*1024).
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
}

func TestDownloader_CDNUnalignedPartSizeRespectsCDNRequestLimits(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, 131072*5+7777)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data:                   data,
		migrate:                true,
		enforceCDNRequestRules: true,
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

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithPartSize(160*1024).
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.Greater(t, m.cdnGetCalls.Load(), int32(0))
}

// Covers tricky tail case: second hash window is shorter than nominal limit
// and also split by unaligned part size. Download must stay successful.
func TestDownloader_CDNUnalignedPartSizeFinalSplitNoHashWindowError(t *testing.T) {
	ctx := context.Background()
	// Two hash windows where the second one is short and split by part size:
	// [0, 128KB) + [128KB, ~200KB), partSize=160KB.
	data := make([]byte, 200*1024)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithPartSize(160*1024).
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
}

// Integrity regression test for small part size: tampered payload must be
// rejected even when hash windows are split between chunks.
func TestDownloader_CDNSmallPartSizeDetectsHashMismatch(t *testing.T) {
	ctx := context.Background()
	original := make([]byte, 131072*3+10093)
	if _, err := io.ReadFull(rand.Reader, original); err != nil {
		t.Fatal(err)
	}
	hashRanges := countHashes(original, 128*1024)

	// Tamper payload after hash snapshot: downloader must reject.
	data := append([]byte(nil), original...)
	data[100] ^= 0xFF

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: hashRanges,
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	_, err := NewDownloader().
		WithPartSize(64*1024).
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, io.Discard)
	require.ErrorIs(t, err, ErrHashMismatch)
}

// Integrity regression test for unaligned part size. Corruption inside a
// window that spans chunks must still be detected.
func TestDownloader_CDNUnalignedPartSizeDetectsHashMismatch(t *testing.T) {
	ctx := context.Background()
	original := make([]byte, 131072*5+7777)
	if _, err := io.ReadFull(rand.Reader, original); err != nil {
		t.Fatal(err)
	}
	hashRanges := countHashes(original, 128*1024)

	// Corrupt bytes inside hash window [128KB, 256KB) that spans request chunks
	// when part size is 160KB.
	data := append([]byte(nil), original...)
	data[140*1024] ^= 0xFF

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	m := &mock{
		data:    data,
		migrate: true,
		hashes: mockHashes{
			ranges: hashRanges,
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}

	_, err := NewDownloader().
		WithPartSize(160*1024).
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, io.Discard)
	require.ErrorIs(t, err, ErrHashMismatch)
}

// Ensures hashes returned by UploadReuploadCDNFile are consumed immediately:
// retry should proceed without extra UploadGetCDNFileHashes call.
func TestDownloader_CDNUsesReuploadHashes(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

	hashRanges := countHashes(data, 128*1024)
	m := &mock{
		data:      data,
		migrate:   true,
		reupload:  hashRanges[0],
		hashesErr: true,
		hashes: mockHashes{
			ranges: hashRanges,
		},
		redirect: &tg.UploadFileCDNRedirect{
			DCID:          203,
			FileToken:     []byte{10},
			EncryptionKey: key,
			EncryptionIv:  iv,
		},
	}
	m.reuploadNeeded.Store(true)

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())
	require.EqualValues(t, 1, m.cdnReupCalls.Load())
	require.Zero(t, m.cdnHashCalls.Load())
}

func TestDownloader_CDNReuploadTimeoutDoesNotFallbackToMaster(t *testing.T) {
	ctx := context.Background()
	data := make([]byte, defaultPartSize)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		t.Fatal(err)
	}

	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		t.Fatal(err)
	}

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
	// First CDN chunk requests reupload; first reupload call returns
	// CDN_UPLOAD_TIMEOUT. TDesktop does not switch to master refresh flow for
	// this error and fails the task.
	m.reuploadNeeded.Store(true)
	m.cdnUploadTO.Store(true)

	output := new(bytes.Buffer)
	_, err := NewDownloader().
		WithAllowCDN(true).
		Download(m, nil).
		Stream(ctx, output)
	require.Error(t, err)
	require.ErrorContains(t, err, "CDN_UPLOAD_TIMEOUT")
	// Only initial redirect request to master is expected: no fallback refresh.
	require.EqualValues(t, 1, m.getFileCalls.Load())
	require.EqualValues(t, 1, m.cdnReupCalls.Load())
}
