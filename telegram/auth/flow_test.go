package auth_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
)

func askCode(code string, err error) auth.CodeAuthenticatorFunc {
	return func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		return code, err
	}
}

func TestConstantAuth(t *testing.T) {
	a := require.New(t)
	authConst := auth.Constant("phone", "password", askCode("123", nil))
	ctx := context.Background()

	result, err := authConst.Code(ctx, nil)
	a.NoError(err)
	a.Equal("123", result)

	result, err = authConst.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	result, err = authConst.Password(ctx)
	a.NoError(err)
	a.Equal("password", result)
}

func TestCodeOnlyAuth(t *testing.T) {
	a := require.New(t)
	authCodeOnly := auth.CodeOnly("phone", askCode("123", nil))
	ctx := context.Background()

	result, err := authCodeOnly.Code(ctx, nil)
	a.NoError(err)
	a.Equal("123", result)

	result, err = authCodeOnly.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	_, err = authCodeOnly.Password(ctx)
	a.ErrorIs(err, auth.ErrPasswordNotProvided)
}

func TestEnvAuth(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	prefix := "TEST_ENV_AUTH_"
	authEnv := auth.Env(prefix, askCode("123", nil))

	result, err := authEnv.Code(ctx, nil)
	a.NoError(err)
	a.Equal("123", result)

	_, err = authEnv.Phone(ctx)
	a.Error(err)

	_, err = authEnv.Password(ctx)
	a.ErrorIs(err, auth.ErrPasswordNotProvided)

	// Set envs.
	testutil.SetEnv(t, prefix+"PHONE", "phone")
	testutil.SetEnv(t, prefix+"PASSWORD", "password")

	result, err = authEnv.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	result, err = authEnv.Password(ctx)
	a.NoError(err)
	a.Equal("password", result)
}

func TestTestAuth(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	testAuth := auth.Test(testutil.ZeroRand{}, 2)

	_, err := testAuth.Code(ctx, &tg.AuthSentCode{
		Type: &tg.AuthSentCodeTypeFlashCall{},
	})
	a.Error(err)

	result, err := testAuth.Code(ctx, nil)
	a.NoError(err)
	a.Equal("22222", result)

	result, err = testAuth.Code(ctx, &tg.AuthSentCode{
		Type: &tg.AuthSentCodeTypeApp{
			Length: 1,
		},
	})
	a.NoError(err)
	a.Equal("2", result)

	result, err = testAuth.Phone(ctx)
	a.NoError(err)
	a.True(strings.HasPrefix(result, "999662"))

	_, err = testAuth.Password(ctx)
	a.ErrorIs(err, auth.ErrPasswordNotProvided)
}
