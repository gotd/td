package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestPasswordHash(t *testing.T) {
	a := require.New(t)
	_, err := PasswordHash(nil, 0, nil, nil, nil)
	a.Error(err, "unsupported algo")
}

var testAlgo = &tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{
	Salt1: []uint8{
		230, 200, 149, 125, 223, 152, 141, 72,
	},
	Salt2: []uint8{
		159, 99, 68, 130, 43, 9, 108, 255, 135, 239, 164, 38, 245, 120, 87, 182,
	},
	G: 3,
	P: []uint8{
		199, 28, 174, 185, 198, 177, 201, 4, 142, 108, 82, 47, 112, 241, 63, 115,
		152, 13, 64, 35, 142, 62, 33, 193, 73, 52, 208, 55, 86, 61, 147, 15,
		72, 25, 138, 10, 167, 193, 64, 88, 34, 148, 147, 210, 37, 48, 244, 219,
		250, 51, 111, 110, 10, 201, 37, 19, 149, 67, 174, 212, 76, 206, 124, 55,
		32, 253, 81, 246, 148, 88, 112, 90, 198, 140, 212, 254, 107, 107, 19, 171,
		220, 151, 70, 81, 41, 105, 50, 132, 84, 241, 143, 175, 140, 89, 95, 100,
		36, 119, 254, 150, 187, 42, 148, 29, 91, 205, 29, 74, 200, 204, 73, 136,
		7, 8, 250, 155, 55, 142, 60, 79, 58, 144, 96, 190, 230, 124, 249, 164,
		164, 166, 149, 129, 16, 81, 144, 126, 22, 39, 83, 181, 107, 15, 107, 65,
		13, 186, 116, 216, 168, 75, 42, 20, 179, 20, 78, 14, 241, 40, 71, 84,
		253, 23, 237, 149, 13, 89, 101, 180, 185, 221, 70, 88, 45, 177, 23, 141,
		22, 156, 107, 196, 101, 176, 214, 255, 156, 163, 146, 143, 239, 91, 154, 228,
		228, 24, 252, 21, 232, 62, 190, 160, 248, 127, 169, 255, 94, 237, 112, 5,
		13, 237, 40, 73, 244, 123, 249, 89, 217, 86, 133, 12, 233, 41, 133, 31,
		13, 129, 21, 246, 53, 177, 5, 238, 46, 78, 21, 208, 75, 36, 84, 191,
		111, 79, 173, 240, 52, 177, 4, 3, 17, 156, 216, 227, 185, 47, 204, 91,
	},
}

func TestClient_UpdatePassword(t *testing.T) {
	ctx := context.Background()
	expectCall := func(a *require.Assertions, m *tgmock.Mock, hasPassword bool) *tgmock.RequestBuilder {
		p := &tg.AccountPassword{
			HasPassword:   hasPassword,
			NewAlgo:       testAlgo,
			NewSecureAlgo: &tg.SecurePasswordKdfAlgoUnknown{},
		}
		if hasPassword {
			p.CurrentAlgo = testAlgo
		}
		p.SetFlags()
		return m.ExpectCall(&tg.AccountGetPasswordRequest{}).
			ThenResult(p).ExpectFunc(func(b bin.Encoder) {
			a.IsType(&tg.AccountUpdatePasswordSettingsRequest{}, b)
			r := b.(*tg.AccountUpdatePasswordSettingsRequest)

			if !hasPassword {
				a.Equal(emptyPassword, r.Password)
			} else {
				a.NotEqual(emptyPassword, r.Password)
			}
			a.NotEmpty(r.NewSettings.NewPasswordHash)
			a.Equal("hint", r.NewSettings.Hint)
		})
	}

	t.Run("PasswordNotRequired", mockTest(func(
		a *require.Assertions,
		m *tgmock.Mock,
		client *Client,
	) {
		m.ExpectCall(&tg.AccountGetPasswordRequest{}).ThenErr(testutil.TestError())
		a.Error(client.UpdatePassword(ctx, "", UpdatePasswordOptions{}))

		expectCall(a, m, false).ThenTrue()
		a.NoError(client.UpdatePassword(ctx, "", UpdatePasswordOptions{
			Hint: "hint",
		}))
	}))

	t.Run("PasswordRequired", mockTest(func(
		a *require.Assertions,
		m *tgmock.Mock,
		client *Client,
	) {
		m.ExpectCall(&tg.AccountGetPasswordRequest{}).
			ThenResult(&tg.AccountPassword{
				HasPassword:   true,
				NewAlgo:       testAlgo,
				CurrentAlgo:   testAlgo,
				NewSecureAlgo: &tg.SecurePasswordKdfAlgoUnknown{},
			})
		a.ErrorIs(client.UpdatePassword(ctx, "", UpdatePasswordOptions{}), ErrPasswordNotProvided)

		m.ExpectCall(&tg.AccountGetPasswordRequest{}).
			ThenResult(&tg.AccountPassword{
				HasPassword:   true,
				NewAlgo:       testAlgo,
				CurrentAlgo:   testAlgo,
				NewSecureAlgo: &tg.SecurePasswordKdfAlgoUnknown{},
			})
		a.ErrorIs(client.UpdatePassword(ctx, "", UpdatePasswordOptions{
			Hint: "hint",
			Password: func(ctx context.Context) (string, error) {
				return "", testutil.TestError()
			},
		}), testutil.TestError())

		expectCall(a, m, true).ThenTrue()
		a.NoError(client.UpdatePassword(ctx, "", UpdatePasswordOptions{
			Hint: "hint",
			Password: func(ctx context.Context) (string, error) {
				return "password", nil
			},
		}))
	}))
}

func TestClient_ResetPassword(t *testing.T) {
	ctx := context.Background()
	wait := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	mockTest(func(a *require.Assertions, mock *tgmock.Mock, client *Client) {
		mock.ExpectCall(&tg.AccountResetPasswordRequest{}).ThenErr(testutil.TestError())
		_, err := client.ResetPassword(ctx)
		a.Error(err)

		mock.ExpectCall(&tg.AccountResetPasswordRequest{}).ThenResult(&tg.AccountResetPasswordFailedWait{
			RetryDate: int(wait),
		})
		var waitErr *ResetFailedWaitError
		_, err = client.ResetPassword(ctx)
		a.ErrorAs(err, &waitErr)
		a.Equal(int(wait), waitErr.Result.RetryDate)
		a.NotEmpty(waitErr.Error())

		mock.ExpectCall(&tg.AccountResetPasswordRequest{}).ThenResult(&tg.AccountResetPasswordOk{})
		r, err := client.ResetPassword(ctx)
		a.NoError(err)
		a.True(r.IsZero())

		mock.ExpectCall(&tg.AccountResetPasswordRequest{}).ThenResult(&tg.AccountResetPasswordRequestedWait{
			UntilDate: int(wait),
		})
		r, err = client.ResetPassword(ctx)
		a.NoError(err)
		a.False(r.IsZero())
	})(t)
}

func TestClient_CancelPasswordReset(t *testing.T) {
	ctx := context.Background()
	mockTest(func(a *require.Assertions, mock *tgmock.Mock, client *Client) {
		mock.ExpectCall(&tg.AccountDeclinePasswordResetRequest{}).ThenErr(testutil.TestError())
		a.Error(client.CancelPasswordReset(ctx))

		mock.ExpectCall(&tg.AccountDeclinePasswordResetRequest{}).ThenTrue()
		a.NoError(client.CancelPasswordReset(ctx))
	})(t)
}
