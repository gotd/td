package qrlogin

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

func testPassword(t *testing.T) *tg.AccountPassword {
	t.Helper()
	hexDecode := func(s string) []byte {
		b, err := hex.DecodeString(s)
		require.NoError(t, err)
		return b
	}
	algo := &tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{
		Salt1: hexDecode("4D11FB6BEC38F9D2546BB0F61E4F1C99A1BC0DB8F0D5F35B1291B37B213123D7ED48F3C6794D495B"),
		Salt2: hexDecode("A1B181AAFE88188680AE32860D60BB01"),
		G:     3,
		P: hexDecode("C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F" +
			"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37" +
			"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64" +
			"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4" +
			"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754" +
			"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4" +
			"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F" +
			"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B"),
	}
	pwd := &tg.AccountPassword{
		HasPassword:   true,
		NewAlgo:       algo,
		NewSecureAlgo: &tg.SecurePasswordKdfAlgoPBKDF2HMACSHA512iter100000{},
		SRPID:         1,
	}
	pwd.SetCurrentAlgo(algo)
	return pwd
}

func TestQR_ImportPassword(t *testing.T) {
	ctx := context.Background()
	want := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	export := &tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}

	newQR := func(t *testing.T, password func(ctx context.Context) (string, error)) (*tgmock.Mock, QR) {
		mock := tgmock.NewRequire(t)
		return mock, NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{
			Password: password,
		})
	}

	t.Run("Password", func(t *testing.T) {
		var asked int
		mock, qr := newQR(t, func(ctx context.Context) (string, error) {
			asked++
			return "password", nil
		})
		mock.ExpectCall(export).ThenRPCErr(tgerr.New(401, "SESSION_PASSWORD_NEEDED"))
		mock.ExpectCall(&tg.AccountGetPasswordRequest{}).ThenResult(testPassword(t))
		mock.ExpectFunc(func(b bin.Encoder) {
			req, ok := b.(*tg.AuthCheckPasswordRequest)
			require.True(t, ok, "%T", b)
			srp, ok := req.Password.(*tg.InputCheckPasswordSRP)
			require.True(t, ok)
			require.Equal(t, int64(1), srp.SRPID)
			require.NotEmpty(t, srp.A)
			require.NotEmpty(t, srp.M1)
		}).ThenResult(want)

		got, err := qr.Import(ctx)
		require.NoError(t, err)
		require.Equal(t, want, got)
		require.Equal(t, 1, asked)
	})
	t.Run("PasswordError", func(t *testing.T) {
		mock, qr := newQR(t, func(ctx context.Context) (string, error) {
			return "", testutil.TestError()
		})
		mock.ExpectCall(export).ThenRPCErr(tgerr.New(401, "SESSION_PASSWORD_NEEDED"))
		_, err := qr.Import(ctx)
		require.ErrorIs(t, err, testutil.TestError())
	})
	t.Run("NoPassword", func(t *testing.T) {
		mock, qr := newQR(t, nil)
		mock.ExpectCall(export).ThenRPCErr(tgerr.New(401, "SESSION_PASSWORD_NEEDED"))
		_, err := qr.Import(ctx)
		require.True(t, tgerr.Is(err, "SESSION_PASSWORD_NEEDED"))
	})
}

func TestQR_AuthPasswordOnExpire(t *testing.T) {
	// Token accepted around expiration: Export in the timer branch gets
	// SESSION_PASSWORD_NEEDED instead of Import.
	mock := tgmock.NewRequire(t)
	clk := neo.NewTime(time.Now())
	want := &tg.AuthAuthorization{User: &tg.User{ID: 10}}
	export := &tg.AuthExportLoginTokenRequest{
		APIID:   constant.TestAppID,
		APIHash: constant.TestAppHash,
	}
	mock.ExpectCall(export).ThenResult(&tg.AuthLoginToken{
		Expires: int(clk.Now().Add(time.Minute).Unix()),
		Token:   []byte("token"),
	})
	mock.ExpectCall(export).ThenRPCErr(tgerr.New(401, "SESSION_PASSWORD_NEEDED"))
	mock.ExpectCall(&tg.AccountGetPasswordRequest{}).ThenResult(testPassword(t))
	mock.ExpectFunc(func(b bin.Encoder) {
		_, ok := b.(*tg.AuthCheckPasswordRequest)
		require.True(t, ok, "%T", b)
	}).ThenResult(want)

	qr := NewQR(tg.NewClient(mock), constant.TestAppID, constant.TestAppHash, Options{
		Clock:    clk,
		Password: func(context.Context) (string, error) { return "password", nil },
	})

	show := make(chan struct{}, 1)
	type result struct {
		a   *tg.AuthAuthorization
		err error
	}
	done := make(chan result, 1)
	planned := clk.Observe()
	go func() {
		a, err := qr.Auth(context.Background(), make(chan struct{}), func(ctx context.Context, token Token) error {
			show <- struct{}{}
			return nil
		})
		done <- result{a, err}
	}()
	<-show
	<-planned
	clk.Travel(time.Minute + time.Second)

	r := <-done
	require.NoError(t, r.err)
	require.Equal(t, want, r.a)
}
