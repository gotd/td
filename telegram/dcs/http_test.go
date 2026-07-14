package dcs_test

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

// httpWaiter mirrors the optional capability the mtproto layer type-asserts on
// the transport to drive http_wait long-polling.
type httpWaiter interface {
	HTTPWaitParams() (maxDelay, waitAfter, maxWait int)
	StartHTTPWait(frame func(ctx context.Context) (*bin.Buffer, error))
}

// listFor builds a DC list with a single primary DC pointing at the given test
// server, and returns the resolver options aimed at it.
func listFor(t *testing.T, srv *httptest.Server) (dcs.List, dcs.HTTPOptions) {
	t.Helper()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(u.Port())
	require.NoError(t, err)

	list := dcs.List{Options: []tg.DCOption{{
		ID:        2,
		IPAddress: u.Hostname(),
		Port:      port,
	}}}
	return list, dcs.HTTPOptions{Scheme: u.Scheme, Port: port}
}

func TestHTTP_SendRecv(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodPost, r.Method)
		a.Equal("/api", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		a.NoError(err)
		_, _ = w.Write(body) // echo
	}))
	defer srv.Close()

	list, opts := listFor(t, srv)
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn.Close() }()

	data, err := io.ReadAll(io.LimitReader(rand.Reader, 256))
	a.NoError(err)
	a.NoError(conn.Send(ctx, &bin.Buffer{Buf: data}))

	var b bin.Buffer
	a.NoError(conn.Recv(ctx, &b))
	a.Equal(data, b.Buf)
}

func TestHTTP_EmptyResponseNoFrame(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK) // empty body: "dummy" long-poll response
	}))
	defer srv.Close()

	list, opts := listFor(t, srv)
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn.Close() }()

	a.NoError(conn.Send(ctx, &bin.Buffer{Buf: []byte{1, 2, 3}}))

	recvCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	var b bin.Buffer
	err = conn.Recv(recvCtx, &b)
	a.ErrorIs(err, context.DeadlineExceeded, "empty response must not produce a frame")
}

func TestHTTP_WaitParams(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	list, opts := listFor(t, srv)

	// Defaults.
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn.Close() }()
	w, ok := conn.(httpWaiter)
	a.True(ok, "http transport must expose the http_wait capability")
	md, wa, mw := w.HTTPWaitParams()
	a.Equal(0, md)
	a.Equal(0, wa)
	a.Equal(25000, mw)

	// Custom.
	opts.MaxDelay, opts.WaitAfter, opts.MaxWait = 5, 10, 15000
	conn2, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn2.Close() }()
	md, wa, mw = conn2.(httpWaiter).HTTPWaitParams()
	a.Equal(5, md)
	a.Equal(10, wa)
	a.Equal(15000, mw)
}

func TestHTTP_StartHTTPWaitDeliversServerData(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	const waitFrame = "http-wait-frame"
	payload := []byte("server-pushed-update")

	var gotWaitFrame atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) == waitFrame {
			gotWaitFrame.Store(true)
		}
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	list, opts := listFor(t, srv)
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn.Close() }()

	var calls atomic.Int64
	conn.(httpWaiter).StartHTTPWait(func(context.Context) (*bin.Buffer, error) {
		calls.Add(1)
		return &bin.Buffer{Buf: []byte(waitFrame)}, nil
	})

	// The long-poll loop must deliver server data with no application Send.
	recvCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var b bin.Buffer
	a.NoError(conn.Recv(recvCtx, &b))
	a.Equal(payload, b.Buf)
	a.True(gotWaitFrame.Load(), "server must have received the http_wait frame")
	a.Positive(calls.Load(), "frame factory must be invoked by the poll loop")
}

func TestHTTP_UnsupportedMediaCDN(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	list := dcs.List{Options: []tg.DCOption{{ID: 2, IPAddress: "127.0.0.1"}}}
	res := dcs.HTTP(dcs.HTTPOptions{})

	_, err := res.MediaOnly(ctx, 2, list)
	a.Error(err)
	_, err = res.CDN(ctx, 2, list)
	a.Error(err)
}

func TestHTTP_NoAddresses(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	_, err := dcs.HTTP(dcs.HTTPOptions{}).Primary(ctx, 2, dcs.List{})
	a.Error(err)
}

// A 4-byte response body is a transport-level protocol error (e.g. -404 auth key
// not found), not an encrypted message, and must surface from Recv as a
// codec.ProtocolErr so the read loop can react.
func TestHTTP_TransportProtocolError(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		code := int32(-codec.CodeAuthKeyNotFound) // -404
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(code))
		_, _ = w.Write(buf[:])
	}))
	defer srv.Close()

	list, opts := listFor(t, srv)
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn.Close() }()

	a.NoError(conn.Send(ctx, &bin.Buffer{Buf: []byte{1, 2, 3, 4, 5}}))

	var b bin.Buffer
	err = conn.Recv(ctx, &b)
	var pe *codec.ProtocolErr
	a.True(errors.As(err, &pe), "4-byte body must surface as ProtocolErr, got %v", err)
	a.Equal(int32(codec.CodeAuthKeyNotFound), pe.Code)
}

func TestHTTP_CloseDuringInflight(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		<-release // hold the response until the test releases it
	}))
	// LIFO: release first (unblock handler), then Close (waits for handlers).
	defer srv.Close()
	defer close(release)

	list, opts := listFor(t, srv)
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)

	a.NoError(conn.Send(ctx, &bin.Buffer{Buf: []byte{1, 2, 3}}))
	a.NoError(conn.Close())

	// Recv after Close returns promptly with an error, not a hang.
	recvCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var b bin.Buffer
	a.Error(conn.Recv(recvCtx, &b))
	// Send after Close is rejected.
	a.Error(conn.Send(ctx, &bin.Buffer{Buf: []byte{1}}))
}

// Run with -race: many concurrent Send plus the poll loop must not corrupt the
// inbox or race on the URL index.
func TestHTTP_ConcurrentSendRace(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	list, opts := listFor(t, srv)
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn.Close() }()

	conn.(httpWaiter).StartHTTPWait(func(context.Context) (*bin.Buffer, error) {
		return &bin.Buffer{Buf: []byte("wait")}, nil
	})

	drain := make(chan struct{})
	go func() {
		defer close(drain)
		for i := 0; i < 30; i++ {
			rctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
			var b bin.Buffer
			_ = conn.Recv(rctx, &b)
			cancel()
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = conn.Send(ctx, &bin.Buffer{Buf: []byte("x")})
		}()
	}
	wg.Wait()
	<-drain
}

// A middlebox returning an instant empty 200 (not honoring the long-poll) must
// be throttled so the poll loop does not storm the DC with requests.
func TestHTTP_ThrottlesInstantEmptyResponse(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	var reqs atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs.Add(1)
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK) // instant empty response
	}))
	defer srv.Close()

	list, opts := listFor(t, srv)
	conn, err := dcs.HTTP(opts).Primary(ctx, 2, list)
	a.NoError(err)
	defer func() { _ = conn.Close() }()

	conn.(httpWaiter).StartHTTPWait(func(context.Context) (*bin.Buffer, error) {
		return &bin.Buffer{Buf: []byte("wait")}, nil
	})

	time.Sleep(1200 * time.Millisecond)
	n := reqs.Load()
	// Throttled to ~one poll per httpWaitRetryInterval; without the throttle this
	// would be thousands over the same window.
	a.Positive(n, "poll loop should still send http_wait")
	a.Less(n, int64(20), "instant empty responses must be throttled, got %d", n)
}
