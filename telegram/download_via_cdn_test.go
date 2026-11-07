package telegram

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"sync"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/tg"
)

// fakeInvoker answers requests with handler and records their types.
type fakeInvoker struct {
	handler func(input bin.Encoder) (bin.Encoder, error)

	mux   sync.Mutex
	calls []string
}

func (f *fakeInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	f.mux.Lock()
	if obj, ok := input.(interface{ TypeName() string }); ok {
		f.calls = append(f.calls, obj.TypeName())
	}
	f.mux.Unlock()

	result, err := f.handler(input)
	if err != nil {
		return err
	}
	var b bin.Buffer
	if err := result.Encode(&b); err != nil {
		return err
	}
	return output.Decode(&b)
}

func (f *fakeInvoker) Close() error { return nil }

func (f *fakeInvoker) called() []string {
	f.mux.Lock()
	defer f.mux.Unlock()
	return append([]string(nil), f.calls...)
}

func TestClientDownloadViaCDN(t *testing.T) {
	ctx := context.Background()

	// One CDN hash window of data.
	data := make([]byte, 128*1024)
	key := make([]byte, 32)
	iv := make([]byte, aes.BlockSize)
	for _, b := range [][]byte{data, key, iv} {
		_, err := io.ReadFull(rand.Reader, b)
		require.NoError(t, err)
	}
	redirect := &tg.UploadFileCDNRedirect{
		DCID:          203,
		FileToken:     []byte{1, 2, 3},
		EncryptionKey: key,
		EncryptionIv:  iv,
		FileHashes: []tg.FileHash{{
			Offset: 0,
			Limit:  len(data),
			Hash:   crypto.SHA256(data),
		}},
	}

	// Master DC requests must go to the given invoker.
	master := &fakeInvoker{handler: func(input bin.Encoder) (bin.Encoder, error) {
		req, ok := input.(*tg.UploadGetFileRequest)
		if !ok || !req.CDNSupported {
			return nil, errors.Errorf("unexpected master request %T", input)
		}
		return redirect, nil
	}}

	// CDN data requests must go to the CDN pool of the client.
	cdnPool := &fakeInvoker{handler: func(input bin.Encoder) (bin.Encoder, error) {
		req, ok := input.(*tg.UploadGetCDNFileRequest)
		if !ok {
			return nil, errors.Errorf("unexpected CDN request %T", input)
		}
		if !bytes.Equal(req.FileToken, redirect.FileToken) {
			return nil, errors.New("unexpected file token")
		}

		var part []byte
		if req.Offset < int64(len(data)) {
			part = data[req.Offset:]
			if len(part) > req.Limit {
				part = part[:req.Limit]
			}
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		ctrIV := append([]byte(nil), iv...)
		binary.BigEndian.PutUint32(ctrIV[len(ctrIV)-4:], uint32(req.Offset/16))
		encrypted := make([]byte, len(part))
		cipher.NewCTR(block, ctrIV).XORKeyStream(encrypted, part)
		return &tg.UploadCDNFile{Bytes: encrypted}, nil
	}}

	client := newCDNPoolTestClient()
	defer client.cancel()
	client.allowCDN = true
	// Prefill CDN pool of DC 203 to avoid network: unlimited pool covers any
	// requested size.
	client.cdnPools.publishOrAcquire(203, 0, cdnPool)

	output := new(bytes.Buffer)
	_, err := client.DownloadVia(master, &tg.InputDocumentFileLocation{ID: 1}).Stream(ctx, output)
	require.NoError(t, err)
	require.Equal(t, data, output.Bytes())

	require.NotEmpty(t, master.called())
	for _, name := range master.called() {
		require.Equal(t, "upload.getFile", name)
	}
	require.NotEmpty(t, cdnPool.called())
	for _, name := range cdnPool.called() {
		require.Equal(t, "upload.getCdnFile", name)
	}
}
