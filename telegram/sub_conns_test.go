package telegram

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type subConnsTestServer struct {
	mux     sync.Mutex
	exports map[int]int
	// block export to DC until channel is closed.
	block map[int]chan struct{}
	// fail export to DC this many times.
	fail map[int]int
}

func (s *subConnsTestServer) handle(_ int64, dc int, body bin.Encoder) (bin.Encoder, error) {
	switch r := body.(type) {
	case *tg.AuthExportAuthorizationRequest:
		s.mux.Lock()
		s.exports[r.DCID]++
		block := s.block[r.DCID]
		fail := s.fail[r.DCID] > 0
		if fail {
			s.fail[r.DCID]--
		}
		s.mux.Unlock()

		if block != nil {
			<-block
		}
		if fail {
			return nil, tgerr.New(500, "EXPORT_FAILED")
		}
		return &tg.AuthExportedAuthorization{ID: int64(r.DCID), Bytes: []byte{1}}, nil
	case *tg.AuthImportAuthorizationRequest:
		return &tg.AuthAuthorization{User: &tg.User{ID: 1}}, nil
	case *tg.HelpGetNearestDCRequest:
		return &tg.NearestDC{ThisDC: dc}, nil
	default:
		return nil, tgerr.New(400, "UNEXPECTED")
	}
}

func (s *subConnsTestServer) exportCount(dc int) int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.exports[dc]
}

func newSubConnsTestClient(t *testing.T, s *subConnsTestServer) *Client {
	client := newMigrationClient(t, s.handle)
	client.onTransfer = noopOnTransfer

	cfg := client.cfg.Load()
	for _, dc := range []int{3, 4, 5} {
		cfg.DCOptions = append(cfg.DCOptions, tg.DCOption{ID: dc, IPAddress: "dc"})
	}
	client.cfg.Store(cfg)
	return client
}

func invokeSubNearest(ctx context.Context, c *Client, dc int) error {
	var r tg.NearestDC
	return c.invokeSub(ctx, dc, &tg.HelpGetNearestDCRequest{}, &r)
}

func TestInvokeSubDoesNotBlockOtherDCs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	release := make(chan struct{})
	s := &subConnsTestServer{
		exports: map[int]int{},
		block:   map[int]chan struct{}{4: release},
		fail:    map[int]int{},
	}
	client := newSubConnsTestClient(t, s)

	// Cached sub-connection to DC 3.
	require.NoError(t, invokeSubNearest(ctx, client, 3))

	// Slow creation of sub-connection to DC 4, shared by two callers.
	errs := make(chan error, 2)
	for range 2 {
		go func() { errs <- invokeSubNearest(ctx, client, 4) }()
	}
	require.Eventually(t, func() bool { return s.exportCount(4) == 1 }, 5*time.Second, time.Millisecond)

	// Neither cached DC 3, nor new DC 5 wait for DC 4.
	require.NoError(t, invokeSubNearest(ctx, client, 3))
	require.NoError(t, invokeSubNearest(ctx, client, 5))

	// Caller can give up waiting while creation goes on.
	shortCtx, shortCancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer shortCancel()
	require.ErrorIs(t, invokeSubNearest(shortCtx, client, 4), context.DeadlineExceeded)

	close(release)
	for range 2 {
		require.NoError(t, <-errs)
	}
	require.Equal(t, 1, s.exportCount(3))
	require.Equal(t, 1, s.exportCount(4), "creation must be shared")
	require.Equal(t, 1, s.exportCount(5))
}

func TestInvokeSubFailureNotCached(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := &subConnsTestServer{
		exports: map[int]int{},
		block:   map[int]chan struct{}{},
		fail:    map[int]int{3: 1},
	}
	client := newSubConnsTestClient(t, s)

	err := invokeSubNearest(ctx, client, 3)
	require.True(t, tgerr.Is(err, "EXPORT_FAILED"), "%+v", err)

	require.NoError(t, invokeSubNearest(ctx, client, 3))
	require.NoError(t, invokeSubNearest(ctx, client, 3))
	require.Equal(t, 2, s.exportCount(3))
}
