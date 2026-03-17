package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestDownloadClientDelegatesUploadMethods(t *testing.T) {
	ctx := context.Background()
	location := &tg.InputFileLocation{
		VolumeID:      1,
		LocalID:       2,
		Secret:        3,
		FileReference: []byte{4},
	}
	webLocation := &tg.InputWebFileLocation{
		URL:        "https://example.com/file.bin",
		AccessHash: 55,
	}

	client := newTestClient(func(_ int64, body bin.Encoder) (bin.Encoder, error) {
		switch req := body.(type) {
		case *tg.UploadGetFileRequest:
			require.EqualValues(t, 64, req.Offset)
			require.EqualValues(t, 16, req.Limit)
			require.Same(t, location, req.Location)
			return &tg.UploadFile{
				Type:  &tg.StorageFileUnknown{},
				Bytes: []byte("file"),
			}, nil
		case *tg.UploadGetFileHashesRequest:
			require.EqualValues(t, 128, req.Offset)
			require.Same(t, location, req.Location)
			return &tg.FileHashVector{Elems: []tg.FileHash{
				{Offset: 128, Limit: 4, Hash: []byte{1, 2, 3, 4}},
			}}, nil
		case *tg.UploadReuploadCDNFileRequest:
			require.Equal(t, []byte{9}, req.FileToken)
			require.Equal(t, []byte{8}, req.RequestToken)
			return &tg.FileHashVector{Elems: []tg.FileHash{
				{Offset: 0, Limit: 8, Hash: []byte{5, 6}},
			}}, nil
		case *tg.UploadGetCDNFileHashesRequest:
			require.Equal(t, []byte{9}, req.FileToken)
			require.EqualValues(t, 256, req.Offset)
			return &tg.FileHashVector{Elems: []tg.FileHash{
				{Offset: 256, Limit: 8, Hash: []byte{7, 8}},
			}}, nil
		case *tg.UploadGetWebFileRequest:
			require.Same(t, webLocation, req.Location)
			require.EqualValues(t, 7, req.Offset)
			require.EqualValues(t, 11, req.Limit)
			return &tg.UploadWebFile{
				FileType: &tg.StorageFileUnknown{},
				Bytes:    []byte("web"),
			}, nil
		default:
			t.Fatalf("unexpected request type %T", body)
			return nil, nil
		}
	})

	d := downloadClient{client: client}

	file, err := d.UploadGetFile(ctx, &tg.UploadGetFileRequest{
		Location: location,
		Offset:   64,
		Limit:    16,
	})
	require.NoError(t, err)
	typedFile, ok := file.(*tg.UploadFile)
	require.True(t, ok)
	require.Equal(t, []byte("file"), typedFile.Bytes)

	hashes, err := d.UploadGetFileHashes(ctx, &tg.UploadGetFileHashesRequest{
		Location: location,
		Offset:   128,
	})
	require.NoError(t, err)
	require.Len(t, hashes, 1)
	require.EqualValues(t, 128, hashes[0].Offset)

	reupload, err := d.UploadReuploadCDNFile(ctx, &tg.UploadReuploadCDNFileRequest{
		FileToken:    []byte{9},
		RequestToken: []byte{8},
	})
	require.NoError(t, err)
	require.Len(t, reupload, 1)
	require.EqualValues(t, 8, reupload[0].Limit)

	cdnHashes, err := d.UploadGetCDNFileHashes(ctx, &tg.UploadGetCDNFileHashesRequest{
		FileToken: []byte{9},
		Offset:    256,
	})
	require.NoError(t, err)
	require.Len(t, cdnHashes, 1)
	require.EqualValues(t, 256, cdnHashes[0].Offset)

	web, err := d.UploadGetWebFile(ctx, &tg.UploadGetWebFileRequest{
		Location: webLocation,
		Offset:   7,
		Limit:    11,
	})
	require.NoError(t, err)
	require.Equal(t, []byte("web"), web.Bytes)
}

func TestClientDownloadBuilders(t *testing.T) {
	c := &Client{allowCDN: true}

	require.NotNil(t, c.Downloader())
	require.NotNil(t, c.Download(&tg.InputFileLocation{}))
	require.NotNil(t, c.DownloadWeb(&tg.InputWebFileLocation{
		URL:        "https://example.com/file.bin",
		AccessHash: 1,
	}))
}

func TestDownloadClientCDNErrorPropagation(t *testing.T) {
	c := newCDNPoolTestClient()
	defer c.cancel()

	d := downloadClient{client: c}
	_, _, err := d.CDN(context.Background(), 404, 1)
	require.Error(t, err)
}
