package telegram

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestClientFileDC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := &subConnsTestServer{
		exports: map[int]int{},
		block:   map[int]chan struct{}{},
		fail:    map[int]int{},
	}
	var calls atomic.Int64
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		if _, ok := body.(*tg.UploadGetFileRequest); ok {
			calls.Add(1)
			if dc == 3 {
				return nil, tgerr.New(303, "FILE_MIGRATE_4")
			}
			return &tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: []byte{byte(dc)}}, nil
		}
		return s.handle(id, dc, body)
	})
	client.onTransfer = noopOnTransfer
	cfg := client.cfg.Load()
	for _, dc := range []int{3, 4} {
		cfg.DCOptions = append(cfg.DCOptions, tg.DCOption{ID: dc, IPAddress: "dc"})
	}
	client.cfg.Store(cfg)

	var mwCalls atomic.Int64
	client.mw = []Middleware{MiddlewareFunc(func(next tg.Invoker) InvokeFunc {
		return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			mwCalls.Add(1)
			return next.Invoke(ctx, input, output)
		}
	})}

	getFile := func(dc int) (byte, error) {
		inv, err := client.FileDC(ctx, dc)
		if err != nil {
			return 0, err
		}
		r, err := tg.NewClient(inv).UploadGetFile(ctx, &tg.UploadGetFileRequest{
			Location: &tg.InputDocumentFileLocation{ID: 1},
			Limit:    1024,
		})
		if err != nil {
			return 0, err
		}
		return r.(*tg.UploadFile).Bytes[0], nil
	}

	// Cached pool, one transfer.
	for range 3 {
		b, err := getFile(4)
		require.NoError(t, err)
		require.Equal(t, byte(4), b)
	}
	require.Equal(t, 1, s.exportCount(4))

	// FILE_MIGRATE from DC 3 is followed to DC 4.
	b, err := getFile(3)
	require.NoError(t, err)
	require.Equal(t, byte(4), b)
	require.Equal(t, 1, s.exportCount(3))
	require.Equal(t, 1, s.exportCount(4), "DC 4 pool is reused")

	// Home DC: primary session, no authorization transfer. Checked last: test
	// primary connections reset client config.
	b, err = getFile(2)
	require.NoError(t, err)
	require.Equal(t, byte(2), b)
	require.Zero(t, s.exportCount(2))

	require.EqualValues(t, 5, mwCalls.Load(), "middlewares are applied once per request")

	client.fileInvokersMux.Lock()
	require.Len(t, client.fileInvokers, 3, "invokers are cached per DC")
	client.fileInvokersMux.Unlock()
}

func TestClientFileDCNotRunning(t *testing.T) {
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		return nil, tgerr.New(400, "UNEXPECTED")
	})
	client.ctx = nil

	_, err := client.FileDC(context.Background(), 4)
	require.ErrorIs(t, err, errClientNotRunning)
}

func TestClientFileDCUnauthorizedTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := &subConnsTestServer{
		exports: map[int]int{},
		block:   map[int]chan struct{}{},
		fail:    map[int]int{},
	}
	var (
		mux        sync.Mutex
		exported   bool
		authorized bool
	)
	client := newMigrationClient(t, func(id int64, dc int, body bin.Encoder) (bin.Encoder, error) {
		switch body.(type) {
		case *tg.AuthExportAuthorizationRequest:
			mux.Lock()
			first := !exported
			exported = true
			mux.Unlock()
			if first {
				// Not signed in yet: transfer fails, pool is still returned.
				_, _ = s.handle(id, dc, body)
				return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
			}
		case *tg.AuthImportAuthorizationRequest:
			mux.Lock()
			authorized = true
			mux.Unlock()
		case *tg.UploadGetFileRequest:
			mux.Lock()
			ok := authorized
			mux.Unlock()
			if !ok {
				return nil, tgerr.New(401, "AUTH_KEY_UNREGISTERED")
			}
			return &tg.UploadFile{Type: &tg.StorageFileUnknown{}, Bytes: []byte{byte(dc)}}, nil
		}
		return s.handle(id, dc, body)
	})
	client.onTransfer = noopOnTransfer
	cfg := client.cfg.Load()
	cfg.DCOptions = append(cfg.DCOptions, tg.DCOption{ID: 4, IPAddress: "dc"})
	client.cfg.Store(cfg)

	inv, err := client.FileDC(ctx, 4)
	require.NoError(t, err)

	r, err := tg.NewClient(inv).UploadGetFile(ctx, &tg.UploadGetFileRequest{
		Location: &tg.InputDocumentFileLocation{ID: 1},
		Limit:    1024,
	})
	require.NoError(t, err, "pool with failed transfer must be recreated")
	require.Equal(t, byte(4), r.(*tg.UploadFile).Bytes[0])
	require.Equal(t, 2, s.exportCount(4))
}
