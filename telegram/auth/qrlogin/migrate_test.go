package qrlogin

import (
	"context"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

// The QR of an account on another DC is accepted and the refresh (not
// updateLoginToken) reports it with auth.loginTokenMigrateTo: Auth must
// migrate and import the token, as seen on production with a DC4 account.
func TestQR_Auth_MigrateOnRefresh(t *testing.T) {
	mock := tgmock.New(t)
	clk := neo.NewTime(time.Now())
	var migratedTo int
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{
		Clock: clk,
		Migrate: func(ctx context.Context, dcID int) error {
			migratedTo = dcID
			return nil
		},
	})
	auth := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginToken{
		Expires: int(clk.Now().Add(30 * time.Second).Unix()), Token: []byte("t"),
	}).
		ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenMigrateTo{DCID: 4, Token: []byte("m")}).
		ExpectCall(&tg.AuthImportLoginTokenRequest{Token: []byte("m")}).
		ThenResult(&tg.AuthLoginTokenSuccess{Authorization: auth})

	done := make(chan authResult, 1)
	go func() {
		a, err := q.Auth(context.Background(), make(chan struct{}), func(context.Context, Token) error {
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
			require.Equal(t, 4, migratedTo)
			return
		case <-deadline:
			t.Fatal("Auth did not finish")
		case <-time.After(5 * time.Millisecond):
			clk.Travel(time.Second)
		}
	}
}

// The token was accepted before Auth (e.g. on a restart): the first export
// already returns auth.loginTokenMigrateTo.
func TestQR_Auth_MigrateOnFirstExport(t *testing.T) {
	var migratedTo int
	mock, q := testQR(t, func(ctx context.Context, dcID int) error {
		migratedTo = dcID
		return nil
	})
	auth := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenMigrateTo{DCID: 4, Token: []byte("m")}).
		ExpectCall(&tg.AuthImportLoginTokenRequest{Token: []byte("m")}).
		ThenResult(&tg.AuthLoginTokenSuccess{Authorization: auth})

	res, err := q.Auth(context.Background(), make(chan struct{}), func(context.Context, Token) error {
		t.Fatal("token must not be shown")
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, auth, res)
	require.Equal(t, 4, migratedTo)
}

// 2FA is checked after the import on the migrated DC.
func TestQR_Auth_MigratePassword(t *testing.T) {
	mock := tgmock.New(t)
	q := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{
		Migrate: func(ctx context.Context, dcID int) error { return nil },
		Password: func(ctx context.Context) (string, error) {
			return "", errors.New("password requested")
		},
	})
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenMigrateTo{DCID: 4, Token: []byte("m")}).
		ExpectCall(&tg.AuthImportLoginTokenRequest{Token: []byte("m")}).
		ThenRPCErr(&tgerr.Error{Code: 401, Message: "SESSION_PASSWORD_NEEDED", Type: "SESSION_PASSWORD_NEEDED"})

	_, err := q.Auth(context.Background(), make(chan struct{}), func(context.Context, Token) error {
		return nil
	})
	require.ErrorContains(t, err, "password requested")
}

// Without a migrate function the migration is reported to the caller.
func TestQR_Auth_MigrateNotConfigured(t *testing.T) {
	mock, q := testQR(t, nil)
	mock.ExpectCall(exportReq()).ThenResult(&tg.AuthLoginTokenMigrateTo{DCID: 4, Token: []byte("m")})

	_, err := q.Auth(context.Background(), make(chan struct{}), func(context.Context, Token) error {
		return nil
	})
	var mig *MigrationNeededError
	require.ErrorAs(t, err, &mig)
	require.Equal(t, 4, mig.MigrateTo.DCID)
}
