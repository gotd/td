package qrlogin

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func exportReq() *tg.AuthExportLoginTokenRequest {
	return &tg.AuthExportLoginTokenRequest{APIID: constant.TestAppID, APIHash: constant.TestAppHash}
}

func TestQR_ExportResult(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	mock := tgmock.New(t)
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{})

	auth := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenSuccess{Authorization: auth})
	r, err := q.ExportResult(ctx)
	a.NoError(err)
	a.Equal(auth, r.Authorization)
	a.True(r.Token.Empty())
	a.Nil(r.MigrateTo)

	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginToken{Expires: 10, Token: []byte("t")})
	r, err = q.ExportResult(ctx)
	a.NoError(err)
	a.Equal(NewToken([]byte("t"), 10), r.Token)
	a.Nil(r.Authorization)

	migrate := &tg.AuthLoginTokenMigrateTo{DCID: 2, Token: []byte("t")}
	mock.ExpectCall(exportReq()).ThenResult(migrate)
	r, err = q.ExportResult(ctx)
	a.NoError(err)
	a.Equal(migrate, r.MigrateTo)
}

// Auth must return the authorization from the first export instead of waiting
// for an updateLoginToken that has already happened.
func TestQR_Auth_InitialSuccess(t *testing.T) {
	mock := tgmock.New(t)
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{})
	auth := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenSuccess{Authorization: auth})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := q.Auth(ctx, make(chan struct{}), func(ctx context.Context, token Token) error {
		t.Error("show must not be called")
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, auth, res)
}

type authResult struct {
	auth *tg.AuthAuthorization
	err  error
}

// On loginTokenSuccess from the refresh export, Auth must return that
// authorization without exporting a third time.
func TestQR_Auth_RefreshSuccessNoReExport(t *testing.T) {
	mock := tgmock.New(t)
	clk := neo.NewTime(time.Now())
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{Clock: clk})
	auth := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginToken{
		Expires: int(clk.Now().Add(time.Minute).Unix()), Token: []byte("t"),
	}).ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenSuccess{Authorization: auth})

	shown := make(chan struct{}, 4)
	done := make(chan authResult, 1)
	go func() {
		a, err := q.Auth(context.Background(), make(chan struct{}), func(ctx context.Context, token Token) error {
			shown <- struct{}{}
			return nil
		})
		done <- authResult{a, err}
	}()
	<-shown
	clk.Travel(time.Minute + time.Second)
	r := <-done
	require.NoError(t, r.err)
	require.Equal(t, auth, r.auth)
	// tgmock fails on an unexpected (third) exportLoginToken call.
}
