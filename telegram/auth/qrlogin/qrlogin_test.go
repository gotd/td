package qrlogin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func testQR(t *testing.T, migrate func(ctx context.Context, dcID int) error) (*tgmock.Mock, QR) {
	mock := tgmock.New(t)
	return mock, NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{
		Migrate: migrate,
	})
}

var testToken = NewToken([]byte("token"), 0)

func TestQR_Export(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	mock, qr := testQR(t, nil)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:     constant.TestAppID,
		APIHash:   constant.TestAppHash,
		ExceptIDs: []int64{0},
	}).ThenResult(&tg.AuthLoginToken{
		Expires: 0,
		Token:   testToken.token,
	})
	result, err := qr.Export(ctx, 0)
	a.NoError(err)
	a.Equal(Token{
		token:   testToken.token,
		expires: time.Unix(0, 0),
	}, result)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginTokenMigrateTo{
		Token: testToken.token,
	})
	_, err = qr.Export(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenErr(testutil.TestError())
	_, err = qr.Export(ctx)
	a.ErrorIs(err, testutil.TestError())
}

func TestQR_Accept(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	mock, qr := testQR(t, nil)

	auth := &tg.Authorization{
		APIID: 1,
	}
	mock.ExpectCall(&tg.AuthAcceptLoginTokenRequest{
		Token: testToken.token,
	}).ThenResult(auth)
	result, err := qr.Accept(ctx, testToken)
	a.NoError(err)
	a.Equal(auth, result)

	mock.ExpectCall(&tg.AuthAcceptLoginTokenRequest{
		Token: testToken.token,
	}).ThenErr(testutil.TestError())
	_, err = qr.Accept(ctx, testToken)
	a.ErrorIs(err, testutil.TestError())
}

func TestQR_Import(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	mock, qr := testQR(t, nil)

	auth := &tg.AuthAuthorization{
		User: &tg.User{ID: 10},
	}
	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginTokenSuccess{
		Authorization: auth,
	})
	result, err := qr.Import(ctx)
	a.NoError(err)
	a.Equal(auth, result)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginTokenMigrateTo{
		DCID: 1,
	})
	_, err = qr.Import(ctx)
	var mig *MigrationNeededError
	a.ErrorAs(err, &mig)
	a.Equal(1, mig.MigrateTo.DCID)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginToken{
		Token: testToken.token,
	})
	_, err = qr.Import(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenErr(testutil.TestError())
	_, err = qr.Import(ctx)
	a.ErrorIs(err, testutil.TestError())
}

func TestQR_Auth(t *testing.T) {
	a := require.New(t)
	mock := tgmock.New(t)
	clock := neo.NewTime(time.Now())

	auth := &tg.AuthAuthorization{
		User: &tg.User{ID: 10},
	}
	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginToken{
		Expires: int(clock.Now().Add(time.Minute).Unix()),
		Token:   testToken.token,
	}).ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginToken{
		Expires: int(clock.Now().Add(2 * time.Minute).Unix()),
		Token:   testToken.token,
	}).ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginTokenSuccess{
		Authorization: auth,
	})

	qr := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{
		Clock: clock,
	})

	show := make(chan struct{})
	done := make(chan error)
	loggedIn := make(chan struct{})
	go func() {
		_, err := qr.Auth(context.Background(), loggedIn, func(ctx context.Context, token Token) error {
			show <- struct{}{}
			return nil
		})
		done <- err
	}()

	// Show QR first time.
	<-show

	// Skip 1 minute, token expires.
	clock.Travel(time.Minute + 1)

	// Show QR second time.
	<-show

	// Emulate update, auth done.
	loggedIn <- struct{}{}

	a.NoError(<-done)
}
