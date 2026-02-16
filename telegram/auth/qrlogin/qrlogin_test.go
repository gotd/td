package qrlogin

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"
	"rsc.io/qr"

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

func TestMigrationNeededError_Error(t *testing.T) {
	a := require.New(t)
	err := &MigrationNeededError{
		MigrateTo: &tg.AuthLoginTokenMigrateTo{
			DCID: 2,
		},
	}
	a.Equal("migration to 2 needed", err.Error())
}

// Mock dispatcher that implements the required interface.
type mockDispatcher struct {
	handler tg.LoginTokenHandler
}

func (m *mockDispatcher) OnLoginToken(h tg.LoginTokenHandler) {
	m.handler = h
}

func TestOnLoginToken(t *testing.T) {
	t.Skip("Flaky test")

	a := require.New(t)

	dispatcher := &mockDispatcher{}
	loggedIn := OnLoginToken(dispatcher)

	// Verify that handler was set.
	a.NotNil(dispatcher.handler)

	// Test the handler
	ctx := context.Background()
	entities := tg.Entities{}
	update := &tg.UpdateLoginToken{}

	// First call should send to channel.
	done := make(chan error, 1)
	go func() {
		done <- dispatcher.handler(ctx, entities, update)
	}()

	// Should receive signal.
	select {
	case <-loggedIn:
		// Good
	case <-time.After(time.Second * 5):
		t.Fatal("should receive signal")
	}

	// Handler should return nil.
	a.NoError(<-done)

	// Second call when channel is full should not block.
	err := dispatcher.handler(ctx, entities, update)
	a.NoError(err)
}

func TestToken_Image(t *testing.T) {
	a := require.New(t)
	token := NewToken([]byte("test_token"), int(time.Now().Unix()))

	// Test with valid QR level.
	img, err := token.Image(qr.L)
	a.NoError(err)
	a.NotNil(img)

	// Test with different QR levels.
	levels := []qr.Level{qr.L, qr.M, qr.Q, qr.H}

	for _, level := range levels {
		img, err := token.Image(level)
		a.NoError(err)
		a.NotNil(img)
	}
}

func TestQR_Import_WithMigration(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	// Test with migration function.
	migrateCalled := false
	migrate := func(ctx context.Context, dcID int) error {
		migrateCalled = true
		a.Equal(2, dcID)
		return nil
	}

	mock, qr := testQR(t, migrate)

	auth := &tg.AuthAuthorization{
		User: &tg.User{ID: 10},
	}

	// First call returns migration needed.
	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginTokenMigrateTo{
		DCID:  2,
		Token: testToken.token,
	}).ExpectCall(&tg.AuthImportLoginTokenRequest{
		Token: testToken.token,
	}).ThenResult(&tg.AuthLoginTokenSuccess{
		Authorization: auth,
	})

	result, err := qr.Import(ctx)
	a.NoError(err)
	a.Equal(auth, result)
	a.True(migrateCalled)
}

func TestQR_Import_MigrationError(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)

	// Test with migration function that returns error,
	migrate := func(ctx context.Context, dcID int) error {
		return testutil.TestError()
	}

	mock, qr := testQR(t, migrate)

	mock.ExpectCall(&tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}).ThenResult(&tg.AuthLoginTokenMigrateTo{
		DCID: 2,
	})

	_, err := qr.Import(ctx)
	a.ErrorIs(err, testutil.TestError())
}
