package qrlogin

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

type loginTokenDispatcher struct{ h tg.LoginTokenHandler }

func (d *loginTokenDispatcher) OnLoginToken(h tg.LoginTokenHandler) { d.h = h }

// updateLoginToken delivered while Auth is inside show() must not be dropped.
func TestQR_Auth_UpdateDuringShow(t *testing.T) {
	mock := tgmock.New(t)
	clk := neo.NewTime(time.Now())
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{Clock: clk})
	auth := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginToken{
		Expires: int(clk.Now().Add(time.Minute).Unix()), Token: []byte("t"),
	}).ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenSuccess{Authorization: auth})
	d := &loginTokenDispatcher{}
	loggedIn := OnLoginToken(d)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := q.Auth(ctx, loggedIn, func(ctx context.Context, token Token) error {
		// The user scans immediately; the update arrives while show() is still running.
		return d.h(ctx, tg.Entities{}, &tg.UpdateLoginToken{})
	})
	require.NoError(t, err)
	require.Equal(t, auth, res)
}

type skewInvoker struct {
	calls   atomic.Int64
	expires int
}

func (s *skewInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	s.calls.Add(1)
	r := &tg.AuthLoginTokenBox{LoginToken: &tg.AuthLoginToken{Expires: s.expires, Token: []byte("t")}}
	b := &bin.Buffer{}
	if err := r.Encode(b); err != nil {
		return err
	}
	return output.Decode(b)
}

// Local clock ahead of server time must not cause a tight export/show loop.
func TestQR_Auth_ClockSkew(t *testing.T) {
	// Server says the token expires in 30s of server time; local clock is 60s ahead.
	inv := &skewInvoker{expires: int(time.Now().Add(-30 * time.Second).Unix())}
	q := NewQR(tg.NewClient(inv), constant.TestAppID, constant.TestAppHash, Options{Clock: clock.System})
	var shows atomic.Int64
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err := q.Auth(ctx, make(chan struct{}), func(ctx context.Context, token Token) error {
		shows.Add(1)
		return nil
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	// Refresh is delayed by at least minTokenWait, so only the first export happens.
	require.Equal(t, int64(1), inv.calls.Load())
	require.Equal(t, int64(1), shows.Load())
}

// Transient refresh errors are retried with backoff; the current token stays shown.
func TestQR_Auth_RefreshRetry(t *testing.T) {
	mock := tgmock.New(t)
	clk := neo.NewTime(time.Now())
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{Clock: clk})
	auth := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginToken{
		Expires: int(clk.Now().Add(time.Minute).Unix()), Token: []byte("t"),
	}).
		ExpectCall(exportReq()).ThenRPCErr(&tgerr.Error{Code: 500, Message: "INTERNAL", Type: "INTERNAL"}).
		ExpectCall(exportReq()).ThenFlood(5).
		ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenSuccess{Authorization: auth})

	var shows atomic.Int64
	done := make(chan authResult, 1)
	go func() {
		a, err := q.Auth(context.Background(), make(chan struct{}), func(ctx context.Context, token Token) error {
			shows.Add(1)
			return nil
		})
		done <- authResult{a, err}
	}()

	deadline := time.After(10 * time.Second)
	for {
		select {
		case r := <-done:
			require.NoError(t, r.err)
			require.Equal(t, auth, r.auth)
			require.Equal(t, int64(1), shows.Load(), "token must not be re-shown on retry")
			return
		case <-deadline:
			t.Fatal("Auth did not retry")
		case <-time.After(5 * time.Millisecond):
			clk.Travel(time.Second)
		}
	}
}

// Non-transient refresh errors abort Auth.
func TestQR_Auth_RefreshFatalError(t *testing.T) {
	mock := tgmock.New(t)
	clk := neo.NewTime(time.Now())
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{Clock: clk})
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginToken{
		Expires: int(clk.Now().Add(time.Minute).Unix()), Token: []byte("t"),
	}).ExpectCall(exportReq()).ThenRPCErr(&tgerr.Error{Code: 400, Type: "API_ID_INVALID"})

	shown := make(chan struct{}, 1)
	done := make(chan error, 1)
	go func() {
		_, err := q.Auth(context.Background(), make(chan struct{}), func(ctx context.Context, token Token) error {
			shown <- struct{}{}
			return nil
		})
		done <- err
	}()
	<-shown
	clk.Travel(time.Minute + time.Second)
	err := <-done
	require.True(t, tgerr.Is(err, "API_ID_INVALID"), "unexpected error: %v", err)
}

func TestNextRetryDelay(t *testing.T) {
	a := require.New(t)
	err := &tgerr.Error{Code: 500, Type: "INTERNAL"}
	var d time.Duration
	for _, want := range []time.Duration{1, 2, 4, 8, 16, 32, 60, 60} {
		d = nextRetryDelay(d, err)
		a.Equal(want*time.Second, d)
	}
	flood := &tgerr.Error{Code: 420, Type: "FLOOD_WAIT", Argument: 30}
	a.Equal(30*time.Second, nextRetryDelay(time.Second, flood))
	a.True(retryableExportError(flood))
	a.True(retryableExportError(err))
	a.False(retryableExportError(&tgerr.Error{Code: 400, Type: "API_ID_INVALID"}))
}
