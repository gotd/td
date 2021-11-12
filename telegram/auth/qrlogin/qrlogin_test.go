package qrlogin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/internal/testutil"
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
	result, err = qr.Export(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenErr(testutil.TestError())
	result, err = qr.Export(ctx)
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
	result, err = qr.Accept(ctx, testToken)
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
	result, err = qr.Import(ctx)
	var mig *MigrationNeededError
	a.ErrorAs(err, &mig)
	a.Equal(1, mig.MigrateTo.DCID)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginToken{
		Token: testToken.token,
	})
	result, err = qr.Import(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenErr(testutil.TestError())
	result, err = qr.Import(ctx)
	a.ErrorIs(err, testutil.TestError())
}
